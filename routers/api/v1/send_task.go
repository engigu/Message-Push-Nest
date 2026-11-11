package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
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

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	if _, err := models.GetTaskByIDForOwner(req.ID, ownerID); err != nil {
		appG.CResponse(http.StatusNotFound, "发信任务不存在或无权访问", nil)
		return
	}

	MsgSendTaskService := send_task_service.SendTaskService{
		ID:         req.ID,
		CreatedBy:  ownerID,
		ModifiedBy: user.GetAuditUserID(),
		IsAdmin:    user.IsAdmin() && !user.IsImpersonating(),
	}
	err := MsgSendTaskService.Delete()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("删除发信任务失败！%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "删除发信任务成功！", nil)
}

// GetMsgSendTaskList 获取消息任务列表
func GetMsgSendTaskList(c *gin.Context) {
	appG := app.Gin{C: c}
	name := c.Query("name")

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	offset, limit := util.GetPageSize(c)
	MsgSendTaskService := send_task_service.SendTaskService{
		Name:      name,
		CreatedBy: ownerID,
		IsAdmin:   user.IsAdmin() && !user.IsImpersonating(),
		PageNum:   offset,
		PageSize:  limit,
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

	user, _, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	targetOwner := user.GetEffectiveUserID()
	MsgSendTaskService := send_task_service.SendTaskService{
		Name:       req.Name,
		CreatedBy:  targetOwner,
		ModifiedBy: user.GetAuditUserID(),
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

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	if _, err := models.GetTaskByIDForOwner(req.ID, ownerID); err != nil {
		appG.CResponse(http.StatusNotFound, "发信任务不存在或无权访问", nil)
		return
	}

	MsgSendTaskService := send_task_service.SendTaskService{
		ID:         req.ID,
		CreatedBy:  ownerID,
		ModifiedBy: user.GetAuditUserID(),
		IsAdmin:    user.IsAdmin() && !user.IsImpersonating(),
	}

	var data = map[string]interface{}{
		"name": req.TaskName,
	}

	err := MsgSendTaskService.Edit(data)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("编辑发信任务失败！%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "编辑发信任务成功！", nil)
}

// GetMsgSendTask 获取消息任务
func GetMsgSendTask(c *gin.Context) {
	appG := app.Gin{C: c}
	id := c.Query("id")

	if id == "" {
		appG.CResponse(http.StatusBadRequest, "任务id为空！", nil)
		return
	}

	_, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	sendTaskService := send_task_service.SendTaskService{
		ID:        id,
		CreatedBy: ownerID,
	}

	task, err := sendTaskService.GetByID()
	if err != nil {
		appG.CResponse(http.StatusNotFound, "任务不存在或无权访问！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取任务信息成功", task)
}
