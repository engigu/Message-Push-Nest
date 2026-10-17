package oidc_service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	coreos_oidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"message-nest/pkg/constant"
)

// ProviderCache 缓存发现文档与 JWKS，避免每次登录重复获取
type ProviderCache struct {
	mu       sync.Mutex
	issuer   string
	clientID string
	provider *coreos_oidc.Provider
	verifier *coreos_oidc.IDTokenVerifier
}

var globalCache = &ProviderCache{}

// NewSSOHTTPClient 统一的 SSO HTTP 客户端，支持私签证书环境与自签名根证书
func NewSSOHTTPClient() *http.Client {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}
}

// GetProviderFor 获取可用的 verifier 和 oauth2.Config（动态组装与多级缓存）
func GetProviderFor(ctx context.Context, cfg *Config) (*coreos_oidc.IDTokenVerifier, *oauth2.Config, error) {
	if cfg == nil {
		return nil, nil, fmt.Errorf("配置不能为空")
	}

	// 通用 OAuth 2.0 模式：直接构建标准 oauth2.Config
	if cfg.Type == constant.SSOTypeOAuth2 {
		oauthCfg := &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  cfg.AuthURL,
				TokenURL: cfg.TokenURL,
			},
			RedirectURL: cfg.RedirectURL,
			Scopes:      cfg.Scopes,
		}
		return nil, oauthCfg, nil
	}

	var provider *coreos_oidc.Provider
	var verifier *coreos_oidc.IDTokenVerifier

	globalCache.mu.Lock()
	if globalCache.provider != nil && globalCache.issuer == cfg.Issuer && globalCache.clientID == cfg.ClientID {
		provider = globalCache.provider
		verifier = globalCache.verifier
		globalCache.mu.Unlock()
	} else {
		globalCache.mu.Unlock()

		reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()

		httpClient := NewSSOHTTPClient()
		reqCtx = coreos_oidc.ClientContext(reqCtx, httpClient)
		reqCtx = context.WithValue(reqCtx, oauth2.HTTPClient, httpClient)

		issuerCtx := coreos_oidc.InsecureIssuerURLContext(reqCtx, cfg.Issuer)

		p, err := coreos_oidc.NewProvider(issuerCtx, cfg.Issuer)
		if err != nil {
			// 若末尾斜杠不一致（如发现文档末尾带斜杠），尝试自动切换斜杠
			altIssuer := strings.TrimRight(cfg.Issuer, "/")
			if altIssuer == cfg.Issuer {
				altIssuer = cfg.Issuer + "/"
			}
			altCtx := coreos_oidc.InsecureIssuerURLContext(reqCtx, altIssuer)
			if p2, err2 := coreos_oidc.NewProvider(altCtx, altIssuer); err2 == nil {
				p = p2
				err = nil
			}
		}
		if err != nil {
			return nil, nil, err
		}

		provider = p
		verifier = p.Verifier(&coreos_oidc.Config{ClientID: cfg.ClientID})

		globalCache.mu.Lock()
		globalCache.provider = provider
		globalCache.verifier = verifier
		globalCache.issuer = cfg.Issuer
		globalCache.clientID = cfg.ClientID
		globalCache.mu.Unlock()
	}

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
	}

	return verifier, oauthCfg, nil
}

// TestRequest 连通性测试请求
type TestRequest struct {
	Type        string `json:"type"`
	Issuer      string `json:"issuer"`
	AuthURL     string `json:"oauth2_auth_url"`
	TokenURL    string `json:"oauth2_token_url"`
	UserinfoURL string `json:"oauth2_userinfo_url"`
}

// TestResult 连通性测试返回结果
type TestResult struct {
	Message   string           `json:"message"`
	Endpoints *oauth2.Endpoint `json:"endpoints,omitempty"`
	Claims    map[string]any   `json:"claims,omitempty"`
}

// TestConnection 测试端点连通性（支持 OIDC 文档探测与通用 OAuth 2.0 端点探测）
func TestConnection(ctx context.Context, req TestRequest) (*TestResult, error) {
	httpClient := NewSSOHTTPClient()

	if req.Type == constant.SSOTypeOAuth2 || req.AuthURL != "" {
		endpoints := map[string]string{
			"授权端点":   req.AuthURL,
			"令牌端点":   req.TokenURL,
			"用户信息端点": req.UserinfoURL,
		}
		for name, url := range endpoints {
			u := strings.TrimSpace(url)
			if u == "" {
				return nil, fmt.Errorf("%s 不能为空", name)
			}
			reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			reqObj, err := http.NewRequestWithContext(reqCtx, http.MethodGet, u, nil)
			if err != nil {
				cancel()
				return nil, fmt.Errorf("%s 地址非法: %w", name, err)
			}
			reqObj.Header.Set("User-Agent", "Message-Nest-SSO/1.0")
			resp, err := httpClient.Do(reqObj)
			cancel()
			if err != nil {
				return nil, fmt.Errorf("%s 无法连接: %w", name, err)
			}
			resp.Body.Close()
		}
		return &TestResult{
			Message: "OAuth 2.0 端点连通性测试通过",
		}, nil
	}

	// OIDC 模式
	issuer := NormalizeIssuer(req.Issuer)
	if issuer == "" {
		return nil, fmt.Errorf("Issuer 服务地址不能为空")
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	reqCtx = coreos_oidc.ClientContext(reqCtx, httpClient)
	reqCtx = coreos_oidc.InsecureIssuerURLContext(reqCtx, issuer)

	provider, err := coreos_oidc.NewProvider(reqCtx, issuer)
	if err != nil {
		altIssuer := strings.TrimRight(issuer, "/")
		if altIssuer == issuer {
			altIssuer = issuer + "/"
		}
		altCtx := coreos_oidc.InsecureIssuerURLContext(reqCtx, altIssuer)
		if p2, err2 := coreos_oidc.NewProvider(altCtx, altIssuer); err2 == nil {
			provider = p2
			err = nil
		}
	}
	if err != nil {
		errDetail := err.Error()
		if strings.Contains(errDetail, "404") && strings.Contains(errDetail, "API endpoint not found") {
			return nil, fmt.Errorf("服务提供商探测失败 (404 Not Found)。请检查：1. Issuer 仅填写如 https://id.example.com，无需包含 /.well-known 等后缀；2. 若平台非标准 OIDC，可切换为通用 OAuth 2.0 模式手动配置端点。原始错误: %s", errDetail)
		}
		return nil, fmt.Errorf("服务提供商探测失败: %s", errDetail)
	}

	claims := map[string]any{}
	_ = provider.Claims(&claims)
	ep := provider.Endpoint()

	return &TestResult{
		Message:   "发现文档解析成功，OIDC 通信正常",
		Endpoints: &ep,
		Claims:    claims,
	}, nil
}
