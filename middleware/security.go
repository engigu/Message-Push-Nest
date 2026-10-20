package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeaders 全局安全响应头中间件
// 为所有 HTTP 响应注入企业级安全防护头，防 MIME 嗅探、防点击劫持、防 XSS 反射与隐私泄漏
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止浏览器根据文件内容猜测 MIME 类型 (防 XSS)
		c.Header("X-Content-Type-Options", "nosniff")

		// 防止网页被外部不可信域名通过 iframe 嵌入 (防点击劫持 Clickjacking)
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// 启用浏览器内置的 XSS 过滤器，检测到攻击时直接阻止页面渲染
		c.Header("X-XSS-Protection", "1; mode=block")

		// 规范 Referrer 策略，避免跳转第三方时泄露敏感 URL 参数
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}
