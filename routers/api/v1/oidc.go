package v1

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/constant"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"message-nest/service/oidc_service"
	"message-nest/service/settings_service"
)

const (
	oidcStateCookie    = "MNOIDCState"
	oidcVerifierCookie = "MNOIDCVerifier"
	oidcStateTTL       = 10 * time.Minute
)

// GetSSOStatus 获取单点登录公共状态与可用身份源
func GetSSOStatus(c *gin.Context) {
	appG := app.Gin{C: c}
	status := oidc_service.ResolveSSOStatus()
	appG.CResponse(http.StatusOK, "获取成功", status)
}

func redirectWithError(c *gin.Context, postLogin, errMsg string) {
	prefix := setting.ServerSetting.UrlPrefix
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = strings.TrimRight(prefix, "/")

	targetURL := fmt.Sprintf("%s/#/notice?type=warning&title=%s&error=%s&message=%s",
		prefix,
		url.QueryEscape("单点登录提示"),
		url.QueryEscape(errMsg),
		url.QueryEscape(errMsg),
	)
	c.Redirect(http.StatusFound, targetURL)
}

func startSSOLogin(c *gin.Context, cfg *oidc_service.Config) {
	if cfg == nil {
		c.String(http.StatusBadRequest, "该单点登录通道未开启或未配置完整")
		return
	}

	state := util.GenerateRandomString(32)
	postLogin := "/"
	if to := c.Query("redirect"); to != "" && !strings.HasPrefix(to, "//") {
		postLogin = to
	}

	verifier := oauth2.GenerateVerifier()

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oidcStateCookie, state+"."+cfg.Type+"."+postLogin, int(oidcStateTTL.Seconds()), "/", "", false, true)
	c.SetCookie(oidcVerifierCookie, verifier, int(oidcStateTTL.Seconds()), "/", "", false, true)

	authURL, err := oidc_service.BuildAuthURL(c.Request.Context(), cfg, state, verifier)
	if err != nil {
		c.String(http.StatusInternalServerError, "无法连接登录提供商: "+err.Error())
		return
	}

	c.Redirect(http.StatusFound, authURL)
}

func handleSSOCallback(c *gin.Context, defaultCfg *oidc_service.Config) {
	stateCookie, err := c.Cookie(oidcStateCookie)
	verifier, _ := c.Cookie(oidcVerifierCookie)

	// 清除 Cookie
	c.SetCookie(oidcStateCookie, "", -1, "/", "", false, true)
	c.SetCookie(oidcVerifierCookie, "", -1, "/", "", false, true)

	if err != nil || stateCookie == "" {
		redirectWithError(c, "/", "登录会话超时失效，请重新发起登录")
		return
	}

	parts := strings.SplitN(stateCookie, ".", 3)
	baseState := parts[0]
	channelType := ""
	postLogin := "/"
	if len(parts) == 3 {
		channelType = parts[1]
		postLogin = parts[2]
	} else if len(parts) == 2 {
		postLogin = parts[1]
	}

	if got := c.Query("state"); got == "" || got != baseState {
		redirectWithError(c, postLogin, "登录 state 校验失败，请重试")
		return
	}
	if e := c.Query("error"); e != "" {
		redirectWithError(c, postLogin, "授权失败: "+e)
		return
	}
	code := c.Query("code")
	if code == "" {
		redirectWithError(c, postLogin, "缺少授权码 (code)")
		return
	}

	cfg := defaultCfg
	if channelType != "" {
		if matched := oidc_service.ResolveActiveConfig(channelType, c); matched != nil {
			cfg = matched
		}
	}
	if cfg == nil {
		cfg = oidc_service.ResolveActiveConfig("", c)
	}
	if cfg == nil {
		redirectWithError(c, postLogin, "登录通道未开启或配置不完整")
		return
	}

	user, _, err := oidc_service.ExchangeAndResolve(c.Request.Context(), cfg, code, verifier)
	if err != nil {
		redirectWithError(c, postLogin, err.Error())
		return
	}

	expDays := settings_service.GetCookieExpDays()
	token, err := util.GenerateToken(user.Username, user.Password, expDays)
	if err != nil {
		redirectWithError(c, postLogin, "生成系统登录凭据失败")
		return
	}

	// 记录登录日志
	_ = models.AddLoginLog(user.ID, user.Username, c.ClientIP(), c.GetHeader("User-Agent"))

	// 拼接前端跳转 URL
	prefix := setting.ServerSetting.UrlPrefix
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = strings.TrimRight(prefix, "/")

	cleanPath := strings.TrimPrefix(postLogin, "#")
	if cleanPath == "" || cleanPath == "/login" {
		cleanPath = "/"
	}
	if !strings.HasPrefix(cleanPath, "/") {
		cleanPath = "/" + cleanPath
	}

	sep := "?"
	if strings.Contains(cleanPath, "?") {
		sep = "&"
	}
	targetURL := fmt.Sprintf("%s/#%s%stoken=%s", prefix, cleanPath, sep, url.QueryEscape(token))
	c.Redirect(http.StatusFound, targetURL)
}

// OidcLogin 发起标准 OIDC 登录
func OidcLogin(c *gin.Context) {
	startSSOLogin(c, oidc_service.ResolveActiveConfig(constant.SSOTypeOIDC, c))
}

// OidcCallback 标准 OIDC 授权回调
func OidcCallback(c *gin.Context) {
	handleSSOCallback(c, oidc_service.ResolveActiveConfig(constant.SSOTypeOIDC, c))
}

// OAuth2Login 发起通用 OAuth 2.0 登录
func OAuth2Login(c *gin.Context) {
	startSSOLogin(c, oidc_service.ResolveActiveConfig(constant.SSOTypeOAuth2, c))
}

// OAuth2Callback 通用 OAuth 2.0 授权回调
func OAuth2Callback(c *gin.Context) {
	handleSSOCallback(c, oidc_service.ResolveActiveConfig(constant.SSOTypeOAuth2, c))
}

// OidcTest 测试 OIDC / OAuth2 服务连通性
func OidcTest(c *gin.Context) {
	appG := app.Gin{C: c}
	var req oidc_service.TestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "请求参数格式错误: "+err.Error(), nil)
		return
	}

	result, err := oidc_service.TestConnection(c.Request.Context(), req)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}

	appG.CResponse(http.StatusOK, result.Message, result)
}

// OAuth2Test 测试 OAuth 2.0 平台端点连通性
func OAuth2Test(c *gin.Context) {
	OidcTest(c)
}
