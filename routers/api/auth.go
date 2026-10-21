package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/captcha"
	"message-nest/pkg/e"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"message-nest/service/auth_service"
	"message-nest/service/settings_service"
)

type AttemptInfo struct {
	Count       uint16
	LastAttempt int64 // 秒级时间戳，降低内存开销
	LockUntil   int64 // 秒级时间戳
}

var loginAttempts sync.Map

const (
	MaxFailures         = 5
	FailResetSeconds    = 10 * 60 // 10 分钟重置试错次数
	LockDurationSeconds = 30 * 60 // 30 分钟熔断锁定
	MaxAttemptsCapacity = 2000    // 最大记录上限，严格锁死内存占用
)

// CleanExpiredLoginAttempts 由系统统一 cron 周期性调度，清理过期无用的试错记录
func CleanExpiredLoginAttempts() {
	now := time.Now().Unix()
	loginAttempts.Range(func(key, value interface{}) bool {
		info, ok := value.(*AttemptInfo)
		if !ok || (now-info.LastAttempt > FailResetSeconds && now >= info.LockUntil) {
			loginAttempts.Delete(key)
		}
		return true
	})
}

type auth struct {
	Username string `valid:"Required; MaxSize(50)"`
	Password string `valid:"Required; MaxSize(50)"`
}

type ReqAuth struct {
	Username    string `json:"username" validate:"required,max=50" label:"用户名"`
	Password    string `json:"passwd" validate:"required,max=50" label:"密码"`
	CaptchaID   string `json:"captcha_id" label:"验证码ID"`
	CaptchaCode string `json:"captcha_code" label:"验证码"`
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
	usernameKey := "user:" + req.Username
	now := time.Now().Unix()

	// 1. 检查客户端 IP 是否处于锁定冷却期
	if val, ok := loginAttempts.Load(ip); ok {
		info := val.(*AttemptInfo)
		if now < info.LockUntil {
			remainMin := int((info.LockUntil-now)/60) + 1
			appG.CResponse(http.StatusTooManyRequests, fmt.Sprintf("当前网络尝试登录失败过多，请于 %d 分钟后再试！", remainMin), nil)
			return
		}
	}

	// 2. 检查目标账号是否正被分布式撞库攻击处于锁定状态
	if val, ok := loginAttempts.Load(usernameKey); ok {
		info := val.(*AttemptInfo)
		if now < info.LockUntil {
			remainMin := int((info.LockUntil-now)/60) + 1
			appG.CResponse(http.StatusTooManyRequests, fmt.Sprintf("该账号尝试登录失败过多已被临时锁定，请于 %d 分钟后再试！", remainMin), nil)
			return
		}
	}

	// 3. 人机安全验证码校验 (若站点设置开启了验证码且非演示模式)
	if settings_service.IsCaptchaEnabled() && !setting.IsDemoMode() {
		if req.CaptchaID == "" || req.CaptchaCode == "" || !captcha.VerifyDirect(req.CaptchaID, req.CaptchaCode) {
			appG.CResponse(http.StatusBadRequest, "验证码错误或已失效，请重新输入！", nil)
			return
		}
	}

	authService := auth_service.Auth{Username: req.Username, Password: req.Password}
	isExist, err := authService.Check()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("校验失败：%s", err), nil)
		return
	}

	// 记录单次失败并检查是否触发熔断锁定的辅助函数
	recordFailure := func(key string) {
		var info *AttemptInfo
		if val, ok := loginAttempts.Load(key); ok {
			info = val.(*AttemptInfo)
		} else {
			info = &AttemptInfo{}
		}
		if now-info.LastAttempt > FailResetSeconds && now >= info.LockUntil {
			info.Count = 0
		}
		info.Count++
		info.LastAttempt = now
		if info.Count >= MaxFailures {
			info.LockUntil = now + LockDurationSeconds
		}
		loginAttempts.Store(key, info)
	}

	if !isExist {
		recordFailure(ip)
		recordFailure(usernameKey)
		appG.CResponse(http.StatusUnauthorized, "账号或密码不正确！", nil)
		return
	}

	u, err := models.GetUserByUsername(req.Username)
	if err != nil || u == nil {
		recordFailure(ip)
		recordFailure(usernameKey)
		appG.CResponse(http.StatusUnauthorized, "账号或密码不正确！", nil)
		return
	}
	if u.Status != 1 {
		appG.CResponse(http.StatusForbidden, "账号已被禁用，请联系管理员！", nil)
		return
	}

	// 成功登录，立即清除当前 IP 及该账号的所有失败锁定记录
	loginAttempts.Delete(ip)
	loginAttempts.Delete(usernameKey)

	// 获取配置的 cookie 过期天数
	expDays := settings_service.GetCookieExpDays()
	token, err := util.GenerateToken(u.ID, u.Username, req.Password, u.Role, expDays)
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("生成token失败：%s", err), nil)
		return
	}

	// 查询用户ID并记录登录日志
	_ = models.AddLoginLog(u.ID, req.Username, c.ClientIP(), c.GetHeader("User-Agent"))

	appG.CResponse(http.StatusOK, "登录成功!", map[string]string{
		"token": token,
	})
}
