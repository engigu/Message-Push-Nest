package middleware

import (
	"fmt"
	"math"
	"message-nest/pkg/logging"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//var timeFormat = "02/Jan/2006:15:04:05 -0700"

// LogMiddleware  Logger is the logrus logger handler
func LogMiddleware(notLogged ...string) gin.HandlerFunc {
	//hostname, err := os.Hostname()
	//if err != nil {
	//	hostname = "unknown"
	//}

	var skip map[string]struct{}

	if length := len(notLogged); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, p := range notLogged {
			skip[p] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		start := time.Now()

		c.Next()

		stop := time.Since(start)
		latency := int(math.Ceil(float64(stop.Nanoseconds()) / 1000000.0))
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		//clientUserAgent := c.Request.UserAgent()
		//referer := c.Request.Referer()
		dataLength := c.Writer.Size()
		if dataLength < 0 {
			dataLength = 0
		}
		if raw != "" {
			raw, _ := url.QueryUnescape(raw)
			path = path + "?" + raw
		}
		if _, ok := skip[path]; ok {
			return
		}

		entry := logging.Logger.WithFields(logrus.Fields{
			//"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency, // time to process
			"clientIP":   clientIP,
			"method":     c.Request.Method,
			"path":       path,
			//"referer":    referer,
			"dataLength": dataLength,
			//"userAgent":  clientUserAgent,
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("[Gin] %s [%s] %s %d %d (%dms)", clientIP, c.Request.Method, path, statusCode, dataLength, latency)
			if statusCode >= http.StatusInternalServerError {
				entry.Error(msg)
			} else if statusCode >= http.StatusBadRequest {
				entry.Warn(msg)
			} else {
				entry.Info(msg)
			}
		}
	}
}
