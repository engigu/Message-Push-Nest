package v1

import (
	"message-nest/pkg/e"
	"message-nest/pkg/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
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

	CronMsgService := cron_msg_service.CronMsgService{
		ID: req.ID,
	}
	msg, _ := CronMsgService.GetByID()

	err := CronMsgService.Delete()
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

	offset, limit := util.GetPageSize(c)
	CronMsgService := cron_msg_service.CronMsgService{
		Name:     name,
		PageNum:  offset,
		PageSize: limit,
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

	currentUser := app.GetCurrentUserName(c)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	CronMsgService := cron_msg_service.CronMsgService{
		Name:       req.Name,
		TaskID:     req.TaskID,
		Cron:       req.Cron,
		Title:      req.Title,
		Content:    req.Content,
		Url:        req.Url,
		CreatedBy:  currentUser,
		ModifiedBy: currentUser,
	}

	uuidstr, err := CronMsgService.Add()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "添加定时消息失败！", nil)
		return
	}
	CronMsgService.ID = uuidstr
	msg, _ := CronMsgService.GetByID()
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

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	CronMsgService := cron_msg_service.CronMsgService{
		ID: req.ID,
	}

	data := make(map[string]interface{})
	data["name"] = req.Name
	data["task_id"] = req.TaskID
	data["cron"] = req.Cron
	data["title"] = req.Title
	data["content"] = req.Content
	data["url"] = req.Url
	data["enable"] = req.Enable
	err := CronMsgService.Edit(data)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "编辑定时消息失败！", nil)
		return
	}
	msg, _ := CronMsgService.GetByID()
	cron_msg_service.UpdateCronMsgToCronServer(msg)
	appG.CResponse(http.StatusOK, "编辑定时消息成功！", nil)
}
