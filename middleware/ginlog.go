package middleware

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 需要记录响应内容的 API 路径前缀
var logResponsePaths = []string{
	"/api/v1/message/send",
	"/api/v2/message/send",
}

// responseBodyWriter 用于捕获响应内容
type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

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

		// 判断是否需要捕获响应内容
		needCaptureResponse := false
		for _, logPath := range logResponsePaths {
			if strings.HasPrefix(path, logPath) {
				needCaptureResponse = true
				break
			}
		}
		
		var bodyWriter *responseBodyWriter
		if needCaptureResponse {
			bodyWriter = &responseBodyWriter{
				ResponseWriter: c.Writer,
				body:           bytes.NewBufferString(""),
			}
			c.Writer = bodyWriter
		}

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

		entry := logrus.WithFields(logrus.Fields{
			////"hostname":   hostname,
			//"statusCode": statusCode,
			//"latency":    latency,
			//"clientIP":   clientIP,
			//"method":     c.Request.Method,
			//"path":       path,
			////"referer":    referer,
			//"dataLength": dataLength,
			////"userAgent":  clientUserAgent,
			"prefix": "[Gin]",
		})

		if len(c.Errors) > 0 {
			entry.Error(c.Errors.ByType(gin.ErrorTypePrivate).String())
		} else {
			msg := fmt.Sprintf("%s [%s] %s %d %d (%dms)", clientIP, c.Request.Method, path, statusCode, dataLength, latency)
			
			// 如果是发送消息的 API，打印返回内容
			if needCaptureResponse && bodyWriter != nil {
				responseBody := bodyWriter.body.String()
				msg = fmt.Sprintf("%s | Response: %s", msg, responseBody)
			}
			
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
