package routers

import (
	"time"

	"github.com/gin-gonic/gin"
	"message-nest/middleware"
	"message-nest/routers/api"
	"message-nest/routers/api/v1"
)

// RegisterPublicRoutes 统一注册所有公开、免 JWT 鉴权的业务接口
// 严格划分领域层级：/api/v1/auth (认证域) 与 /api/v1/public (公开资源域)
func RegisterPublicRoutes(r *gin.RouterGroup) {
	// ==========================================
	// 1. 认证与凭据管理域 (/api/v1/auth)
	// 安全防线：挂载单 IP 限流，每分钟最多 30 次，防刷号注册与自动化攻击
	// ==========================================
	authGroup := r.Group("/auth")
	authGroup.Use(middleware.IPRateLimiter(30, time.Minute))
	{
		// 基础账号认证 (内部已含 IP + 账号双重密码试错熔断锁定)
		authGroup.POST("/login", api.GetAuth)
		authGroup.POST("/register", v1.Register)
		authGroup.GET("/check-username", v1.CheckUsernameAvailability)
		authGroup.GET("/captcha", v1.GetCaptcha)
		authGroup.POST("/captcha/verify", v1.VerifyCaptchaAndGetToken)

		// 单点登录子组 (/api/v1/auth/sso)
		ssoGroup := authGroup.Group("/sso")
		{
			ssoGroup.GET("/status", v1.GetSSOStatus)
			ssoGroup.GET("/:provider/login", v1.SSOLogin)
			ssoGroup.GET("/:provider/callback", v1.SSOCallback)
			ssoGroup.POST("/bind-login", v1.SSOBindLogin)
			ssoGroup.POST("/register-bind", v1.SSORegisterBind)
		}
	}

	// ==========================================
	// 2. 站点公开资源域 (/api/v1/public)
	// 安全防线：挂载单 IP 限流，每分钟最多 60 次，防爬虫高频扫盘
	// ==========================================
	publicGroup := r.Group("/public")
	publicGroup.Use(middleware.IPRateLimiter(60, time.Minute))
	{
		publicGroup.GET("/site/config", v1.GetPublicSiteConfig)
		publicGroup.GET("/messages/preview", v1.GetHostMessagePreview)
	}
}
