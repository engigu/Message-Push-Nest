package cron_service

import (
	"fmt"
	"math/rand"
	"message-nest/models"
	"message-nest/pkg/constant"
	"message-nest/pkg/logging"
	"message-nest/pkg/util"
	"message-nest/service/send_message_service"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/unknwon/com"
)

var (
	userLogsCleanTaskMap      = make(map[string]cron.EntryID)
	userHostedMsgCleanTaskMap = make(map[string]cron.EntryID)
	cleanTasksMutex           sync.Mutex
)

// executeUserCleanTask 执行用户级清理任务的通用逻辑
func executeUserCleanTask(taskID, taskName, sectionName, enabledKey, keepNumKey, resourceName string, deleteFunc func(int, ...string) (int, error), userID string) {
	var errStr string
	sm := send_message_service.SendMessageService{
		TaskID:        taskID,
		Name:          taskName,
		CreatedBy:     userID,
		DefaultLogger: logging.Scheduler,
	}
	sm.Status = send_message_service.SendSuccess

	// 检查是否启用（支持按用户或系统回退）
	enabledSetting, err := models.GetSettingByKey(sectionName, enabledKey, userID)
	if err != nil {
		errStr = fmt.Sprintf("获取%s清理开关失败，原因：%s", resourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
		sm.RecordSendLog()
		return
	}

	if enabledSetting.Value != "true" {
		sm.LogsAndStatusMark(fmt.Sprintf("%s清理功能未启用，跳过执行", resourceName), sm.Status)
		sm.RecordSendLog()
		return
	}

	// 针对多用户任务加入 0~3 秒的随机微小抖动，彻底抹平数据库并发锁争用
	if userID != "" {
		jitterMs := rand.Intn(3000)
		time.Sleep(time.Duration(jitterMs) * time.Millisecond)
	}

	sm.LogsAndStatusMark(fmt.Sprintf("开始清除%s", resourceName), sm.Status)

	setting, err := models.GetSettingByKey(sectionName, keepNumKey, userID)
	if err != nil {
		errStr = fmt.Sprintf("获取%s的保留数失败，原因：%s", resourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
	}

	keepNum := com.StrTo(setting.Value).MustInt()
	affectedRows, err := deleteFunc(keepNum, userID)
	if err != nil {
		errStr = fmt.Sprintf("删除%s失败，原因：%s", resourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
	} else {
		errStr = fmt.Sprintf("删除%s成功，删除条数：%d，保留数目：%d", resourceName, affectedRows, keepNum)
		sm.LogsAndStatusMark(errStr, sm.Status)
	}

	sm.RecordSendLog()
}

// ClearLogs 清除全局/默认日志的定时任务（兼顾历史兼容）
func ClearLogs() {
	ClearUserLogs("")
}

// ClearUserLogs 清除指定用户的日志
func ClearUserLogs(userID string) {
	taskName := "日志定时清除"
	if userID != "" {
		taskName = fmt.Sprintf("用户[%s]日志定时清除", userID)
	}
	executeUserCleanTask(
		constant.CleanLogsTaskId,
		taskName,
		constant.LogsCleanSectionName,
		constant.LogsCleanEnabledKeyName,
		constant.LogsCleanKeepKeyName,
		"日志",
		models.DeleteOutDateLogs,
		userID,
	)
}

// ClearHostedMessages 清除全局/默认托管消息的定时任务（兼顾历史兼容）
func ClearHostedMessages() {
	ClearUserHostedMessages("")
}

// ClearUserHostedMessages 清除指定用户的托管消息
func ClearUserHostedMessages(userID string) {
	taskName := "托管消息定时清除"
	if userID != "" {
		taskName = fmt.Sprintf("用户[%s]托管消息定时清除", userID)
	}
	executeUserCleanTask(
		constant.CleanHostedMsgTaskId,
		taskName,
		constant.HostedMsgCleanSectionName,
		constant.HostedMsgCleanEnabledKeyName,
		constant.HostedMsgCleanKeepKeyName,
		"托管消息",
		models.DeleteOutDateHostedMessages,
		userID,
	)
}

type CronService struct {
}

// updateUserCleanCronTask 更新用户级清理任务的通用逻辑
func updateUserCleanCronTask(taskMap map[string]cron.EntryID, userID string, cronExpr string, enabled bool, resourceName string, job func()) {
	cleanTasksMutex.Lock()
	defer cleanTasksMutex.Unlock()

	oldTaskID, exists := taskMap[userID]
	if exists && oldTaskID > 0 {
		RemoveTask(oldTaskID)
		delete(taskMap, userID)
	}

	if enabled && cronExpr != "" {
		taskID := AddTask(ScheduledTask{
			Schedule: cronExpr,
			Job:      job,
		})
		if taskID > 0 {
			taskMap[userID] = taskID
			logging.Scheduler.Infof("更新用户[%s]%s清理任务成功，cron: %s", userID, resourceName, cronExpr)
		}
	} else {
		logging.Scheduler.Infof("用户[%s]%s清理任务已停止/未启用", userID, resourceName)
	}
}

// UpdateUserLogsCronRun 更新指定用户的日志清理定时任务
func (cs *CronService) UpdateUserLogsCronRun(userID string, cronExpr string, enabled bool) {
	updateUserCleanCronTask(userLogsCleanTaskMap, userID, cronExpr, enabled, "日志", func() {
		ClearUserLogs(userID)
	})
}

// UpdateUserHostedMsgCronRun 更新指定用户的托管消息清理定时任务
func (cs *CronService) UpdateUserHostedMsgCronRun(userID string, cronExpr string, enabled bool) {
	updateUserCleanCronTask(userHostedMsgCleanTaskMap, userID, cronExpr, enabled, "托管消息", func() {
		ClearUserHostedMessages(userID)
	})
}

// UpdateLogsCronRun 兼容全局更新日志清理定时任务
func (cs *CronService) UpdateLogsCronRun(cronExpr string, enabled bool) {
	cs.UpdateUserLogsCronRun("", cronExpr, enabled)
}

// UpdateHostedMsgCronRun 兼容全局更新托管消息清理定时任务
func (cs *CronService) UpdateHostedMsgCronRun(cronExpr string, enabled bool) {
	cs.UpdateUserHostedMsgCronRun("", cronExpr, enabled)
}

// StartLogsCronRunOnStartup 启动的时候注册所有用户的定时清除日志任务
func StartLogsCronRunOnStartup() {
	logging.Scheduler.Infof("开始注册定时清除日志任务...")
	cs := CronService{}

	// 1. 注册全局/默认日志清理任务
	cronSetting, _ := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanCronKeyName)
	enabledSetting, _ := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanEnabledKeyName)
	if cronSetting.Value != "" && enabledSetting.Value == "true" {
		cs.UpdateUserLogsCronRun("", cronSetting.Value, true)
	}

	// 2. 注册所有配置了私有日志清理规则的用户任务
	users, err := models.GetConfiguredUsersBySection(constant.LogsCleanSectionName)
	if err == nil {
		for _, u := range users {
			uCron, _ := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanCronKeyName, u)
			uEnabled, _ := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanEnabledKeyName, u)
			if uEnabled.Value == "true" {
				cronExpr := uCron.Value
				if cronExpr == "" {
					cronExpr = util.GetDefaultCronForUser(u, 0)
				}
				cs.UpdateUserLogsCronRun(u, cronExpr, true)
			}
		}
	}
}

