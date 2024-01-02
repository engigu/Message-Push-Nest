package cron_service

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"message-nest/pkg/logging"
	"sync"
)

type ScheduledTask struct {
	//ID       int
	Schedule string
	Job      func()
}

var (
	CronInstance *cron.Cron
	TaskList     map[cron.EntryID]*ScheduledTask
	mutex        sync.Mutex
)

func init() {
	CronInstance = cron.New()
	CronInstance.Start()
	TaskList = make(map[cron.EntryID]*ScheduledTask)
}

// 添加定时任务
func AddTask(task ScheduledTask) cron.EntryID {
	mutex.Lock()
	defer mutex.Unlock()

	taskId, err := CronInstance.AddFunc(task.Schedule, task.Job)
	if err != nil {
		// 处理错误
		logging.Logger.Error(fmt.Sprintf("添加定时任务失败，原因：%s", err))
	} else {
		TaskList[taskId] = &task
		logging.Logger.Error(fmt.Sprintf("添加定时任务成功，entryID: %d, cron: %s", taskId, task.Schedule))
	}
	return taskId
}

// 删除定时任务
func RemoveTask(taskID cron.EntryID) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := TaskList[taskID]; ok {
		CronInstance.Remove(taskID)
		delete(TaskList, taskID)
	}
}
