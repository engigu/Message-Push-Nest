package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// GetCurrentUserId 从请求中获取当前用户ID
func GetCurrentUserId(c *gin.Context) string {
	userID, ok := c.Get("currentUserId")
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s", userID)
}

// GetCurrentUserName 从请求中获取当前用户名
func GetCurrentUserName(c *gin.Context) string {
	userName, ok := c.Get("currentUserName")
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s", userName)
}

// GetCurrentUserRole 从请求中获取当前用户角色
func GetCurrentUserRole(c *gin.Context) string {
	role, ok := c.Get("currentUserRole")
	if !ok {
		return "user"
	}
	return fmt.Sprintf("%s", role)
}

// IsAdmin 检查当前请求用户是否为管理员
func IsAdmin(c *gin.Context) bool {
	return GetCurrentUserRole(c) == "admin" || GetCurrentUserName(c) == "admin"
}
