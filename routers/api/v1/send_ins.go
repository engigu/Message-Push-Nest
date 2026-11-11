package v1

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_task_service"
)

type DeleteMsgTaskInsReq struct {
	ID string `json:"id" validate:"required,len=12" label:"实例id"`
}

// DeleteMsgTaskIns 删除消息渠道实例
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

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	InsService := send_ins_service.SendTaskInsService{
		ID:        req.ID,
		CreatedBy: ownerID,
		IsAdmin:   user.IsAdmin() && !user.IsImpersonating(),
	}
	err := InsService.Delete()
	if err != nil {
		if errors.Is(err, models.ErrNoPermissionOrNotFound) {
			appG.CResponse(http.StatusNotFound, "实例不存在或无权操作！", nil)
			return
		}
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

	_, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	task, err := models.GetTasksInsForOwner(id, ownerID)
	if err != nil {
		appG.CResponse(http.StatusNotFound, "获取到的任务信息为空或无权访问！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取任务信息成功", task)
}

type SendTasksInsReq struct {
	ID          string `json:"id" validate:"required,len=12" label:"实例id"`
	TaskId      string `json:"task_id" validate:"required,len=12" label:"ins任务id"`
	WayID       string `json:"way_id" validate:"required,len=12" label:"渠道id"`
	ContentType string `json:"content_type" validate:"required,max=100" label:"实例内容类型"`
	Config      string `json:"config" validate:"" label:"任务配置"`
	Extra       string `json:"extra" validate:"" label:"任务额外信息"`
	WayType     string `json:"way_type" validate:"required,max=100" label:"渠道类型"`
}

type AddManyTasksInsReq struct {
	TaskId   string            `json:"id" validate:"required,len=12" label:"任务id"`
	TaskName string            `json:"name" validate:"required,max=100" label:"任务名"`
	InsData  []SendTasksInsReq `json:"ins_data"`
}

// AddManyTasksIns 添加发送任务关联的实例id
func AddManyTasksIns(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  AddManyTasksInsReq
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

	// 校验目标任务归属：如果已存在，必须归属当前租户
	if existingTask, err := models.GetTaskByID(req.TaskId); err == nil && existingTask.ID != "" {
		if ownerID != "" && existingTask.CreatedBy != ownerID {
			appG.CResponse(http.StatusForbidden, "任务已被其他租户占用，无权关联实例！", nil)
			return
		}
	}

	// 校验提交的各个实例信息及其关联渠道归属
	targetOwner := user.GetEffectiveUserID()
	auditUserID := user.GetAuditUserID()
	var taskIns []models.SendTasksIns

	if len(req.InsData) > 0 {
		for _, data := range req.InsData {
			code, dataErrStr := app.CommonPlaygroundValid(data)
			if code != http.StatusOK {
				appG.CResponse(code, dataErrStr, nil)
				return
			}

			// 校验渠道归属，防止跨租户盗用渠道
			way, err := models.GetWayByIDForOwner(data.WayID, ownerID)
			if err != nil || way.Type != data.WayType {
				appG.CResponse(http.StatusNotFound, fmt.Sprintf("发信渠道 %s 不存在或无权访问！", data.WayID), nil)
				return
			}

			taskIns = append(taskIns, models.SendTasksIns{
				UUIDModel: models.UUIDModel{
					ID:         data.ID,
					CreatedBy:  targetOwner,
					ModifiedBy: auditUserID,
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
		ID:         req.TaskId,
		Name:       req.TaskName,
		CreatedBy:  targetOwner,
		ModifiedBy: auditUserID,
	}
	errAdd := sendTaskService.AddWithID()
	if errAdd != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加任务失败！错误原因：%s", errAdd), nil)
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

	// 校验目标任务归属
	if existingTask, err := models.GetTaskByIDForOwner(req.TaskId, ownerID); err != nil || existingTask.ID == "" {
		appG.CResponse(http.StatusNotFound, "发信任务不存在或无权访问！", nil)
		return
	}

	// 校验渠道归属
	way, err := models.GetWayByIDForOwner(req.WayID, ownerID)
	if err != nil || way.Type != req.WayType {
		appG.CResponse(http.StatusNotFound, "发信渠道不存在或无权访问！", nil)
		return
	}

	targetOwner := user.GetEffectiveUserID()
	auditUserID := user.GetAuditUserID()

	sendTaskInsService := send_ins_service.SendTaskInsService{}
	errStr = sendTaskInsService.AddOne(models.SendTasksIns{
		UUIDModel: models.UUIDModel{
			ID:         req.ID,
			CreatedBy:  targetOwner,
			ModifiedBy: auditUserID,
		},
		TaskID:      req.TaskId,
		WayID:       req.WayID,
		WayType:     req.WayType,
		ContentType: req.ContentType,
		Config:      req.Config,
		Extra:       req.Extra,
	})
	if errStr != "" {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加实例失败！错误原因：%s", errStr), nil)
		return
	}

	appG.CResponse(http.StatusOK, "添加实例成功！", nil)
}

type UpdateMsgTaskInsEnableReq struct {
	ID     string `json:"ins_id" validate:"required,len=12" label:"实例id"`
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

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	InsService := send_ins_service.SendTaskInsService{
		ID:        req.ID,
		CreatedBy: ownerID,
		IsAdmin:   user.IsAdmin() && !user.IsImpersonating(),
	}
	err := InsService.Update(map[string]interface{}{
		"enable":      req.Enable,
		"modified_by": user.GetAuditUserID(),
	})
	if err != nil {
		if errors.Is(err, models.ErrNoPermissionOrNotFound) {
			appG.CResponse(http.StatusNotFound, "实例不存在或无权操作！", nil)
			return
		}
		appG.CResponse(http.StatusBadRequest, "更新实例失败！", nil)
		return
	}

	msg := ""
	if req.Enable == 1 {
		msg = "设置实例发送为开启成功！"
	} else {
		msg = "设置实例发送为暂停成功！"
	}
	appG.CResponse(http.StatusOK, msg, nil)
}
