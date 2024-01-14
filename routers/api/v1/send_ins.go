package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_task_service"
	"net/http"
)

type DeleteMsgTaskInsReq struct {
	ID string `json:"id" validate:"required,len=36" label:"实例id"`
}

// DeleteMsgSendWay 删除消息渠道
func DeleteMsgTaskIns(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  DeleteMsgTaskInsReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	InsService := send_ins_service.SendTaskInsService{
		ID: req.ID,
	}
	err := InsService.Delete()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "删除实例失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "删除实例成功！", nil)
}

// GetMsgSendWayIns 获取消息任务实例
func GetMsgSendWayIns(c *gin.Context) {
	appG := app.Gin{C: c}
	id := c.Query("id")

	if id == "" {
		appG.CResponse(http.StatusBadRequest, "任务id为空！", nil)
		return
	}

	sendTaskService := send_task_service.SendTaskService{
		ID: id,
	}

	task, err := sendTaskService.GetTaskWithIns()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "获取到的任务信息为空！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取任务信息成功", task)
}

type SendTasksInsReq struct {
	ID          string `json:"id" validate:"required,len=36" label:"实例id"`
	TaskId      string `json:"task_id" validate:"required,len=36" label:"ins任务id"`
	WayID       string `json:"way_id" validate:"required,len=36" label:"渠道id"`
	ContentType string `json:"content_type" validate:"required,max=100" label:"实例内容类型"`
	Config      string `json:"config" validate:"" label:"任务配置"`
	Extra       string `json:"extra" validate:"" label:"任务额外信息"`
	//WayName     string `json:"way_name" validate:"required,max=100" label:"渠道名"`
	WayType string `json:"way_type" validate:"required,max=100" label:"渠道类型"`
}

type AddManyTasksInsReq struct {
	TaskId   string            `json:"id" validate:"required,len=36" label:"任务id"`
	TaskName string            `json:"name" validate:"required,max=100" label:"任务名"`
	InsData  []SendTasksInsReq `json:"ins_data"`
}

// AddManyTasksIns 添加发送任务关联的实例id
func AddManyTasksIns(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  AddManyTasksInsReq
	)

	currentUser := app.GetCurrentUserName(c)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	// 校验提交的各个实例信息
	var taskIns []models.SendTasksIns
	if len(req.InsData) > 0 {
		for _, data := range req.InsData {
			code, dataErrStr := app.CommonPlaygroundValid(data)
			if code != http.StatusOK {
				appG.CResponse(code, dataErrStr, nil)
				return
			}
			uuidObj, _ := uuid.Parse(data.ID)
			taskIns = append(taskIns, models.SendTasksIns{
				UUIDModel: models.UUIDModel{
					ID:         uuidObj,
					CreatedBy:  currentUser,
					ModifiedBy: currentUser,
				},
				TaskID:      data.TaskId,
				WayID:       data.WayID,
				WayType:     data.WayType,
				ContentType: data.ContentType,
				Config:      data.Config,
				Extra:       data.Extra,
			})
		}
	}

	sendTaskInsService := send_ins_service.SendTaskInsService{}
	err := sendTaskInsService.ManyAdd(taskIns)
	if err != "" {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加实例失败！错误原因：%s", err), nil)
		return
	}

	sendTaskService := send_task_service.SendTaskService{
		ID:        req.TaskId,
		Name:      req.TaskName,
		CreatedBy: currentUser,
	}
	errAdd := sendTaskService.AddWithID()
	if errAdd != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加任务失败！错误原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "添加实例成功！", nil)

}

// AddTasksIns 添加发送任务关联的实例id
func AddTasksIns(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  SendTasksInsReq
	)

	//currentUser := app.GetCurrentUserName(c)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	sendTaskInsService := send_ins_service.SendTaskInsService{}
	uuidObj, _ := uuid.Parse(req.ID)
	err := sendTaskInsService.AddOne(models.SendTasksIns{
		UUIDModel:   models.UUIDModel{ID: uuidObj},
		TaskID:      req.TaskId,
		WayID:       req.WayID,
		WayType:     req.WayType,
		ContentType: req.ContentType,
		Config:      req.Config,
		Extra:       req.Extra,
		//Cre:         req.Extra,
	})
	if err != "" {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加实例失败！错误原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "添加实例成功！", nil)

}

type UpdateMsgTaskInsEnableReq struct {
	ID     string `json:"ins_id" validate:"required,len=36" label:"实例id"`
	Enable int    `json:"status" validate:"" label:"实例开启状态"`
}

// UpdateMsgTaskInsEnable 更新消息渠道实例的是否开启
func UpdateMsgTaskInsEnable(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  UpdateMsgTaskInsEnableReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	InsService := send_ins_service.SendTaskInsService{
		ID: req.ID,
		//Enable: req.Enable,
	}
	err := InsService.Update(map[string]interface{}{
		"enable": req.Enable,
	})
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "删除实例失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "删除实例成功！", nil)
}
