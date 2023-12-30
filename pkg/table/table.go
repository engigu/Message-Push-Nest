package table

import (
	"fmt"
	"message-nest/pkg/setting"
)

var InsTableName string
var WayTableName string
var LogsTableName string
var TasksTableName string

func Setup() {
	InsTableName = fmt.Sprintf("%ssend_tasks_ins", setting.DatabaseSetting.TablePrefix)
	WayTableName = fmt.Sprintf("%ssend_ways", setting.DatabaseSetting.TablePrefix)
	LogsTableName = fmt.Sprintf("%ssend_tasks_logs", setting.DatabaseSetting.TablePrefix)
	TasksTableName = fmt.Sprintf("%ssend_tasks", setting.DatabaseSetting.TablePrefix)
}
