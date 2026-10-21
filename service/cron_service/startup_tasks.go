package cron_service

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"message-nest/models"
	"message-nest/pkg/constant"
	"message-nest/pkg/logging"
	"message-nest/service/send_message_service"

	"github.com/robfig/cron/v3"
)

var (
	systemCleanTaskEntryID cron.EntryID
	cleanTasksMutex        sync.Mutex
	isCleaningMutex        sync.Mutex
	isCleaning             bool
	lastCleanTime          time.Time
)

// TryExecuteSystemDataClean 带有防并发与冷却防抖的系统数据维护入口
func TryExecuteSystemDataClean() (bool, string) {
	isCleaningMutex.Lock()
	if isCleaning {
		isCleaningMutex.Unlock()
		return false, "系统数据维护清理任务正在执行中，请勿重复触发！"
	}
	// 5 秒冷却防抖限制
	if !lastCleanTime.IsZero() && time.Since(lastCleanTime) < 5*time.Second {
		remain := 5 - int(time.Since(lastCleanTime).Seconds())
		isCleaningMutex.Unlock()
		return false, fmt.Sprintf("触发过于频繁，请在 %d 秒冷却后再试！", remain)
	}
	isCleaning = true
	isCleaningMutex.Unlock()

	go func() {
		defer func() {
			isCleaningMutex.Lock()
			isCleaning = false
			lastCleanTime = time.Now()
			isCleaningMutex.Unlock()
		}()
		ExecuteSystemDataClean()
	}()

	return true, "已触发系统数据维护清理任务，稍后可在推送日志中查看执行结果！"
}

// getUserCleanPolicy 获取用户的单项清理策略（开关与保留条数）
func getUserCleanPolicy(section string, userID string, defaultEnabled bool, defaultKeepNum int) (bool, int) {
	enabled := defaultEnabled
	enabledSetting, err := models.GetSettingByKey(section, constant.LogsCleanEnabledKeyName, userID)
	if err == nil && enabledSetting.ID > 0 && enabledSetting.Value != "" {
		enabled = enabledSetting.Value == "true"
	}

	keepNum := defaultKeepNum
	keepSetting, err := models.GetSettingByKey(section, constant.LogsCleanKeepKeyName, userID)
	if err == nil && keepSetting.ID > 0 && keepSetting.Value != "" {
		if val, err := strconv.Atoi(keepSetting.Value); err == nil && val > 0 {
			keepNum = val
		}
	}
	return enabled, keepNum
}

// ExecuteSystemDataClean 统一调度执行系统数据维护任务：依次按各用户设定的策略独立清理超额数据
func ExecuteSystemDataClean() {
	startTime := time.Now()
	users, err := models.GetAllActiveUsers()
	if err != nil {
		logging.Scheduler.Errorf("获取系统用户列表失败，中止本次数据清理维护: %v", err)
		return
	}

	totalCleanedLogs := 0
	totalCleanedHostedMsg := 0
	var details []string

	for _, u := range users {
		userLogsCleaned := 0
		userMsgCleaned := 0

		// 1. 按该用户的私有策略清理推送日志 (默认启用，保留 1000 条)
		logsEnabled, logsKeepNum := getUserCleanPolicy(constant.LogsCleanSectionName, u.ID, true, 1000)
		if logsEnabled && logsKeepNum > 0 {
			affected, err := models.DeleteOutDateLogs(logsKeepNum, u.ID)
			if err == nil {
				userLogsCleaned = affected
				totalCleanedLogs += affected
			} else {
				logging.Scheduler.Errorf("清理用户[%s]推送日志失败: %v", u.Username, err)
			}
		}

		// 2. 按该用户的私有策略清理托管消息 (默认关闭，保留 50000 条)
		msgEnabled, msgKeepNum := getUserCleanPolicy(constant.HostedMsgCleanSectionName, u.ID, false, 50000)
		if msgEnabled && msgKeepNum > 0 {
			affected, err := models.DeleteOutDateHostedMessages(msgKeepNum, u.ID)
			if err == nil {
				userMsgCleaned = affected
				totalCleanedHostedMsg += affected
			} else {
				logging.Scheduler.Errorf("清理用户[%s]托管消息失败: %v", u.Username, err)
			}
		}

		if userLogsCleaned > 0 || userMsgCleaned > 0 {
			details = append(details, fmt.Sprintf("%s(日志-%d,托管-%d)", u.Username, userLogsCleaned, userMsgCleaned))
		}
	}

	duration := time.Since(startTime).Round(time.Millisecond)
	if duration == 0 {
		duration = time.Millisecond
	}
	summary := fmt.Sprintf("系统数据定时清理完成: 耗时 %v，共清理日志 %d 条，托管消息 %d 条", duration, totalCleanedLogs, totalCleanedHostedMsg)
	if len(details) > 0 {
		summary += "；明细: " + strings.Join(details, "，")
	}
	logging.Scheduler.Infof("%s", summary)

	// 记录一次系统维护日志
	sm := send_message_service.SendMessageService{
		TaskID:        constant.CleanLogsTaskId,
		Name:          "系统数据定时维护清理",
		DefaultLogger: logging.Scheduler,
		Status:        send_message_service.SendSuccess,
	}
	sm.LogsAndStatusMark(summary, send_message_service.SendSuccess)
	sm.RecordSendLog()
}

