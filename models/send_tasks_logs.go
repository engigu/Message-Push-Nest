package models

import (
	"fmt"
	"message-nest/pkg/util"
)

type SendTasksLogs struct {
	ID       int    `gorm:"primary_key" json:"id" `
	TaskID   string `json:"task_id" gorm:"type:varchar(12) comment '任务id';default:'';index:task_id"`
	Log      string `json:"log" gorm:"type:text comment '日志';"`
	Status   *int   `json:"status" gorm:"type:int comment '状态';default:0;"`
	CallerIp string `json:"caller_ip" gorm:"type:varchar(256) comment '发送者的ip';default:'';"`

	CreatedOn  util.Time `json:"created_on" gorm:"type:timestamp comment '创建时间';default:current_timestamp;"`
	ModifiedOn util.Time `json:"modified_on" gorm:"type:timestamp comment '更新时间';"`
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
	CallerIp   string    `json:"caller_ip"`
}

// GetSendLogs 获取所有日志记录
func GetSendLogs(pageNum int, pageSize int, name string, taskId string, maps map[string]interface{}) ([]LogsResult, error) {
	var logs []LogsResult
	logt := db.NewScope(SendTasksLogs{}).TableName()
	taskt := db.NewScope(SendTasks{}).TableName()

	query := db.
		Table(logt).
		Select(fmt.Sprintf("%s.*, %s.name as task_name", logt, taskt)).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.task_id = %s.id", taskt, logt, taskt))

	dayVal, ok := maps["day_created_on"]
	if ok {
		delete(maps, "day_created_on")
		query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
	}

	query = query.Where(maps)
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
func GetSendLogsTotal(name string, taskId string, maps map[string]interface{}) (int, error) {
	var total int
	logt := db.NewScope(SendTasksLogs{}).TableName()
	taskt := db.NewScope(SendTasks{}).TableName()
	query := db.
		Table(logt).
		Joins(fmt.Sprintf("LEFT JOIN %s ON %s.task_id = %s.id", taskt, logt, taskt))

	dayVal, ok := maps["day_created_on"]
	if ok {
		delete(maps, "day_created_on")
		query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
	}

	query = query.Where(maps)
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
func DeleteOutDateLogs(keepNum int) (int, error) {
	var affectedRows int
	logt := db.NewScope(SendTasksLogs{}).TableName()
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
		return affectedRows, result.Error
	}
	affectedRows = int(result.RowsAffected)
	return affectedRows, nil
}

type StatisticData struct {
	TodaySuccNum   int `json:"today_succ_num"`
	TodayFailedNum int `json:"today_failed_num"`
	TodayTotalNum  int `json:"today_total_num"`

	LatestSendData []LatestSendData `json:"latest_send_data"`
	WayCateData    []WayCateData    `json:"way_cate_data"`
}

type LatestSendData struct {
	Day          string `json:"day"`
	Num          int    `json:"num"`
	SuccNum      int    `json:"succ_num"`
	DaySuccNum   int    `json:"day_succ_num"`
	DayFailedNum int    `json:"day_failed_num"`
}

type WayCateData struct {
	WayName  string `json:"way_name"`
	CountNum int    `json:"count_num"`
}

// GetStatisticData 获取统计数据
func GetStatisticData() (StatisticData, error) {
	var statistic StatisticData
	var latestData []LatestSendData
	var wayCateData []WayCateData
	logt := db.NewScope(SendTasksLogs{}).TableName()
	inst := db.NewScope(SendTasksIns{}).TableName()
	wayst := db.NewScope(SendWays{}).TableName()
	currDay := util.GetNowTimeStr()[:10]

	// 今日统计数据
	query := db.
		Table(logt).
		Select(`
	COUNT(*) AS today_total_num,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS today_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS today_failed_num`).
		Where("DATE(created_on) = ?", currDay)

	query.First(&statistic)

	// 最近30天数据
	days := 30
	queryData := db.
		Table(logt).
		Select(`
	CAST(DATE(created_on) AS CHAR) AS day,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS day_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS day_failed_num,
	COUNT(*) AS num`).
		Where(" created_on >= DATE(?) - INTERVAL ? DAY", currDay, days).
		Group("day").
		Order("day")

	queryData.Scan(&latestData)

	// 消息实例分类数目
	db.
		Table(inst).
		Select(fmt.Sprintf("%s.name as way_name, count(%s.id) as count_num", wayst, wayst)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", wayst, inst, wayst)).
		Group(fmt.Sprintf("%s.id", wayst)).
		Scan(&wayCateData)

	statistic.LatestSendData = latestData
	statistic.WayCateData = wayCateData
	return statistic, nil
}
