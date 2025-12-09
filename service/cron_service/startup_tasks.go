package cron_service

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/constant"
	"message-nest/service/send_message_service"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/unknwon/com"
)

var ClearLogsTaskId cron.EntryID
var ClearHostedMsgTaskId cron.EntryID

// CleanConfig 清理任务配置
type CleanConfig struct {
	TaskID       string
	TaskName     string
	LogPrefix    string
	SectionName  string
	EnabledKey   string
	KeepNumKey   string
	DeleteFunc   func(int) (int, error)
	ResourceName string // 资源名称，如"日志"、"托管消息"
}

// executeCleanTask 执行清理任务的通用逻辑
func executeCleanTask(config CleanConfig) {
	var errStr string
	sm := send_message_service.SendMessageService{
		TaskID: config.TaskID,
		Name:   config.TaskName,
		DefaultLogger: logrus.WithFields(logrus.Fields{
			"prefix": config.LogPrefix,
		}),
	}
	sm.Status = send_message_service.SendSuccess

	// 检查是否启用
	enabledSetting, err := models.GetSettingByKey(config.SectionName, config.EnabledKey)
	if err != nil {
		errStr = fmt.Sprintf("获取%s清理开关失败，原因：%s", config.ResourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
		sm.RecordSendLog()
		return
	}

	if enabledSetting.Value != "true" {
		sm.LogsAndStatusMark(fmt.Sprintf("%s清理功能未启用，跳过执行", config.ResourceName), sm.Status)
		sm.RecordSendLog()
		return
	}

	sm.LogsAndStatusMark(fmt.Sprintf("开始清除%s", config.ResourceName), sm.Status)

	setting, err := models.GetSettingByKey(config.SectionName, config.KeepNumKey)
	if err != nil {
		errStr = fmt.Sprintf("获取%s的保留数失败，原因：%s", config.ResourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
	}

	keepNum := com.StrTo(setting.Value).MustInt()
	affectedRows, err := config.DeleteFunc(keepNum)
	if err != nil {
		errStr = fmt.Sprintf("删除%s失败，原因：%s", config.ResourceName, err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
	} else {
		errStr = fmt.Sprintf("删除%s成功，删除条数：%d，保留数目：%d", config.ResourceName, affectedRows, keepNum)
		sm.LogsAndStatusMark(errStr, sm.Status)
	}

	sm.RecordSendLog()
}

// ClearLogs 清除日志的定时任务
func ClearLogs() {
	executeCleanTask(CleanConfig{
		TaskID:       constant.CleanLogsTaskId,
		TaskName:     "日志定时清除",
		LogPrefix:    "[Cron Clear Logs]",
		SectionName:  constant.LogsCleanSectionName,
		EnabledKey:   constant.LogsCleanEnabledKeyName,
		KeepNumKey:   constant.LogsCleanKeepKeyName,
		DeleteFunc:   models.DeleteOutDateLogs,
		ResourceName: "日志",
	})
}

type CronService struct {
}

// startCleanCronTask 启动清理任务的通用逻辑
func startCleanCronTask(sectionName, enabledKey, cronKey, resourceName string, job func(), taskId *cron.EntryID) {
	// 检查是否启用
	enabledSetting, err := models.GetSettingByKey(sectionName, enabledKey)
	if err != nil {
		logrus.Error(fmt.Sprintf("获取[%s]清理开关失败，原因：%s", resourceName, err))
		return
	}

	if enabledSetting.Value != "true" {
		logrus.Info(fmt.Sprintf("[%s]清理功能未启用", resourceName))
		return
	}

	// 注册任务
	setting, err := models.GetSettingByKey(sectionName, cronKey)
	if err != nil {
		logrus.Error(fmt.Sprintf("获取[%s]的cron失败，原因：%s", resourceName, err))
		return
	}
	*taskId = AddTask(ScheduledTask{
		Schedule: setting.Value,
		Job:      job,
	})
	logrus.Info(fmt.Sprintf("[%s]清理任务已启动", resourceName))
}

// updateCleanCronTask 更新清理任务的通用逻辑
func updateCleanCronTask(cron string, enabled bool, resourceName string, job func(), taskId *cron.EntryID) {
	// 先移除旧任务
	if *taskId > 0 {
		RemoveTask(*taskId)
		*taskId = 0
	}

	// 如果启用，则添加新任务
	if enabled {
		*taskId = AddTask(ScheduledTask{
			Schedule: cron,
			Job:      job,
		})
		logrus.Info(fmt.Sprintf("更新%s的cron成功，%s", resourceName, cron))
	} else {
		logrus.Info(fmt.Sprintf("%s清理任务已停止", resourceName))
	}
	logrus.Info(fmt.Sprintf("所有的定时任务总数： %d", len(TaskList)))
}

// StartLogsCronRun 启动注册清除任务定时任务
func (cs *CronService) StartLogsCronRun() {
	startCleanCronTask(
		constant.LogsCleanSectionName,
		constant.LogsCleanEnabledKeyName,
		constant.LogsCleanCronKeyName,
		"日志",
		ClearLogs,
		&ClearLogsTaskId,
	)
}

// UpdateLogsCronRun 更新清除任务定时任务
func (cs *CronService) UpdateLogsCronRun(cron string, enabled bool) {
	updateCleanCronTask(cron, enabled, "日志", ClearLogs, &ClearLogsTaskId)
}

// ClearHostedMessages 清除托管消息的定时任务
func ClearHostedMessages() {
	executeCleanTask(CleanConfig{
		TaskID:       constant.CleanHostedMsgTaskId,
		TaskName:     "托管消息定时清除",
		LogPrefix:    "[Cron Clear Hosted Messages]",
		SectionName:  constant.HostedMsgCleanSectionName,
		EnabledKey:   constant.HostedMsgCleanEnabledKeyName,
		KeepNumKey:   constant.HostedMsgCleanKeepKeyName,
		DeleteFunc:   models.DeleteOutDateHostedMessages,
		ResourceName: "托管消息",
	})
}

// StartHostedMsgCronRun 启动注册托管消息清除任务定时任务
func (cs *CronService) StartHostedMsgCronRun() {
	startCleanCronTask(
		constant.HostedMsgCleanSectionName,
		constant.HostedMsgCleanEnabledKeyName,
		constant.HostedMsgCleanCronKeyName,
		"托管消息",
		ClearHostedMessages,
		&ClearHostedMsgTaskId,
	)
}

// UpdateHostedMsgCronRun 更新托管消息清除任务定时任务
func (cs *CronService) UpdateHostedMsgCronRun(cron string, enabled bool) {
	updateCleanCronTask(cron, enabled, "托管消息", ClearHostedMessages, &ClearHostedMsgTaskId)
}

// StartLogsCronRunOnStartup 启动的时候开启定时任务
func StartLogsCronRunOnStartup() {
	logrus.Infof("开始注册定时清除日志任务...")
	cs := CronService{}
	cs.StartLogsCronRun()
}

// StartHostedMsgCronRunOnStartup 启动的时候开启托管消息清理定时任务
func StartHostedMsgCronRunOnStartup() {
	logrus.Infof("开始注册定时清除托管消息任务...")
	cs := CronService{}
	cs.StartHostedMsgCronRun()
}

func StartTasksRunOnStartup() {
	StartLogsCronRunOnStartup()
	StartHostedMsgCronRunOnStartup()
}
