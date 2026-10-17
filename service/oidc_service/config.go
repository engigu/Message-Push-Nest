package oidc_service

import (
	"fmt"
	"net/url"
	"strings"

	coreos_oidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"message-nest/models"
	"message-nest/pkg/constant"
	"message-nest/pkg/setting"
)

// NormalizeIssuer 规范化 Issuer 格式，自动剔除尾部斜杠以及误填的发现文档路径
func NormalizeIssuer(raw string) string {
	u := strings.TrimSpace(raw)
	u = strings.TrimRight(u, "/")
	for {
		trimmed := strings.TrimSuffix(u, "/.well-known/openid-configuration")
		trimmed = strings.TrimSuffix(trimmed, "/.well-known/openid-configuration/")
		trimmed = strings.TrimRight(trimmed, "/")
		if trimmed == u {
			break
		}
		u = trimmed
	}
	return u
}

// Config 统一单点登录配置模型
type Config struct {
	Type         string   `json:"type"` // "oidc" 或 "oauth2"
	Issuer       string   `json:"issuer"`
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RedirectURL  string   `json:"redirect_url"`
	Scopes       []string `json:"scopes"`
	AutoCreate   bool     `json:"auto_create"`

	// 通用 OAuth 2.0 专用字段
	AuthURL     string `json:"auth_url"`
	TokenURL    string `json:"token_url"`
	UserinfoURL string `json:"userinfo_url"`
	DisplayName string `json:"display_name"`
	UserKey     string `json:"user_key"`
}

// SecurityReader 封装配置读取
type SecurityReader struct{}

func NewSecurityReader() SecurityReader {
	return SecurityReader{}
}

func (r SecurityReader) Get(keys ...string) string {
	if models.GetDB() == nil {
		return ""
	}
	for _, k := range keys {
		s, err := models.GetSettingByKey(constant.SecuritySectionName, k)
		if err == nil && strings.TrimSpace(s.Value) != "" {
			return strings.TrimSpace(s.Value)
		}
	}
	return ""
}

func (r SecurityReader) GetWithDefault(defaultVal string, keys ...string) string {
	if v := r.Get(keys...); v != "" {
		return v
	}
	return defaultVal
}

func (r SecurityReader) Bool(keys ...string) bool {
	return r.Get(keys...) == "true"
}

func (r SecurityReader) Scopes(raw string, defaultScopes ...string) []string {
	seen := make(map[string]bool)
	var list []string
	for _, s := range defaultScopes {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			list = append(list, s)
		}
	}
	for _, s := range strings.FieldsFunc(raw, func(c rune) bool {
		return c == ',' || c == ' ' || c == '\n'
	}) {
		s = strings.TrimSpace(s)
		if s != "" && !seen[s] {
			seen[s] = true
			list = append(list, s)
		}
	}
	return list
}

// ResolveRedirectURL 计算最终回调地址
func ResolveRedirectURL(custom string, defaultPath string, c *gin.Context) string {
	if trimmed := strings.TrimSpace(custom); trimmed != "" {
		return trimmed
	}
	prefix := setting.ServerSetting.UrlPrefix
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = strings.TrimRight(prefix, "/")
	fullPath := prefix + defaultPath

	if c != nil && c.Request != nil {
		scheme := "http"
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := c.GetHeader("X-Forwarded-Host")
		if host == "" {
			if ref := c.GetHeader("Referer"); ref != "" {
				if u, err := url.Parse(ref); err == nil && u.Host != "" {
					host = u.Host
					if u.Scheme != "" {
						scheme = u.Scheme
					}
				}
			}
		}
		if host == "" {
			host = c.Request.Host
		}
		if host != "" {
			return fmt.Sprintf("%s://%s%s", scheme, host, fullPath)
		}
	}
	return fullPath
}

