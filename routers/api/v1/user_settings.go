package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/service/settings_service"
	"net/http"
)

type GetSettingReq struct {
	Section string `json:"section" validate:"required,max=50" label:"节点"`
}

// GetUserSetting 用户重设密码
func GetUserSetting(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)
	section := c.Query("section")

	settingService := settings_service.UserSettings{}
	settings, err := settingService.GetUserSetting(section)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("获取失败！错误原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取成功！", settings)

}

// EditUserPasswdReq 编辑设置请求
type EditUserPasswdReq struct {
	OldPassword string `json:"old_passwd" validate:"required,max=50" label:"旧密码"`
	NewPassword string `json:"new_passwd" validate:"required,max=50" label:"新密码"`
}

// EditPasswd 用户重设密码
func EditPasswd(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  EditUserPasswdReq
	)

	currentUser := app.GetCurrentUserName(c)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	settingService := settings_service.UserSettings{
		UserName:    currentUser,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}
	err := settingService.EditUserPasswd()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("修改密码失败！错误原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "修改成功！", nil)

}

type SettingReq struct {
	Section string `json:"section" validate:"required,max=50" label:"一级分类"`
	Data    map[string]string
}

// EditSettings 编辑设置
func EditSettings(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  SettingReq
	)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	currentUser := app.GetCurrentUserName(c)
	settingService := settings_service.UserSettings{}
	diffStr := settingService.ValidateDiffSetting(req.Section, req.Data)
	if diffStr != "" {
		appG.CResponse(http.StatusBadRequest, diffStr, nil)
		return
	}
	for key, value := range req.Data {
		err := settingService.EditSettings(req.Section, key, value, currentUser)
		if err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("修改密码失败！错误原因：%s", err), nil)
			return
		}
	}

	appG.CResponse(http.StatusOK, "修改成功！", nil)

}
