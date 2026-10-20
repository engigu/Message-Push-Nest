package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/constant"
	"message-nest/pkg/e"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"message-nest/service/auth_service"
	"message-nest/service/settings_service"
)

// GetUserList 获取用户列表（仅管理员）
func GetUserList(c *gin.Context) {
	appG := app.Gin{C: c}
	search := c.Query("search")
	offset, limit := util.GetPageSize(c)

	userService := auth_service.UserService{
		Search:   search,
		PageNum:  offset,
		PageSize: limit,
	}

	users, total, err := userService.GetUserList()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取用户列表失败: %v", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取成功", map[string]interface{}{
		"lists": users,
		"total": total,
	})
}

type AddUserReq struct {
	Username string `json:"username" validate:"required,min=2,max=50" label:"用户名"`
	Password string `json:"password" validate:"required,min=4,max=50" label:"密码"`
	Email    string `json:"email" validate:"omitempty,email,max=100" label:"邮箱"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user" label:"角色"`
}

// AddUser 添加用户（仅管理员）
func AddUser(c *gin.Context) {
	appG := app.Gin{C: c}
	var req AddUserReq
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	currentUserID := app.GetCurrentUserId(c)
	userService := auth_service.UserService{
		Username:  req.Username,
		Password:  req.Password,
		Email:     req.Email,
		Role:      req.Role,
		CreatedBy: currentUserID,
	}

	user, err := userService.AddUser()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("添加用户失败: %v", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "添加用户成功！", map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

type EditUserReq struct {
	ID     string `json:"id" validate:"required" label:"用户ID"`
	Email  string `json:"email" validate:"omitempty,email,max=100" label:"邮箱"`
	Role   string `json:"role" validate:"required,oneof=admin user" label:"角色"`
	Status int    `json:"status" validate:"oneof=0 1" label:"状态"`
}

// EditUser 编辑用户信息（仅管理员）
func EditUser(c *gin.Context) {
	appG := app.Gin{C: c}
	var req EditUserReq
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	currentUserID := app.GetCurrentUserId(c)
	userService := auth_service.UserService{
		ID:         req.ID,
		Email:      req.Email,
		Role:       req.Role,
		Status:     req.Status,
		ModifiedBy: currentUserID,
	}

	err := userService.EditUser()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("修改用户失败: %v", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "修改成功！", nil)
}

type DeleteUserReq struct {
	ID string `json:"id" validate:"required" label:"用户ID"`
}

// DeleteUser 删除用户（仅管理员）
func DeleteUser(c *gin.Context) {
	appG := app.Gin{C: c}
	var req DeleteUserReq
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	currentUserID := app.GetCurrentUserId(c)
	userService := auth_service.UserService{
		ID: req.ID,
	}

	err := userService.DeleteUser(currentUserID)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("删除用户失败: %v", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "删除成功！", nil)
}

type ResetUserPasswdReq struct {
	ID       string `json:"id" validate:"required" label:"用户ID"`
	Password string `json:"password" validate:"required,min=4,max=50" label:"新密码"`
}

// ResetUserPassword 重置用户密码（仅管理员）
func ResetUserPassword(c *gin.Context) {
	appG := app.Gin{C: c}
	var req ResetUserPasswdReq
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	currentUserID := app.GetCurrentUserId(c)
	userService := auth_service.UserService{
		ID:         req.ID,
		Password:   req.Password,
		ModifiedBy: currentUserID,
	}

	err := userService.ResetPassword()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("重置密码失败: %v", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "密码重置成功！", nil)
}

type RegisterReq struct {
	Username string `json:"username" validate:"required,min=2,max=30" label:"用户名"`
	Password string `json:"password" validate:"required,min=4,max=50" label:"密码"`
	Email    string `json:"email" validate:"omitempty,email,max=100" label:"邮箱"`
}

// Register 访客自主注册接口（公开免鉴权）
func Register(c *gin.Context) {
	appG := app.Gin{C: c}
	if setting.IsDemoMode() {
		appG.CResponse(http.StatusForbidden, "演示模式下已禁用自主注册功能", nil)
		return
	}

	var req RegisterReq
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	userService := auth_service.UserService{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	}

	user, err := userService.Register()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}

	// 注册成功后直接签发 Token 实现自动登录
	expDays := settings_service.GetCookieExpDays()
	token, err := util.GenerateToken(user.ID, user.Username, user.Password, user.Role, expDays)
	if err != nil {
		appG.CResponse(http.StatusOK, "注册成功，但自动生成登录凭证失败，请手动登录", map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"token":    "",
		})
		return
	}

	// 记录登录日志
	_ = models.AddLoginLog(user.ID, user.Username, c.ClientIP(), c.GetHeader("User-Agent"))

	appG.CResponse(http.StatusOK, "注册成功！", map[string]interface{}{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
		"token":    token,
	})
}

