package v1

import (
	"message-nest/pkg/e"
	"net/http"

	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
	"message-nest/pkg/util"
	"message-nest/service/send_task_service"
)

type DeleteMsgSendTaskReq struct {
	ID string `json:"id" validate:"required,len=12" label:"任务id"`
}

// DeleteMsgSendTask 删除消息任务
func DeleteMsgSendTask(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  DeleteMsgSendTaskReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	MsgSendTaskService := send_task_service.SendTaskService{
		ID: req.ID,
	}
	err := MsgSendTaskService.Delete()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "删除发信任务失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "删除发信任务成功！", nil)
}

// GetMsgSendTaskList 获取消息任务列表
func GetMsgSendTaskList(c *gin.Context) {
	appG := app.Gin{C: c}
	name := c.Query("name")

	offset, limit := util.GetPageSize(c)
	MsgSendTaskService := send_task_service.SendTaskService{
		Name:     name,
		PageNum:  offset,
		PageSize: limit,
	}
	tasks, err := MsgSendTaskService.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取任务任务失败！", nil)
		return
	}

	count, err := MsgSendTaskService.Count()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取任务任务总数失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取任务任务成功", map[string]interface{}{
		"lists": tasks,
		"total": count,
	})
}

type AddMsgSendTaskReq struct {
	Name string `json:"name" validate:"required,max=100,min=1" label:"任务名"`
}

// AddMsgSendTask 添加发送任务
func AddMsgSendTask(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  AddMsgSendTaskReq
	)

	currentUser := app.GetCurrentUserName(c)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	MsgSendTaskService := send_task_service.SendTaskService{
		Name:       req.Name,
		CreatedBy:  currentUser,
		ModifiedBy: currentUser,
	}

	err := MsgSendTaskService.Add()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "添加任务任务失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "添加任务任务成功！", nil)

}

type EditMsgSendTaskReq struct {
	ID       string `json:"id" validate:"required,len=12" label:"任务id"`
	TaskName string `json:"name" validate:"required,max=100,min=1" label:"任务任务名"`
}

// EditMsgSendTask 编辑消息任务
func EditMsgSendTask(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  EditMsgSendTaskReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	MsgSendTaskService := send_task_service.SendTaskService{
		ID: req.ID,
	}

	var data = map[string]string{}
	data["name"] = req.TaskName
	err := MsgSendTaskService.Edit(data)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "编辑发信任务失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "编辑发信任务成功！", nil)
}
