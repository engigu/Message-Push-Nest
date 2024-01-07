package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"

	"github.com/gin-gonic/gin"

	"message-nest/pkg/e"
	"message-nest/pkg/util"
)

var ExcludedRoutes = []string{
	"/api/v1/message/send",
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
				c.Set("currentUserName", claims.Username)
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
