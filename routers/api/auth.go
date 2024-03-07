package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/pkg/util"
	"message-nest/service/auth_service"
)

type auth struct {
	Username string `valid:"Required; MaxSize(50)"`
	Password string `valid:"Required; MaxSize(50)"`
}

type ReqAuth struct {
	Username string `json:"username" validate:"required,max=50" label:"用户名"`
	Password string `json:"passwd" validate:"required,max=50" label:"密码"`
}

func GetAuth(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  ReqAuth
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	authService := auth_service.Auth{Username: req.Username, Password: req.Password}
	isExist, err := authService.Check()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("校验失败：%s", err), nil)
		return
	}
	if !isExist {
		appG.CResponse(http.StatusUnauthorized, "账号或密码不正确！", nil)
		return
	}

	token, err := util.GenerateToken(req.Username, req.Password)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("生成token失败：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "登录成功!", map[string]string{
		"token": token,
	})
}
