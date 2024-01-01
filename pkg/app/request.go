package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// 从请求中获取当前用户
func GetCurrentUserName(c *gin.Context) string {
	userName, ok := c.Get("currentUserName")
	if !ok {
		return ""
	} else {
		return fmt.Sprintf("%s", userName)
	}
}
