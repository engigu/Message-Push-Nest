package models

import (
	"fmt"
	"message-nest/pkg/util"
	//"time"
)

type SendTasksLogs struct {
	ID       int    `gorm:"primaryKey" json:"id" `
	TaskID   string `json:"task_id" gorm:"type:varchar(12) ;default:'';index:task_id"`
	Log      string `json:"log" gorm:"type:text ;"`
	Status   *int   `json:"status" gorm:"type:int ;default:0;"`
	CallerIp string `json:"caller_ip" gorm:"type:varchar(256) ;default:'';"`

	CreatedAt util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime "`
	UpdatedAt util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime ;"`
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
	logt := GetSchema(SendTasksLogs{})
	taskt := GetSchema(SendTasks{})

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
func GetSendLogsTotal(name string, taskId string, maps map[string]interface{}) (int64, error) {
	var total int64
	logt := GetSchema(SendTasksLogs{})
	taskt := GetSchema(SendTasks{})
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
	logt := GetSchema(SendTasksLogs{})
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
	TodaySuccNum          int `json:"today_succ_num"`
	TodayFailedNum        int `json:"today_failed_num"`
	TodayTotalNum         int `json:"today_total_num"`
	MessageTotalNum       int `json:"message_total_num"`
	HostedMessageTotalNum int `json:"hosted_message_total_num"`

	LatestSendData []LatestSendData `json:"latest_send_data" gorm:"many2many:latest_send_data;"`
	WayCateData    []WayCateData    `json:"way_cate_data" gorm:"many2many:way_cate_data;"`
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
	logt := GetSchema(SendTasksLogs{})
	inst := GetSchema(SendTasksIns{})
	hostedt := GetSchema(HostedMessage{})
	wayst := GetSchema(SendWays{})
	currDay := util.GetNowTimeStr()[:10]

	// 今日统计数据
	query := db.
		Table(logt).
		Select(`
	COUNT(*) AS today_total_num,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS today_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS today_failed_num`).
		Where("DATE(created_on) = ?", currDay)

	query.Take(&statistic)

	// 	全部消息统计数据
	totalQuery := db.Table(logt).Select(`COUNT(*) AS message_total_num`)
	totalQuery.Take(&statistic)

	// 	托管消息统计数据
	hostedMessageTotalQuery := db.Table(hostedt).Select(`COUNT(*) AS hosted_message_total_num`)
	hostedMessageTotalQuery.Take(&statistic)

	// 最近30天数据
	days := 30
	now := util.GetNowTime()
	past := now.AddDate(0, 0, -days)
	pastDate := past.Format("2006-01-02")
	next := now.AddDate(0, 0, 1)
	nextDate := next.Format("2006-01-02")
	queryData := db.
		Table(logt).
		Select(`
	CAST(DATE(created_on) AS CHAR) AS day,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS day_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS day_failed_num,
	COUNT(*) AS num`).
		Where(fmt.Sprintf(" created_on >= '%s' and created_on <= '%s' ", pastDate, nextDate)).
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
