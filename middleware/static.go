package middleware

import (
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

// StaticCacheMiddleware add embed file static cache
func StaticCacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		
		// 对静态资源应用缓存策略
		if strings.HasPrefix(path, "/assets/") || strings.Contains(path, "/assets/") {
			// 检查文件类型
			if isVersionedAsset(path) {
				// 带版本号的资源（如 index-Bw4BKttg.js）可以长期缓存
				// 使用 immutable 指令，浏览器在缓存期内不会重新验证
				c.Header("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				// 其他静态资源使用较短的缓存时间
				c.Header("Cache-Control", "public, max-age=86400")
			}
			
			// 添加 ETag 支持
			c.Header("ETag", `"`+path+`"`)
			
			// 添加 Expires 头（兼容旧浏览器）
			expires := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC1123)
			c.Header("Expires", expires)
			
			// 添加 Vary 头，告诉缓存服务器根据这些头部区分缓存
			c.Header("Vary", "Accept-Encoding")
		} else if path == "/" || strings.HasSuffix(path, ".html") {
			// HTML 文件不缓存或使用协商缓存
			c.Header("Cache-Control", "no-cache, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}
		
		// Continue to the next middleware or handler
		c.Next()
	}
}

// isVersionedAsset 检查是否是带版本号的资源文件
// Vite 构建的文件通常包含 hash，如：index-Bw4BKttg.js
func isVersionedAsset(path string) bool {
	// 检查是否包含 hash 模式（通常是 -[hash].js 或 -[hash].css）
	return strings.Contains(path, "-") && 
		(strings.HasSuffix(path, ".js") || 
		 strings.HasSuffix(path, ".css") ||
		 strings.HasSuffix(path, ".woff") ||
		 strings.HasSuffix(path, ".woff2") ||
		 strings.HasSuffix(path, ".ttf") ||
		 strings.HasSuffix(path, ".eot") ||
		 strings.HasSuffix(path, ".svg") ||
		 strings.HasSuffix(path, ".png") ||
		 strings.HasSuffix(path, ".jpg") ||
		 strings.HasSuffix(path, ".jpeg") ||
		 strings.HasSuffix(path, ".gif") ||
		 strings.HasSuffix(path, ".webp") ||
		 strings.HasSuffix(path, ".ico"))
}
