package middleware

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"

	"github.com/gin-gonic/gin"

	"message-nest/pkg/app"
	"message-nest/pkg/e"
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
		path := c.FullPath()
		for _, route := range ExcludedRoutes {
			if path == route {
				c.Next()
				return
			}
		}

		var code int
		var data interface{}
		code = e.SUCCESS
		token := c.Request.Header.Get("M-Token")
		if token == "" {
			authHeader := c.Request.Header.Get("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			} else if authHeader != "" {
				token = authHeader
			}
		}

		if token == "" {
			code = e.ERROR_AUTH_NO_TOKEN
		} else {
			claims, err := util.ParseToken(token)
			if err != nil {
				switch {
				case errors.Is(err, jwt.ErrTokenMalformed):
					code = e.ERROR_AUTH
				case errors.Is(err, jwt.ErrTokenSignatureInvalid):
					code = e.ERROR_AUTH_CHECK_TOKEN_FAIL
				case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
					code = e.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
				default:
					code = e.ERROR_AUTH_CHECK_TOKEN_FAIL
				}
			} else {
				c.Set("currentUserId", claims.UserID)
				c.Set("currentUserName", claims.Username)
				c.Set("currentUserRole", claims.Role)

				// 管理员穿越 Context（Impersonation）：仅当具备管理员权限时才允许透传目标租户 ID
				var impersonateUserID string
				if claims.Role == "admin" || claims.Username == "admin" {
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
			}
		}

		if code != e.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg":  e.GetMsg(code),
				"data": data,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminRequired 检查用户是否具备管理员权限
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("currentUserRole")
		username, _ := c.Get("currentUserName")
		if role != "admin" && username != "admin" {
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
