package settings_service

import (
	"errors"
	"message-nest/models"
)

type UserSettings struct {
	UserName    string
	OldPassword string
	NewPassword string
}

// EditUserPasswd 用户设置密码
func (us *UserSettings) EditUserPasswd() error {
	ok, _ := models.CheckAuth(us.UserName, us.OldPassword)
	if !ok {
		return errors.New("旧密码校验失败！")
	}
	var user = make(map[string]string)
	user["password"] = us.NewPassword
	return models.EditUser(us.UserName, user)
}
