package cron_service

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/unknwon/com"
	"message-nest/models"
	"message-nest/pkg/constant"
	"message-nest/pkg/logging"
	"message-nest/service/send_message_service"
)

type CronService struct {
}

var ClearLogsTaskId cron.EntryID

// 清除日志的定时任务
func ClearLogs() {
	var errStr string
	sm := send_message_service.SendMessageService{TaskID: constant.CleanLogsTaskId}
	sm.Status = send_message_service.SendSuccess

	sm.LogsAndStatusMark("开始清除日志", sm.Status)

	setting, err := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanKeepKeyName)
	if err != nil {
		errStr = fmt.Sprintf("获取日志的保留数失败，原因：%s", err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
	}

	keepNum := com.StrTo(setting.Value).MustInt()
	affectedRows, err := models.DeleteOutDateLogs(keepNum)
	if err != nil {
		errStr = fmt.Sprintf("删除日志失败，原因：%s", err)
		sm.LogsAndStatusMark(errStr, send_message_service.SendFail)
	} else {
		errStr = fmt.Sprintf("删除日志成功，删除条数：%d，保留数目：%d", affectedRows, keepNum)
		sm.LogsAndStatusMark(errStr, sm.Status)
	}

	sm.RecordSendLog()
}

// 启动注册清除任务定时任务
func (cs *CronService) InitLogsCronRun() {
	// 注册任务
	setting, err := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanCronKeyName)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("获取日志的cron失败，原因：%s", err))
	}
	ClearLogsTaskId = AddTask(ScheduledTask{
		Schedule: setting.Value,
		Job:      ClearLogs,
	})

	// 添加任务
	err = models.AddSendTaskWithID("日志定时清除", constant.CleanLogsTaskId, "admin")
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("添加日志定时清除任务失败，原因：%s", err))
	}
}

// 更新清除任务定时任务
func (cs *CronService) UpdateLogsCronRun(cron string) {
	RemoveTask(ClearLogsTaskId)
	ClearLogsTaskId = AddTask(ScheduledTask{
		Schedule: cron,
		Job:      ClearLogs,
	})
	logging.Logger.Error(fmt.Sprintf("更新日志的cron成功，%s", cron))
	logging.Logger.Error(fmt.Sprintf("所有的定时任务： %s", TaskList))

}

func Setup() {
	cs := CronService{}
	cs.InitLogsCronRun()
}
