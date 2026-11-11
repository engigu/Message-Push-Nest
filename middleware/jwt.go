package middleware

import (
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"

	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
)

var ExcludedRoutes = []string{
	"/api/v1/message/send",
	"/api/v2/message/send",
	"/api/v1/settings/getsetting",
}

// JWT is jwt middleware
func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		cleanPath := c.FullPath()
		if prefix := setting.ServerSetting.UrlPrefix; prefix != "" {
			cleanPath = strings.TrimPrefix(cleanPath, prefix)
		}
		isExcluded := slices.Contains(ExcludedRoutes, cleanPath)

		token := c.Request.Header.Get("M-Token")
		if token == "" {
			authHeader := c.Request.Header.Get("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			} else if authHeader != "" {
				token = authHeader
			}
		}
		// 浏览器直接跳转 (如 SSO 绑定/导出下载) 从 Query 参数或 Cookie 提取 token
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			if cookieToken, err := c.Cookie("token"); err == nil && cookieToken != "" {
				token = cookieToken
			}
		}

		if token == "" {
			if isExcluded {
				c.Next()
				return
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": e.ERROR_AUTH_NO_TOKEN,
				"msg":  e.GetMsg(e.ERROR_AUTH_NO_TOKEN),
				"data": nil,
			})
			c.Abort()
			return
		}

		claims, err := util.ParseToken(token)
		if err != nil {
			if isExcluded {
				// 对于可选免登录路由，若传入的并非有效用户 JWT（例如外部发信 token），不予阻断
				c.Next()
				return
			}
			code := e.ERROR_AUTH_CHECK_TOKEN_FAIL
			switch {
			case errors.Is(err, jwt.ErrTokenMalformed):
				code = e.ERROR_AUTH
			case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
				code = e.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg":  e.GetMsg(code),
				"data": nil,
			})
			c.Abort()
			return
		}

		c.Set("currentUserId", claims.UserID)
		c.Set("currentUserName", claims.Username)
		c.Set("currentUserRole", claims.Role)

		// 管理员穿越 Context（Impersonation）：仅当具备管理员权限时才允许透传目标租户 ID
		var impersonateUserID string
		if claims.IsAdmin() {
			impersonateUserID = c.Request.Header.Get("X-Tenant-User-Id")
			if impersonateUserID == "" {
				impersonateUserID = c.Request.Header.Get("X-Impersonate-User-Id")
			}
			if impersonateUserID != "" {
				c.Set("impersonateUserId", impersonateUserID)
			}
		}

		// 同步注入标准 context.Context，供 GORM 钩子与 Scope 自动读取
		ctx := app.WithUserContext(c.Request.Context(), app.UserContextInfo{
			UserID:            claims.UserID,
			Username:          claims.Username,
			Role:              claims.Role,
			ImpersonateUserID: impersonateUserID,
		})
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// AdminRequired 检查用户是否具备管理员权限
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !app.IsAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"code": http.StatusForbidden,
				"msg":  "无权访问：该操作仅限管理员执行！",
				"data": nil,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
