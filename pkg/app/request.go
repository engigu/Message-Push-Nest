package app

import (
	"fmt"
	"github.com/astaxie/beego/validation"
	"github.com/gin-gonic/gin"

	"message-nest/pkg/logging"
)

// MarkErrors logs error logs
func MarkErrors(errors []*validation.Error) {
	for _, err := range errors {
		logging.Logger.Error(err.Key, err.Message)
	}
	return
}

// 从请求中获取当前用户
func GetCurrentUserName(c *gin.Context) string {
	userName, ok := c.Get("currentUserName")
	if !ok {
		return ""
	} else {
		return fmt.Sprintf("%s", userName)
	}
}
