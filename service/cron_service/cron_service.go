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
	var logOutput []string
	var errStr string
	status := 1

	sm := send_message_service.SendMessageService{TaskID: constant.CleanLogsTaskId}
	logging.Logger.Error("开始清除日志")
	logOutput = append(logOutput, "开始清除日志")

	setting, err := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanKeepKeyName)
	if err != nil {
		errStr = fmt.Sprintf("获取日志的保留数失败，原因：%s", err)
		logging.Logger.Error(errStr)
		sm.MarkStatus(errStr, &status)
		logOutput = append(logOutput, errStr)
	}

	keepNum := com.StrTo(setting.Value).MustInt()
	err = models.DeleteOutDateLogs(keepNum)
	if err != nil {
		errStr = fmt.Sprintf("删除日志失败，原因：%s", err)
		logging.Logger.Error(errStr)
		sm.MarkStatus(errStr, &status)
		logOutput = append(logOutput, errStr)
	} else {
		errStr = fmt.Sprintf("删除日志成功，保留数目：%d", keepNum)
		logging.Logger.Error(errStr)
		logOutput = append(logOutput, errStr)
	}
	sm.RecordSendLog(logOutput, status)
}

// 启动注册清除任务定时任务
func (cs *CronService) InitLogsCronRun() {
	setting, err := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanCronKeyName)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("获取日志的cron失败，原因：%s", err))
	}
	ClearLogsTaskId = AddTask(ScheduledTask{
		Schedule: setting.Value,
		Job:      ClearLogs,
	})
}

// 更新清除任务定时任务
func (cs *CronService) UpdateLogsCronRun(cron string) {
	RemoveTask(ClearLogsTaskId)
	ClearLogsTaskId = AddTask(ScheduledTask{
		Schedule: cron,
		Job:      ClearLogs,
	})
	logging.Logger.Error(fmt.Sprintf("更新日志的cron成功，%s", cron))
	logging.Logger.Info(fmt.Sprintf("所有的定时任务： %s", TaskList))

}

func Setup() {
	cs := CronService{}
	cs.InitLogsCronRun()
}