// GetSystemCleanCron 获取当前系统统一维护清理任务的执行 Cron 表达式
func GetSystemCleanCron() string {
	s, err := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanCronKeyName)
	if err == nil && s.ID > 0 && strings.TrimSpace(s.Value) != "" {
		return strings.TrimSpace(s.Value)
	}
	return constant.DefaultSystemCleanCron
}

// UpdateSystemCleanCron 更新系统统一维护清理任务的执行周期
func UpdateSystemCleanCron(cronExpr string) {
	cleanTasksMutex.Lock()
	defer cleanTasksMutex.Unlock()

	if systemCleanTaskEntryID > 0 {
		RemoveTask(systemCleanTaskEntryID)
		systemCleanTaskEntryID = 0
	}

	expr := strings.TrimSpace(cronExpr)
	if expr == "" {
		expr = constant.DefaultSystemCleanCron
	}

	entryID := AddTask(ScheduledTask{
		Schedule: expr,
		Job:      ExecuteSystemDataClean,
	})
	if entryID > 0 {
		systemCleanTaskEntryID = entryID
		logging.Scheduler.Infof("更新系统数据清理维护定时任务成功，执行周期: %s", expr)
	} else {
		logging.Scheduler.Errorf("注册系统数据清理维护定时任务失败，非法表达式: %s", expr)
	}
}

type CronService struct {
}

// UpdateUserLogsCronRun 保持接口兼容（若传递了合法 cron 则更新系统统一维护周期）
func (cs *CronService) UpdateUserLogsCronRun(userID string, cronExpr string, enabled bool) {
	if cronExpr != "" {
		UpdateSystemCleanCron(cronExpr)
	}
}

// UpdateUserHostedMsgCronRun 保持接口兼容
func (cs *CronService) UpdateUserHostedMsgCronRun(userID string, cronExpr string, enabled bool) {
	if cronExpr != "" {
		UpdateSystemCleanCron(cronExpr)
	}
}

// UpdateLogsCronRun 更新系统数据清理定时维护任务
func (cs *CronService) UpdateLogsCronRun(cronExpr string, enabled bool) {
	UpdateSystemCleanCron(cronExpr)
}

// UpdateHostedMsgCronRun 兼容接口
func (cs *CronService) UpdateHostedMsgCronRun(cronExpr string, enabled bool) {
	UpdateSystemCleanCron(cronExpr)
}

// ClearLogs 手动触发一次全局数据清理（兼容保留）
func ClearLogs() {
	ExecuteSystemDataClean()
}

// ClearHostedMessages 手动触发一次全局数据清理（兼容保留）
func ClearHostedMessages() {
	ExecuteSystemDataClean()
}

// ClearUserLogs 清理指定用户的超额日志（供单测或单独调用）
func ClearUserLogs(userID string) {
	_, keepNum := getUserCleanPolicy(constant.LogsCleanSectionName, userID, true, 1000)
	_, _ = models.DeleteOutDateLogs(keepNum, userID)
}

// ClearUserHostedMessages 清理指定用户的超额托管消息（供单测或单独调用）
func ClearUserHostedMessages(userID string) {
	_, keepNum := getUserCleanPolicy(constant.HostedMsgCleanSectionName, userID, false, 50000)
	_, _ = models.DeleteOutDateHostedMessages(keepNum, userID)
}

// StartTasksRunOnStartup 启动时注册系统统一数据维护定时任务（单一守护 Job）
func StartTasksRunOnStartup() {
	cronExpr := GetSystemCleanCron()
	UpdateSystemCleanCron(cronExpr)
}
