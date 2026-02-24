package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/pkg/util"
	"message-nest/service/auth_service"
	"message-nest/service/settings_service"
)

type AttemptInfo struct {
	Count       int
	LastAttempt time.Time
	LockUntil   time.Time
}

var loginAttempts sync.Map

const (
	MaxFailures      = 5
	FailResetTime    = 10 * time.Minute
	LockDurationTime = 30 * time.Minute
)

func init() {
	// 启动一个后台 goroutine 每 30 分钟清理一次过期的记录，防止恶意攻击导致内存泄漏
	go func() {
		for {
			time.Sleep(30 * time.Minute)
			now := time.Now()
			loginAttempts.Range(func(key, value interface{}) bool {
				info := value.(*AttemptInfo)
				// 如果距离上次尝试已经超过了重置时间，且也没有在锁定中，就可以清理掉了
				if now.Sub(info.LastAttempt) > FailResetTime && now.After(info.LockUntil) {
					loginAttempts.Delete(key)
				}
				return true
			})
		}
	}()
}

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

	ip := c.ClientIP()
	now := time.Now()
	var info *AttemptInfo
	if val, ok := loginAttempts.Load(ip); ok {
		info = val.(*AttemptInfo)
		if now.Before(info.LockUntil) {
			appG.CResponse(http.StatusForbidden, fmt.Sprintf("登录失败次数过多，请于%d分钟后再试！", int(info.LockUntil.Sub(now).Minutes())+1), nil)
			return
		}
	}

	authService := auth_service.Auth{Username: req.Username, Password: req.Password}
	isExist, err := authService.Check()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("校验失败：%s", err), nil)
		return
	}
	if !isExist {
		if info == nil {
			info = &AttemptInfo{}
		}
		// 如果距离上次失败超过设定时间，且没有处于锁定状态，则重置计数
		if now.Sub(info.LastAttempt) > FailResetTime && now.After(info.LockUntil) {
			info.Count = 0
		}
		info.Count++
		info.LastAttempt = now
		if info.Count >= MaxFailures {
			info.LockUntil = now.Add(LockDurationTime)
		}
		loginAttempts.Store(ip, info)

		appG.CResponse(http.StatusUnauthorized, "账号或密码不正确！", nil)
		return
	}

	// 成功登录则清除失败记录
	loginAttempts.Delete(ip)

	// 获取配置的 cookie 过期天数
	expDays := settings_service.GetCookieExpDays()
	token, err := util.GenerateToken(req.Username, req.Password, expDays)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("生成token失败：%s", err), nil)
		return
	}

	// 查询用户ID并记录登录日志
	if u, _ := models.GetUserByUsername(req.Username); u != nil {
		_ = models.AddLoginLog(u.ID, req.Username, c.ClientIP(), c.GetHeader("User-Agent"))
	}

	appG.CResponse(http.StatusOK, "登录成功!", map[string]string{
		"token": token,
	})
}
