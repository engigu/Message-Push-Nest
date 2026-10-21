package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
	"message-nest/pkg/constant"
	"message-nest/pkg/e"
	"message-nest/pkg/setting"
	"message-nest/service/cron_service"
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
	currentUserId := app.GetCurrentUserId(c)
	isAdmin := app.IsAdmin(c)

	settingService := settings_service.UserSettings{
		UserID:  currentUserId,
		IsAdmin: isAdmin,
	}
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

	if setting.IsDemoMode() {
		appG.CResponse(http.StatusForbidden, "演示模式下已禁止修改密码！", nil)
		return
	}

	currentUserId := app.GetCurrentUserId(c)
	currentUserName := app.GetCurrentUserName(c)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	settingService := settings_service.UserSettings{
		UserName:    currentUserName,
		UserID:      currentUserId,
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

	if setting.IsDemoMode() {
		appG.CResponse(http.StatusForbidden, "演示模式下禁止修改配置！", nil)
		return
	}

	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	currentUserId := app.GetCurrentUserId(c)
	isAdmin := app.IsAdmin(c)

	// 非管理员用户只允许修改用户级配置（日志清理、托管消息清理）
	if !isAdmin {
		if req.Section != constant.LogsCleanSectionName && req.Section != constant.HostedMsgCleanSectionName {
			appG.CResponse(http.StatusForbidden, "普通用户无权修改系统级配置！", nil)
			return
		}
	}

	settingService := settings_service.UserSettings{
		UserID:  currentUserId,
		IsAdmin: isAdmin,
	}
	diffStr := settingService.ValidateDiffSetting(req.Section, req.Data)
	if diffStr != "" {
		appG.CResponse(http.StatusBadRequest, diffStr, nil)
		return
	}
	for key, value := range req.Data {
		err := settingService.EditSettings(req.Section, key, value, currentUserId)
		if err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("修改设置失败！错误原因：%s", err), nil)
			return
		}
	}

	msg := "修改设置成功！"
	if req.Section == constant.LogsCleanSectionName {
		msg = "修改日志清理设置成功！"
	} else if req.Section == constant.HostedMsgCleanSectionName {
		msg = "修改托管消息清理设置成功！"
	}
	appG.CResponse(http.StatusOK, msg, nil)
}

// RestDefaultSettings 重设默认设置
func RestDefaultSettings(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	currentUserId := app.GetCurrentUserId(c)
	isAdmin := app.IsAdmin(c)
	settingService := settings_service.UserSettings{
		UserID:  currentUserId,
		IsAdmin: isAdmin,
	}

	for key, value := range constant.SiteSiteDefaultValueMap {
		err := settingService.EditSettings(constant.SiteSettingSectionName, key, value, currentUserId)
		if err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("重设站点设置失败！错误原因：%s", err), nil)
			return
		}
	}

	appG.CResponse(http.StatusOK, "重设站点设置成功！", nil)

}

// ExecuteCleanNow 管理员手动触发一次系统统一数据维护清理
func ExecuteCleanNow(c *gin.Context) {
	appG := app.Gin{C: c}
	if setting.IsDemoMode() {
		appG.CResponse(http.StatusForbidden, "演示模式下禁止执行数据清理！", nil)
		return
	}
	ok, msg := cron_service.TryExecuteSystemDataClean()
	if !ok {
		appG.CResponse(http.StatusBadRequest, msg, nil)
		return
	}
	appG.CResponse(http.StatusOK, msg, nil)
}