// StartHostedMsgCronRunOnStartup 启动的时候注册所有用户的定时清除托管消息任务
func StartHostedMsgCronRunOnStartup() {
	logging.Scheduler.Infof("开始注册定时清除托管消息任务...")
	cs := CronService{}

	// 1. 注册全局/默认托管消息清理任务
	cronSetting, _ := models.GetSettingByKey(constant.HostedMsgCleanSectionName, constant.HostedMsgCleanCronKeyName)
	enabledSetting, _ := models.GetSettingByKey(constant.HostedMsgCleanSectionName, constant.HostedMsgCleanEnabledKeyName)
	if cronSetting.Value != "" && enabledSetting.Value == "true" {
		cs.UpdateUserHostedMsgCronRun("", cronSetting.Value, true)
	}

	// 2. 注册所有配置了私有托管消息清理规则的用户任务
	users, err := models.GetConfiguredUsersBySection(constant.HostedMsgCleanSectionName)
	if err == nil {
		for _, u := range users {
			uCron, _ := models.GetSettingByKey(constant.HostedMsgCleanSectionName, constant.HostedMsgCleanCronKeyName, u)
			uEnabled, _ := models.GetSettingByKey(constant.HostedMsgCleanSectionName, constant.HostedMsgCleanEnabledKeyName, u)
			if uEnabled.Value == "true" {
				cronExpr := uCron.Value
				if cronExpr == "" {
					cronExpr = util.GetDefaultCronForUser(u, 27)
				}
				cs.UpdateUserHostedMsgCronRun(u, cronExpr, true)
			}
		}
	}
}

func StartTasksRunOnStartup() {
	StartLogsCronRunOnStartup()
	StartHostedMsgCronRunOnStartup()
}

