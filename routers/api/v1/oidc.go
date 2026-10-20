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

func redirectWithBindResult(c *gin.Context, success bool, msg string) {
	prefix := setting.ServerSetting.UrlPrefix
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = strings.TrimRight(prefix, "/")

	param := "sso_bind=success&msg=" + url.QueryEscape(msg)
	if !success {
		param = "sso_bind=failed&error=" + url.QueryEscape(msg)
	}

	targetURL := fmt.Sprintf("%s/#/settings?tab=password&%s", prefix, param)
	c.Redirect(http.StatusFound, targetURL)
}

func startSSOLogin(c *gin.Context, cfg *oidc_service.Config) {
	if setting.IsDemoMode() {
		redirectWithError(c, "/", "演示模式下已禁用第三方单点登录功能")
		return
	}
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

	isBindMode := strings.HasPrefix(postLogin, "bind:")

	if got := c.Query("state"); got == "" || got != baseState {
		if isBindMode {
			redirectWithBindResult(c, false, "单点登录 state 校验失败，请重试")
		} else {
			redirectWithError(c, postLogin, "登录 state 校验失败，请重试")
		}
		return
	}
	if e := c.Query("error"); e != "" {
		if isBindMode {
			redirectWithBindResult(c, false, "授权失败: "+e)
		} else {
			redirectWithError(c, postLogin, "授权失败: "+e)
		}
		return
	}
	code := c.Query("code")
	if code == "" {
		if isBindMode {
			redirectWithBindResult(c, false, "缺少授权码 (code)")
		} else {
			redirectWithError(c, postLogin, "缺少授权码 (code)")
		}
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
		if isBindMode {
			redirectWithBindResult(c, false, "登录通道未开启或配置不完整")
		} else {
			redirectWithError(c, postLogin, "登录通道未开启或配置不完整")
		}
		return
	}

	// 1. 若为已登录状态下的绑定流程
	if isBindMode {
		bindUserID := strings.TrimPrefix(postLogin, "bind:")
		claims, err := oidc_service.ExchangeClaims(c.Request.Context(), cfg, code, verifier)
		if err != nil {
			redirectWithBindResult(c, false, err.Error())
			return
		}

		_, err = oidc_service.BindUserWithClaims(cfg, bindUserID, claims)
		if err != nil {
			redirectWithBindResult(c, false, err.Error())
			return
		}

		redirectWithBindResult(c, true, "已成功绑定第三方单点登录账号！")
		return
	}

	// 2. 正常单点登录流程 (要求必须拥有内部账号方可 SSO)
	claims, err := oidc_service.ExchangeClaims(c.Request.Context(), cfg, code, verifier)
	if err != nil {
		redirectWithError(c, postLogin, err.Error())
		return
	}

	// 检查当前第三方身份是否已绑定内部账号
	user, err := models.GetUserBySSO(cfg.Type, claims.Sub)
	if err == nil && user != nil && user.ID != "" {
		// 已绑定内部账号：签发系统登录凭证
		expDays := settings_service.GetCookieExpDays()
		token, err := util.GenerateToken(user.ID, user.Username, user.Password, user.Role, expDays)
		if err != nil {
			redirectWithError(c, postLogin, "生成系统登录凭据失败")
			return
		}

		_ = models.AddLoginLog(user.ID, user.Username, c.ClientIP(), c.GetHeader("User-Agent"))

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
		return
	}

	// 未绑定内部账号：生成防篡改临时票据，引导用户绑定已有账号或创建内部账号并绑定
	ticket, err := util.GenerateSSOTicket(cfg.Type, claims.Sub, claims.Email, claims.PreferredUsername)
	if err != nil {
		redirectWithError(c, postLogin, "生成单点登录临时凭证失败: "+err.Error())
		return
	}

	prefix := setting.ServerSetting.UrlPrefix
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = strings.TrimRight(prefix, "/")

	targetURL := fmt.Sprintf("%s/#/login?sso_ticket=%s&sso_provider=%s&sso_username=%s&sso_email=%s",
		prefix,
		url.QueryEscape(ticket),
		url.QueryEscape(cfg.Type),
		url.QueryEscape(claims.PreferredUsername),
		url.QueryEscape(claims.Email),
	)
	c.Redirect(http.StatusFound, targetURL)
}

