package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/pkg/util"
	"message-nest/service/send_way_service"
)

type DeleteMsgSendWayReq struct {
	ID string `json:"id" validate:"required,len=12" label:"渠道id"`
}

// DeleteMsgSendWay 删除消息渠道
func DeleteMsgSendWay(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  DeleteMsgSendWayReq
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

	if _, err := models.GetWayByIDForOwner(req.ID, ownerID); err != nil {
		appG.CResponse(http.StatusNotFound, "发信渠道不存在或无权访问", nil)
		return
	}

	sendWayService := send_way_service.SendWay{
		ID:         req.ID,
		CreatedBy:  ownerID,
		ModifiedBy: user.GetAuditUserID(),
		IsAdmin:    user.IsAdmin() && !user.IsImpersonating(),
	}
	err := sendWayService.Delete()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("删除发信渠道失败！%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "删除发信渠道成功！", nil)
}

// GetMsgSendWay 获取消息渠道
func GetMsgSendWay(c *gin.Context) {
	appG := app.Gin{C: c}
	id := c.Query("id")

	if id == "" {
		appG.CResponse(http.StatusBadRequest, "渠道id为空！", nil)
		return
	}

	_, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	way, err := models.GetWayByIDForOwner(id, ownerID)
	if err != nil {
		appG.CResponse(http.StatusNotFound, "发信渠道不存在或无权访问", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取渠道信息成功", way)
}

// GetMsgSendWayList 获取消息渠道列表
func GetMsgSendWayList(c *gin.Context) {
	appG := app.Gin{C: c}
	name := c.Query("name")
	type_ := c.Query("type")

	user, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	offset, limit := util.GetPageSize(c)
	sendWayService := send_way_service.SendWay{
		Name:      name,
		Type:      type_,
		CreatedBy: ownerID,
		IsAdmin:   user.IsAdmin() && !user.IsImpersonating(),
		PageNum:   offset,
		PageSize:  limit,
	}
	ways, err := sendWayService.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取渠道信息失败！", nil)
		return
	}

	// 敏感凭据脱敏（密码、Secret、Token等）
	for i := range ways {
		ways[i].Auth = send_way_service.MaskWayAuth(ways[i].Auth)
	}

	count, err := sendWayService.Count()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取渠道信息总数失败！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取渠道信息成功", map[string]interface{}{
		"lists": ways,
		"total": count,
	})
}

type AddMsgSendWayReq struct {
	Name string `json:"name" validate:"required,max=100,min=1" label:"渠道名"`
	Type string `json:"type" validate:"required,max=100,min=1" label:"渠道类型"`
	Auth string `json:"auth" label:"渠道认证方式"`
}

// AddMsgSendWay 添加发送渠道
func AddMsgSendWay(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  AddMsgSendWayReq
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
	sendWayService := send_way_service.SendWay{
		Name:       req.Name,
		Type:       req.Type,
		Auth:       req.Auth,
		CreatedBy:  targetOwner,
		ModifiedBy: user.GetAuditUserID(),
		IsAdmin:    user.IsAdmin() && !user.IsImpersonating(),
	}

	diffMsg, _ := sendWayService.ValidateDiffWay()
	if diffMsg != "" {
		appG.CResponse(http.StatusBadRequest, diffMsg, nil)
		return
	}

	err := sendWayService.Add()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加渠道失败！原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "添加渠道成功！", nil)
}

type EditSendWayReq struct {
	ID   string `json:"id" validate:"required,len=12" label:"渠道id"`
	Name string `json:"name" validate:"required,max=100,min=1" label:"渠道名"`
	Type string `json:"type" validate:"required,max=100,min=1" label:"渠道类型"`
	Auth string `json:"auth" validate:"required" label:"渠道认证信息"`
}

// EditSendWay 编辑发送渠道
func EditSendWay(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  = EditSendWayReq{}
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

	existingWay, err := models.GetWayByIDForOwner(req.ID, ownerID)
	if err != nil {
		appG.CResponse(http.StatusNotFound, "发信渠道不存在或无权访问", nil)
		return
	}

	// 合并脱敏字段：若提交的字段值为掩码 ******，自动还原保留数据库原有凭据
	mergedAuth := send_way_service.MergeWayAuth(req.Auth, existingWay.Auth)

	sendWayService := send_way_service.SendWay{
		ID:         req.ID,
		Name:       req.Name,
		Type:       req.Type,
		Auth:       mergedAuth,
		CreatedBy:  ownerID,
		ModifiedBy: user.GetAuditUserID(),
		IsAdmin:    user.IsAdmin() && !user.IsImpersonating(),
	}

	diffMsg, _ := sendWayService.ValidateDiffWay()
	if diffMsg != "" {
		appG.CResponse(http.StatusBadRequest, diffMsg, nil)
		return
	}

	err = sendWayService.Edit()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("编辑渠道信息失败！原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "编辑渠道信息成功", nil)
}

type TestSendWayReq struct {
	ID   string `json:"id" validate:"omitempty,len=12" label:"渠道id"`
	Name string `json:"name" validate:"required,max=100,min=1" label:"渠道名"`
	Type string `json:"type" validate:"required,max=100,min=1" label:"渠道类型"`
	Auth string `json:"auth" validate:"required" label:"渠道认证信息"`
}

// TestSendWay 测试发送渠道
func TestSendWay(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  = TestSendWayReq{}
	)

	_, ownerID, ok := app.GetRequestOwnerID(c.Request.Context())
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	authPayload := req.Auth
	if req.ID != "" {
		if existingWay, err := models.GetWayByIDForOwner(req.ID, ownerID); err == nil {
			authPayload = send_way_service.MergeWayAuth(req.Auth, existingWay.Auth)
		}
	}

	sendWayService := send_way_service.SendWay{
		Name: req.Name,
		Type: req.Type,
		Auth: authPayload,
	}

	diffMsg, msgObj := sendWayService.ValidateDiffWay()
	if diffMsg != "" {
		appG.CResponse(http.StatusBadRequest, diffMsg, nil)
		return
	}

	errTestMsg, resText := sendWayService.TestSendWay(msgObj)
	if errTestMsg != "" {
		appG.CResponse(http.StatusInternalServerError, errTestMsg, nil)
		return
	}
	msg := "测试渠道信息成功"
	if resText != "" {
		msg += ", 返回信息：" + resText
	}
	appG.CResponse(http.StatusOK, msg, nil)
}
