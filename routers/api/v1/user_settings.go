package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/service/settings_service"
	"net/http"
)

type EditUserPasswd struct {
	OldPassword string `json:"old_passwd" validate:"required,max=50" label:"旧密码"`
	NewPassword string `json:"new_passwd" validate:"required,max=50" label:"新密码"`
}

// EditPasswd 用户重设密码
func EditPasswd(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  EditUserPasswd
	)

	currentUser := app.GetCurrentUserName(c)
	errCode, errStr := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errStr, nil)
		return
	}

	sendTaskService := settings_service.UserSettings{
		UserName:    currentUser,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}
	err := sendTaskService.EditUserPasswd()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("修改密码失败！错误原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "修改成功！", nil)

}