// SSOBindLoginReq 关联已有内部账号请求
type SSOBindLoginReq struct {
	Ticket   string `json:"ticket" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// SSOBindLogin 校验内部已有账号密码并立即绑定第三方身份源后登录
func SSOBindLogin(c *gin.Context) {
	appG := app.Gin{C: c}
	if setting.IsDemoMode() {
		appG.CResponse(http.StatusForbidden, "演示模式下已禁用第三方单点登录绑定功能", nil)
		return
	}

	var req SSOBindLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "参数格式错误: "+err.Error(), nil)
		return
	}

	claims, err := util.ParseSSOTicket(req.Ticket)
	if err != nil {
		appG.CResponse(http.StatusUnauthorized, "单点登录凭证无效或已过期，请重新发起登录", nil)
		return
	}

	// 校验内部账号密码
	ok, err := models.CheckAuth(req.Username, req.Password)
	if err != nil || !ok {
		appG.CResponse(http.StatusBadRequest, "账号或密码错误", nil)
		return
	}

	user, err := models.GetUserByUsername(req.Username)
	if err != nil || user == nil || user.ID == "" {
		appG.CResponse(http.StatusBadRequest, "账号不存在", nil)
		return
	}

	// 绑定 SSO 身份
	if err := models.BindUserSSO(user.ID, claims.Provider, claims.Subject, claims.Email); err != nil {
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 生成登录 Token
	expDays := settings_service.GetCookieExpDays()
	token, err := util.GenerateToken(user.ID, user.Username, user.Password, user.Role, expDays)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "生成系统凭据失败", nil)
		return
	}

	_ = models.AddLoginLog(user.ID, user.Username, c.ClientIP(), c.GetHeader("User-Agent"))

	appG.CResponse(http.StatusOK, "绑定成功并已登录", gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// SSORegisterBindReq 创建新内部账号并立即绑定
type SSORegisterBindReq struct {
	Ticket   string `json:"ticket" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
}

// SSORegisterBind 创建新内部账号并立即与第三方身份源绑定后登录
func SSORegisterBind(c *gin.Context) {
	appG := app.Gin{C: c}
	if setting.IsDemoMode() {
		appG.CResponse(http.StatusForbidden, "演示模式下已禁用第三方单点登录绑定与注册功能", nil)
		return
	}

	var req SSORegisterBindReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "参数格式错误: "+err.Error(), nil)
		return
	}

	claims, err := util.ParseSSOTicket(req.Ticket)
	if err != nil {
		appG.CResponse(http.StatusUnauthorized, "单点登录凭证无效或已过期，请重新发起登录", nil)
		return
	}

	// 检查系统是否开放自主注册
	if !settings_service.IsRegistrationAllowed() {
		appG.CResponse(http.StatusForbidden, "系统当前暂未开放自主注册，请联系管理员分配内部账号后选择【绑定已有账号】", nil)
		return
	}

	username := strings.TrimSpace(req.Username)
	// 校验用户名唯一性与格式
	if available, reason := models.CheckUsernameAvailable(username); !available {
		appG.CResponse(http.StatusBadRequest, reason, nil)
		return
	}
	if len(req.Password) < 4 || len(req.Password) > 50 {
		appG.CResponse(http.StatusBadRequest, "密码长度须在 4 到 50 个字符之间", nil)
		return
	}

	// 创建内部账号
	newUser, err := models.AddUser(username, req.Password, "user", "sso_register")
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "创建内部账号失败: "+err.Error(), nil)
		return
	}

	userEmail := strings.TrimSpace(req.Email)
	if userEmail == "" {
		userEmail = claims.Email
	}
	if userEmail != "" {
		_ = models.EditUserByID(newUser.ID, map[string]interface{}{"email": userEmail})
		newUser.Email = userEmail
	}

	// 绑定 SSO 身份
	if err := models.BindUserSSO(newUser.ID, claims.Provider, claims.Subject, claims.Email); err != nil {
		appG.CResponse(http.StatusInternalServerError, "绑定第三方身份失败: "+err.Error(), nil)
		return
	}

	// 生成登录 Token
	expDays := settings_service.GetCookieExpDays()
	token, err := util.GenerateToken(newUser.ID, newUser.Username, newUser.Password, newUser.Role, expDays)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "生成系统凭据失败", nil)
		return
	}

	_ = models.AddLoginLog(newUser.ID, newUser.Username, c.ClientIP(), c.GetHeader("User-Agent"))

	appG.CResponse(http.StatusOK, "注册并绑定成功", gin.H{
		"token": token,
		"user": gin.H{
			"id":       newUser.ID,
			"username": newUser.Username,
			"email":    newUser.Email,
			"role":     newUser.Role,
		},
	})
}

// StartSSOBind 当前已登录用户发起单点登录绑定（管理员及普通用户通用）
func StartSSOBind(c *gin.Context) {
	currentUserID := app.GetCurrentUserId(c)
	if currentUserID == "" {
		c.String(http.StatusUnauthorized, "请先登录系统再绑定第三方账号")
		return
	}

	channelType := c.Query("type")
	cfg := oidc_service.ResolveActiveConfig(channelType, c)
	if cfg == nil {
		c.String(http.StatusBadRequest, "该单点登录通道未开启或未配置完整")
		return
	}

	state := util.GenerateRandomString(32)
	verifier := oauth2.GenerateVerifier()
	postLogin := "bind:" + currentUserID

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

// SSOLogin 通用单点登录发起端点 (/api/v1/auth/sso/:provider/login)
func SSOLogin(c *gin.Context) {
	provider := c.Param("provider")
	cfg := oidc_service.ResolveActiveConfig(provider, c)
	if cfg == nil {
		c.String(http.StatusBadRequest, "未知的单点登录提供商或该通道未启用: "+provider)
		return
	}
	startSSOLogin(c, cfg)
}

// SSOCallback 通用单点登录授权回调端点 (/api/v1/auth/sso/:provider/callback)
func SSOCallback(c *gin.Context) {
	provider := c.Param("provider")
	cfg := oidc_service.ResolveActiveConfig(provider, c)
	handleSSOCallback(c, cfg)
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
