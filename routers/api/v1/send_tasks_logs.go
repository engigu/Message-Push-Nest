package v1

import (
	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
	"message-nest/pkg/util"
	"message-nest/service/send_logs_service"
	"net/http"
)

// GetMsgSendWayList 获取消息渠道列表
func GetTaskSendLogsList(c *gin.Context) {
	appG := app.Gin{C: c}
	name := c.Query("name")
	taskId := c.Query("taskid")

	offset, limit := util.GetPageSize(c)
	logsService := send_logs_service.SendTaskLogsService{
		TaskId:   taskId,
		Name:     name,
		PageNum:  offset,
		PageSize: limit,
	}
	ways, err := logsService.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取日志失败！", nil)
		return
	}

	count, err := logsService.Count()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取日志总数失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取日志成功", map[string]interface{}{
		"lists": ways,
		"total": count,
	})
}
