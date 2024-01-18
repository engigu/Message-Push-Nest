package cron_service

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"reflect"
	"runtime"
	"strings"

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

func getFunctionName(f interface{}) string {
	ptr := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
	name := ptr[strings.LastIndex(ptr, ".")+1:]
	return name
}

// AddTask 添加定时任务
func AddTask(task ScheduledTask) cron.EntryID {
	mutex.Lock()
	defer mutex.Unlock()

	jobName := getFunctionName(task.Job)
	taskId, err := CronInstance.AddFunc(task.Schedule, task.Job)
	if err != nil {
		// 处理错误
		logrus.Errorf("注册定时任务失败, job: %s, 原因：%s", jobName, err)
	} else {
		TaskList[taskId] = &task
		logrus.Infof("注册定时任务成功, job: %s, entryID: %d, cron: %s", jobName, taskId, task.Schedule)
	}
	return taskId
}

// RemoveTask 删除定时任务
func RemoveTask(taskID cron.EntryID) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := TaskList[taskID]; ok {
		CronInstance.Remove(taskID)
		delete(TaskList, taskID)
	}
}
