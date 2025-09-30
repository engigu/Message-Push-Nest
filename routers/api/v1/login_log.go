package v1

import (
    "github.com/gin-gonic/gin"
    "message-nest/pkg/app"
    "message-nest/pkg/e"
    "message-nest/models"
)

// GetRecentLoginLogs 最近登录日志（默认8条）
func GetRecentLoginLogs(c *gin.Context) {
    appG := app.Gin{C: c}
    logs, err := models.GetRecentLoginLogs(8)
    if err != nil {
        appG.CResponse(e.ERROR, err.Error(), nil)
        return
    }
    appG.CResponse(e.SUCCESS, "success", gin.H{
        "lists": logs,
    })
}


