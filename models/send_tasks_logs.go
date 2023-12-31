package models

import (
	"fmt"
	"message-nest/pkg/table"
	"message-nest/pkg/util"
)

type SendTasksLogs struct {
	ID     int    `gorm:"primary_key" json:"id"`
	TaskID string `json:"task_id"`
	Log    string `json:"log"`
	Status int    `json:"status"`

	CreatedOn  util.Time `json:"created_on"`
	ModifiedOn util.Time `json:"modified_on"`
}

// Add 添加日志记录
func (log *SendTasksLogs) Add() error {
	if err := db.Create(&log).Error; err != nil {
		return err
	}
	return nil
}

// 日志列表的结果
type LogsResult struct {
	ID         int       `json:"id"`
	TaskID     string    `json:"task_id"`
	Log        string    `json:"log"`
	CreatedOn  util.Time `json:"created_on"`
	ModifiedOn util.Time `json:"modified_on"`
	TaskName   string    `json:"task_name"`
	Status     int       `json:"status"`
}

// GetSendLogs 获取所有日志记录
func GetSendLogs(pageNum int, pageSize int, name string, taskId string, maps interface{}) ([]LogsResult, error) {
	var logs []LogsResult
	logt := table.LogsTableName
	taskt := table.TasksTableName

	query := db.
		Table(logt).
		Select(fmt.Sprintf("%s.*, %s.name as task_name", logt, taskt)).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.task_id = %s.id", taskt, logt, taskt))

	if name != "" {
		query = query.Where(fmt.Sprintf("%s.name like ?", taskt), fmt.Sprintf("%%%s%%", name))
	}
	if taskId != "" {
		query = query.Where(fmt.Sprintf("%s.task_id = ?", logt), taskId)
	}
	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	query.Scan(&logs)

	return logs, nil
}

// GetSendLogsTotal 获取所有日志总数
func GetSendLogsTotal(name string, taskId string, maps interface{}) (int, error) {
	var total int
	logt := table.LogsTableName
	taskt := table.TasksTableName
	query := db.
		Table(logt).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.task_id = %s.id", taskt, logt, taskt))
	if name != "" {
		query = query.Where(fmt.Sprintf("%s.name like ?", taskt), fmt.Sprintf("%%%s%%", name))
	}
	if taskId != "" {
		query = query.Where(fmt.Sprintf("%s.task_id = ?", logt), taskId)
	}
	query.Count(&total)
	return total, nil
}

// GetSendLogsTotal 获取所有日志总数
func DeleteOutDateLogs(keepNum int) error {
	logt := table.LogsTableName
	sql := fmt.Sprintf(`DELETE FROM %s
WHERE id NOT IN (
    SELECT id FROM (
        SELECT id
        FROM %s
        ORDER BY created_on DESC
        LIMIT %d
    ) tmp
);`, logt, logt, keepNum)

	result := db.Exec(sql)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
