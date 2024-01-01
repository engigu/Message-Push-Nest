package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"message-nest/pkg/logging"
	"time"
)

const (
	status200 = 42
	status404 = 43
	status500 = 41

	methodGET = 44
)

func formatDuration(duration time.Duration) string {
	seconds := duration.Seconds()
	switch {
	case seconds >= 1:
		return fmt.Sprintf("%.2fs", seconds)
	default:
		return fmt.Sprintf("%.3fms", float64(duration.Milliseconds()))
		//default:
		//	return fmt.Sprintf("%dns", duration.Nanoseconds())
		//case duration.Milliseconds() >= 1:
		//	return fmt.Sprintf("%.2fms", float64(duration.Milliseconds()))
		//default:
		//	return fmt.Sprintf("%dns", duration.Nanoseconds())
	}
}

func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped

		// Stop timer
		end := time.Now()
		timeSub := end.Sub(start)
		dr := formatDuration(timeSub)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		//bodySize := c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}

		var statusColor string
		switch statusCode {
		case 200:
			statusColor = fmt.Sprintf("\033[%dm %d \033[0m", status200, statusCode)
		case 500:
			statusColor = fmt.Sprintf("\033[%dm %d \033[0m", status500, statusCode)
		default:
			statusColor = fmt.Sprintf("\033[%dm %d \033[0m", status404, statusCode)
		}

		var methodColor string
		methodColor = fmt.Sprintf("\033[%dm %s \033[0m", methodGET, method)

		logging.Logger.Infof("[GIN]|%s|%s|%s|%s| %s",
			//start.Format("2006-01-02 15:04:06"),
			statusColor,
			dr,
			clientIP,
			methodColor,
			path,
		)

	}
}