// GetPublicSiteConfig 获取公开站点信息（免鉴权，供登录页/注册页展示）
func GetPublicSiteConfig(c *gin.Context) {
	appG := app.Gin{C: c}
	userSettings := settings_service.UserSettings{}
	configs, err := userSettings.GetUserSetting(constant.SiteSettingSectionName)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "获取站点配置失败", nil)
		return
	}

	// 过滤仅暴露给公开界面的字段
	publicData := map[string]interface{}{
		"title":                 configs["title"],
		"slogan":                configs["slogan"],
		"logo":                  configs["logo"],
		"theme_color":           configs["theme_color"],
		"allow_registration":    settings_service.IsRegistrationAllowed(),
		"enable_public_preview": configs["enable_public_preview"] == "true",
		"is_demo":               setting.IsDemoMode(),
	}

	appG.CResponse(http.StatusOK, "获取成功", publicData)
}

// GetSSOProfile 获取当前用户的 SSO 绑定信息
func GetSSOProfile(c *gin.Context) {
	appG := app.Gin{C: c}
	currentUserID := app.GetCurrentUserId(c)
	userService := auth_service.UserService{
		ID: currentUserID,
	}
	profile, err := userService.GetSSOProfile()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}
	appG.CResponse(http.StatusOK, "获取成功", profile)
}

type UnbindMySSOReq struct {
	Provider string `json:"provider" label:"身份提供商类型"`
}

// UnbindMySSO 当前用户解绑自身的 SSO (可按 Provider 解除)
func UnbindMySSO(c *gin.Context) {
	appG := app.Gin{C: c}
	currentUserID := app.GetCurrentUserId(c)

	var req UnbindMySSOReq
	_ = c.ShouldBindJSON(&req)
	if req.Provider == "" {
		req.Provider = c.Query("provider")
	}

	userService := auth_service.UserService{
		ID: currentUserID,
	}
	err := userService.UnbindSSO(req.Provider)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "解除绑定失败: "+err.Error(), nil)
		return
	}
	appG.CResponse(http.StatusOK, "已成功解除该单点登录绑定！", nil)
}

type UnbindUserSSOReq struct {
	ID       string `json:"id" validate:"required" label:"用户ID"`
	Provider string `json:"provider" label:"身份提供商类型"`
}

// AdminUnbindUserSSO 管理员解绑指定用户的 SSO
func AdminUnbindUserSSO(c *gin.Context) {
	appG := app.Gin{C: c}
	var req UnbindUserSSOReq
	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	userService := auth_service.UserService{
		ID: req.ID,
	}
	err := userService.UnbindSSO(req.Provider)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "解除绑定失败: "+err.Error(), nil)
		return
	}
	appG.CResponse(http.StatusOK, "已成功解除该用户的单点登录绑定！", nil)
}

// CheckUsernameAvailability 实时校验用户名是否可用 (公开免鉴权，供注册/SSO绑定/后台添加时排重提示)
func CheckUsernameAvailability(c *gin.Context) {
	appG := app.Gin{C: c}
	username := c.Query("username")
	excludeID := c.Query("exclude_id")

	available, reason := models.CheckUsernameAvailable(username, excludeID)
	appG.CResponse(http.StatusOK, "查询成功", gin.H{
		"available": available,
		"reason":    reason,
		"username":  username,
	})
}

