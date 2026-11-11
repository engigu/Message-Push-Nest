package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/pkg/util"
	"message-nest/service/cron_msg_service"
)

type DeleteCronMsgTaskReq struct {
	ID string `json:"id" validate:"required,len=12" label:"任务id"`
}

// DeleteCronMsgTask 删除定时消息
func DeleteCronMsgTask(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  DeleteCronMsgTaskReq
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

	msg, err := models.GetCronMsgByIDForOwner(req.ID, ownerID)
	if err != nil {
		appG.CResponse(http.StatusNotFound, "定时消息不存在或无权访问！", nil)
		return
	}

	CronMsgService := cron_msg_service.CronMsgService{
		ID:         req.ID,
		CreatedBy:  ownerID,
		ModifiedBy: user.GetAuditUserID(),
		IsAdmin:    user.IsAdmin() && !user.IsImpersonating(),
	}

	err = CronMsgService.Delete()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "删除定时消息失败！", nil)
		return
	}
	cron_msg_service.RemoveCronMsgToCronServer(msg)
	appG.CResponse(http.StatusOK, "删除定时消息成功！", nil)
}

// GetCronMsgList 获取定时消息列表
func GetCronMsgList(c *gin.Context) {
	appG := app.Gin{C: c}
	name := c.Query("name")

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	offset, limit := util.GetPageSize(c)
	CronMsgService := cron_msg_service.CronMsgService{
		Name:      name,
		PageNum:   offset,
		PageSize:  limit,
		CreatedBy: ownerID,
		IsAdmin:   user.IsAdmin() && !user.IsImpersonating(),
	}
	tasks, err := CronMsgService.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取定时消息失败！", nil)
		return
	}

	count, err := CronMsgService.Count()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取定时消息总数失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取定时消息成功", map[string]interface{}{
		"lists": tasks,
		"total": count,
	})
}

type AddCronMsgTaskReq struct {
	Name    string `json:"name" validate:"required,max=100,min=1" label:"消息名称"`
	TaskID  string `json:"task_id" validate:"required,max=100,min=1" label:"关联的任务id"`
	Cron    string `json:"cron" validate:"required,cron" label:"cron表达式"`
	Title   string `json:"title" validate:"required,max=100,min=1" label:"消息标题"`
	Content string `json:"content" validate:"required,max=10000,min=1" label:"消息内容"`
	Url     string `json:"url" validate:"" label:"消息详情地址"`
}

// AddCronMsgTask 添加定时消息
func AddCronMsgTask(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  AddCronMsgTaskReq
	)

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	// 校验关联任务是否归属当前租户
	if _, err := models.GetTaskByIDForOwner(req.TaskID, ownerID); err != nil {
		appG.CResponse(http.StatusNotFound, "关联的发信任务不存在或无权访问！", nil)
		return
	}

	targetOwner := user.GetEffectiveUserID()
	auditUserID := user.GetAuditUserID()

	CronMsgService := cron_msg_service.CronMsgService{
		Name:       req.Name,
		TaskID:     req.TaskID,
		Cron:       req.Cron,
		Title:      req.Title,
		Content:    req.Content,
		Url:        req.Url,
		CreatedBy:  targetOwner,
		ModifiedBy: auditUserID,
	}

	uuidstr, err := CronMsgService.Add()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "添加定时消息失败！", nil)
		return
	}
	CronMsgService.ID = uuidstr
	msg, _ := models.GetCronMsgByIDForOwner(uuidstr, ownerID)
	cron_msg_service.UpdateCronMsgToCronServer(msg)
	appG.CResponse(http.StatusOK, "添加定时消息成功！", nil)
}

type EditCronMsgTaskReq struct {
	ID      string `json:"id" validate:"required,len=12" label:"定时消息id"`
	Name    string `json:"name" validate:"required,max=100,min=1" label:"消息名称"`
	TaskID  string `json:"task_id" validate:"required,max=100,min=1" label:"关联的任务id"`
	Cron    string `json:"cron" validate:"required,cron" label:"cron表达式"`
	Title   string `json:"title" validate:"required,max=100,min=1" label:"消息标题"`
	Content string `json:"content" validate:"required,max=10000,min=1" label:"消息内容"`
	Url     string `json:"url" validate:"" label:"消息详情地址"`
	Enable  int    `json:"enable" validate:"oneof=0 1" label:"是否开启"`
}

// EditCronMsgTask 编辑定时消息任务
func EditCronMsgTask(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  EditCronMsgTaskReq
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

	// 校验定时消息自身所有权
	if _, err := models.GetCronMsgByIDForOwner(req.ID, ownerID); err != nil {
		appG.CResponse(http.StatusNotFound, "定时消息不存在或无权访问！", nil)
		return
	}

	// 校验修改后的关联任务所有权
	if _, err := models.GetTaskByIDForOwner(req.TaskID, ownerID); err != nil {
		appG.CResponse(http.StatusNotFound, "关联的发信任务不存在或无权访问！", nil)
		return
	}

	CronMsgService := cron_msg_service.CronMsgService{
		ID:         req.ID,
		CreatedBy:  ownerID,
		ModifiedBy: user.GetAuditUserID(),
		IsAdmin:    user.IsAdmin() && !user.IsImpersonating(),
	}

	data := make(map[string]interface{})
	data["name"] = req.Name
	data["task_id"] = req.TaskID
	data["cron"] = req.Cron
	data["title"] = req.Title
	data["content"] = req.Content
	data["url"] = req.Url
	data["enable"] = req.Enable
	data["modified_by"] = user.GetAuditUserID()
	err := CronMsgService.Edit(data)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "编辑定时消息失败！", nil)
		return
	}
	msg, _ := models.GetCronMsgByIDForOwner(req.ID, ownerID)
	cron_msg_service.UpdateCronMsgToCronServer(msg)
	appG.CResponse(http.StatusOK, "编辑定时消息成功！", nil)
}

type SendNowCronMsgReq struct {
	TaskID  string `json:"task_id" validate:"required,max=100,min=1" label:"关联的任务id"`
	Title   string `json:"title" validate:"required,max=100,min=1" label:"消息标题"`
	Content string `json:"content" validate:"required,max=10000,min=1" label:"消息内容"`
	Url     string `json:"url" validate:"" label:"消息详情地址"`
}

// SendNowCronMsg 立即发送定时消息
func SendNowCronMsg(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  SendNowCronMsgReq
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

	// 强校验关联任务归属，杜绝跨租户发信越权
	task, err := models.GetTaskByIDForOwner(req.TaskID, ownerID)
	if err != nil || task.ID == "" {
		appG.CResponse(http.StatusNotFound, "关联的发信任务不存在或无权访问！", nil)
		return
	}

	// 确认属于当前租户后，构造消息并发送
	CronMsgService := cron_msg_service.CronMsgService{
		TaskID:     req.TaskID,
		Title:      req.Title,
		Content:    req.Content,
		Url:        req.Url,
		CreatedBy:  task.CreatedBy,
		ModifiedBy: user.GetAuditUserID(),
	}

	err = CronMsgService.SendNowByParams(c.ClientIP())
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("发送失败：%s", err.Error()), nil)
		return
	}

	appG.CResponse(http.StatusOK, "发送成功！", nil)
}
