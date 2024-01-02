package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
)

// StaticCacheMiddleware add embed file static cache
func StaticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Apply the Cache-Control header to the static files
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			c.Header("Cache-Control", "private, max-age=86400")
		}
		// Continue to the next middleware or handler
		c.Next()
	}
}
