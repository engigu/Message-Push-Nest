package models

import (
	"fmt"
	"message-nest/pkg/util"
)

// SendStats 发送统计表
type SendStats struct {
	ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID   *uint  `json:"task_id" gorm:"type:bigint unsigned;index:idx_task_type_day_status"`
	TaskType string `json:"task_type" gorm:"type:varchar(20);default:'task';index:idx_task_type_day_status;comment:'任务类型：task-发信任务，template-模板任务'"`
	Day      string `json:"day" gorm:"type:varchar(10);index:idx_task_type_day_status;index:idx_day_status"`
	Status   string `json:"status" gorm:"type:varchar(20);index:idx_task_type_day_status;index:idx_day_status"`
	Num      int64  `json:"num" gorm:"type:bigint;default:0"`
}

// SendStatsData 首页统计数据结构
type SendStatsData struct {
	TodaySuccNum   int64               `json:"today_succ_num"`
	TodayFailedNum int64               `json:"today_failed_num"`
	TodayTotalNum  int64               `json:"today_total_num"`
	TotalSuccNum   int64               `json:"total_succ_num"`
	TotalFailedNum int64               `json:"total_failed_num"`
	TotalNum       int64               `json:"total_num"`
	DailyStats     []DailyStatsData    `json:"daily_stats"`
	StatusStats    []StatusStatsData   `json:"status_stats"`
	TaskTypeStats  []TaskTypeStatsData `json:"task_type_stats"`
}

// DailyStatsData 每日统计数据
type DailyStatsData struct {
	Day       string `json:"day"`
	SuccNum   int64  `json:"succ_num"`
	FailedNum int64  `json:"failed_num"`
	TotalNum  int64  `json:"total_num"`
}

// StatusStatsData 状态统计数据
type StatusStatsData struct {
	Status string `json:"status"`
	Num    int64  `json:"num"`
}

// TaskTypeStatsData 任务类型统计数据
type TaskTypeStatsData struct {
	TaskType string `json:"task_type"`
	Num      int64  `json:"num"`
}

// GetSendStatsData 获取发送统计数据
func GetSendStatsData(days int) (SendStatsData, error) {
	var result SendStatsData
	statsTable := GetSchema(SendStats{})
	currDay := util.GetNowTimeStr()[:10]

	// 今日统计
	todayQuery := db.Table(statsTable).
		Select(`
			SUM(CASE WHEN status = 'success' THEN num ELSE 0 END) AS today_succ_num,
			SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END) AS today_failed_num,
			SUM(num) AS today_total_num
		`).
		Where("day = ?", currDay)

	todayQuery.Scan(&result)

	// 总计统计
	totalQuery := db.Table(statsTable).
		Select(`
			SUM(CASE WHEN status = 'success' THEN num ELSE 0 END) AS total_succ_num,
			SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END) AS total_failed_num,
			SUM(num) AS total_num
		`)

	totalQuery.Scan(&result)

	// 最近N天每日统计
	if days > 0 {
		now := util.GetNowTime()
		past := now.AddDate(0, 0, -days)
		pastDate := past.Format("2006-01-02")

		var dailyStats []DailyStatsData
		dailyQuery := db.Table(statsTable).
			Select(`
				day,
				SUM(CASE WHEN status = 'success' THEN num ELSE 0 END) AS succ_num,
				SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END) AS failed_num,
				SUM(num) AS total_num
			`).
			Where("day >= ?", pastDate).
			Group("day").
			Order("day DESC")

		dailyQuery.Scan(&dailyStats)
		result.DailyStats = dailyStats
	}

	// 状态分布统计
	var statusStats []StatusStatsData
	statusQuery := db.Table(statsTable).
		Select("status, SUM(num) AS num").
		Group("status").
		Order("num DESC")

	statusQuery.Scan(&statusStats)
	result.StatusStats = statusStats

	// 任务类型分布统计
	var taskTypeStats []TaskTypeStatsData
	taskTypeQuery := db.Table(statsTable).
		Select("task_type, SUM(num) AS num").
		Group("task_type").
		Order("num DESC")

	taskTypeQuery.Scan(&taskTypeStats)
	result.TaskTypeStats = taskTypeStats

	return result, nil
}

// GetSendStatsByTask 获取指定任务的统计数据
func GetSendStatsByTask(taskID uint, days int) (SendStatsData, error) {
	var result SendStatsData
	statsTable := GetSchema(SendStats{})
	currDay := util.GetNowTimeStr()[:10]

	// 今日统计
	todayQuery := db.Table(statsTable).
		Select(`
			SUM(CASE WHEN status = 'success' THEN num ELSE 0 END) AS today_succ_num,
			SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END) AS today_failed_num,
			SUM(num) AS today_total_num
		`).
		Where("day = ? AND task_id = ?", currDay, taskID)

	todayQuery.Scan(&result)

	// 总计统计
	totalQuery := db.Table(statsTable).
		Select(`
			SUM(CASE WHEN status = 'success' THEN num ELSE 0 END) AS total_succ_num,
			SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END) AS total_failed_num,
			SUM(num) AS total_num
		`).
		Where("task_id = ?", taskID)

	totalQuery.Scan(&result)

	// 最近N天每日统计
	if days > 0 {
		now := util.GetNowTime()
		past := now.AddDate(0, 0, -days)
		pastDate := past.Format("2006-01-02")

		var dailyStats []DailyStatsData
		dailyQuery := db.Table(statsTable).
			Select(`
				day,
				SUM(CASE WHEN status = 'success' THEN num ELSE 0 END) AS succ_num,
				SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END) AS failed_num,
				SUM(num) AS total_num
			`).
			Where("day >= ? AND task_id = ?", pastDate, taskID).
			Group("day").
			Order("day DESC")

		dailyQuery.Scan(&dailyStats)
		result.DailyStats = dailyStats
	}

	// 状态分布统计
	var statusStats []StatusStatsData
	statusQuery := db.Table(statsTable).
		Select("status, SUM(num) AS num").
		Where("task_id = ?", taskID).
		Group("status").
		Order("num DESC")

	statusQuery.Scan(&statusStats)
	result.StatusStats = statusStats

	return result, nil
}

// IncrementSendStats 增加发送统计（用于实时更新统计数据）
func IncrementSendStats(taskID *uint, taskType string, day string, status string, num int64) error {
	statsTable := GetSchema(SendStats{})

	// 使用 ON DUPLICATE KEY UPDATE 或 UPSERT 逻辑
	query := fmt.Sprintf(`
		INSERT INTO %s (task_id, task_type, day, status, num)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE num = num + ?
	`, statsTable)

	return db.Exec(query, taskID, taskType, day, status, num, num).Error
}
