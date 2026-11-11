package v1

import (
	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
	"message-nest/pkg/util"
	"message-nest/service/send_logs_service"
	"net/http"
)

// GetTaskSendLogsList 获取发送日志列表
func GetTaskSendLogsList(c *gin.Context) {
	appG := app.Gin{C: c}
	name := c.Query("name")
	taskId := c.Query("taskid")
	query := c.Query("query")

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	offset, limit := util.GetPageSize(c)
	logsService := send_logs_service.SendTaskLogsService{
		TaskId:    taskId,
		Name:      name,
		PageNum:   offset,
		PageSize:  limit,
		Query:     query,
		CreatedBy: ownerID,
		IsAdmin:   user.IsAdmin() && !user.IsImpersonating(),
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