// ResolveSSOStatus 为前端登录页状态与展示信息（支持标准 OIDC 与通用 OAuth 2.0 并行）
func ResolveSSOStatus() gin.H {
	r := NewSecurityReader()
	var providers []gin.H

	prefix := setting.ServerSetting.UrlPrefix
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = strings.TrimRight(prefix, "/")

	// 1. 标准 OIDC 通道
	if r.Bool(constant.KeyOidcEnabled) {
		clientID := r.Get(constant.KeyOidcClientID)
		issuer := NormalizeIssuer(r.Get(constant.KeyOidcIssuer))
		if clientID != "" && issuer != "" {
			providers = append(providers, gin.H{
				"id":        "oidc",
				"type":      constant.SSOTypeOIDC,
				"name":      "OIDC",
				"login_url": prefix + constant.PathAuthOidcLogin,
			})
		}
	}

	// 2. 通用 OAuth 2.0 通道
	if r.Bool(constant.KeyOauth2Enabled) {
		clientID := r.Get(constant.KeyOauth2ClientID)
		authURL := r.Get(constant.KeyOauth2AuthURL)
		tokenURL := r.Get(constant.KeyOauth2TokenURL)
		userinfoURL := r.Get(constant.KeyOauth2UserinfoURL)
		if clientID != "" && authURL != "" && tokenURL != "" && userinfoURL != "" {
			providers = append(providers, gin.H{
				"id":        "oauth2",
				"type":      constant.SSOTypeOAuth2,
				"name":      r.GetWithDefault("OAuth 2.0", constant.KeyOauth2Name),
				"login_url": prefix + constant.PathAuthOAuth2Login,
			})
		}
	}

	if len(providers) == 0 {
		return gin.H{"enabled": false, "providers": []gin.H{}}
	}

	return gin.H{
		"enabled":   true,
		"providers": providers,
		"type":      providers[0]["type"],
		"name":      providers[0]["name"],
	}
}

// LoadOidcConfig 读取标准 OIDC 配置
func LoadOidcConfig(redirectURL string) *Config {
	r := NewSecurityReader()
	if !r.Bool(constant.KeyOidcEnabled) {
		return nil
	}

	clientID := r.Get(constant.KeyOidcClientID)
	issuer := NormalizeIssuer(r.Get(constant.KeyOidcIssuer))
	if clientID == "" || issuer == "" {
		return nil
	}

	return &Config{
		Type:         constant.SSOTypeOIDC,
		Issuer:       issuer,
		ClientID:     clientID,
		ClientSecret: r.Get(constant.KeyOidcClientSecret),
		RedirectURL:  redirectURL,
		Scopes:       r.Scopes(r.Get(constant.KeyOidcScopes), coreos_oidc.ScopeOpenID, "profile", "email"),
		AutoCreate:   r.Bool(constant.KeyOidcAutoCreate),
	}
}

// LoadOAuth2Config 读取通用 OAuth 2.0 配置
func LoadOAuth2Config(redirectURL string) *Config {
	r := NewSecurityReader()
	if !r.Bool(constant.KeyOauth2Enabled) {
		return nil
	}

	clientID := r.Get(constant.KeyOauth2ClientID)
	authURL := r.Get(constant.KeyOauth2AuthURL)
	tokenURL := r.Get(constant.KeyOauth2TokenURL)
	userinfoURL := r.Get(constant.KeyOauth2UserinfoURL)
	if clientID == "" || authURL == "" || tokenURL == "" || userinfoURL == "" {
		return nil
	}

	return &Config{
		Type:         constant.SSOTypeOAuth2,
		ClientID:     clientID,
		ClientSecret: r.Get(constant.KeyOauth2ClientSecret),
		RedirectURL:  redirectURL,
		Scopes:       r.Scopes(r.Get(constant.KeyOauth2Scopes)),
		AutoCreate:   r.Bool(constant.KeyOauth2AutoCreate),
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		UserinfoURL:  userinfoURL,
		DisplayName:  r.GetWithDefault("OAuth 2.0", constant.KeyOauth2Name),
		UserKey:      r.Get(constant.KeyOauth2UserKey),
	}
}

// ResolveActiveConfig 根据偏好（"oidc" 或 "oauth2"，缺省时自动探测）动态获取当前有效的登录配置
func ResolveActiveConfig(preferredType string, c *gin.Context) *Config {
	r := NewSecurityReader()
	oidcRedirectURL := ResolveRedirectURL(r.Get(constant.KeyOidcRedirectURL), constant.PathAuthOidcCallback, c)
	oauth2RedirectURL := ResolveRedirectURL("", constant.PathAuthOAuth2Callback, c)

	if preferredType == constant.SSOTypeOAuth2 {
		if cfg := LoadOAuth2Config(oauth2RedirectURL); cfg != nil {
			return cfg
		}
		return LoadOidcConfig(oidcRedirectURL)
	}

	// 默认 OIDC
	if cfg := LoadOidcConfig(oidcRedirectURL); cfg != nil {
		return cfg
	}
	return LoadOAuth2Config(oauth2RedirectURL)
}
