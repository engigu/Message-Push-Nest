package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
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

	sendWayService := send_way_service.SendWay{
		ID: req.ID,
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

	sendWayService := send_way_service.SendWay{
		ID: id,
	}

	way, err := sendWayService.GetByID()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "获取到的渠道信息为空！", nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取渠道信息成功", way)
}

// GetMsgSendWayList 获取消息渠道列表
func GetMsgSendWayList(c *gin.Context) {
	appG := app.Gin{C: c}
	name := c.Query("name")
	type_ := c.Query("type")

	offset, limit := util.GetPageSize(c)
	sendWayService := send_way_service.SendWay{
		Name:     name,
		Type:     type_,
		PageNum:  offset,
		PageSize: limit,
	}
	ways, err := sendWayService.GetAll()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取渠道信息失败！", nil)
		return
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
	currentUser := app.GetCurrentUserName(c)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	sendWayService := send_way_service.SendWay{
		Name:       req.Name,
		Type:       req.Type,
		Auth:       req.Auth,
		CreatedBy:  currentUser,
		ModifiedBy: currentUser,
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

	currentUser := app.GetCurrentUserName(c)
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	sendWayService := send_way_service.SendWay{
		ID:         req.ID,
		Name:       req.Name,
		Type:       req.Type,
		Auth:       req.Auth,
		ModifiedBy: currentUser,
	}

	diffMsg, _ := sendWayService.ValidateDiffWay()
	if diffMsg != "" {
		appG.CResponse(http.StatusBadRequest, diffMsg, nil)
		return
	}

	err := sendWayService.Edit()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("编辑渠道信息失败！原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "编辑渠道信息成功", nil)
}

type TestSendWayReq struct {
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

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	sendWayService := send_way_service.SendWay{
		Name: req.Name,
		Type: req.Type,
		Auth: req.Auth,
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
