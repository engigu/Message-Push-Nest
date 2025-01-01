package v1

import (
	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
	"message-nest/pkg/util"
	"message-nest/service/hosted_message_service"
	"net/http"
)

// GetHostMessageList 获取托管消息列表
func GetHostMessageList(c *gin.Context) {
	appG := app.Gin{C: c}
	text := c.Query("text")

	offset, limit := util.GetPageSize(c)
	messageService := hosted_message_service.HostMessageService{
		Text:     text,
		PageNum:  offset,
		PageSize: limit,
	}
	ways, err := messageService.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取托管消息失败！", nil)
		return
	}

	count, err := messageService.Count()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取托管消息总数失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取托管消息成功", map[string]interface{}{
		"lists": ways,
		"total": count,
	})
}
