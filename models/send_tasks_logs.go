package models

import (
	"fmt"
	"message-nest/pkg/util"
	//"time"
)

type SendTasksLogs struct {
	ID       int    `gorm:"primaryKey" json:"id" `
	TaskID   string `json:"task_id" gorm:"type:varchar(12) ;default:'';index:task_id"`
	Type     string `json:"type" gorm:"type:varchar(20) ;default:'task';comment:'类型：task-任务，template-模板'"`
	Name     string `json:"name" gorm:"type:varchar(256) ;default:'';comment:'任务或模板名称'"`
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
	Type       string    `json:"type"`
	Name       string    `json:"name"`
	Log        string    `json:"log"`
	CreatedOn  util.Time `json:"created_on"`
	ModifiedOn util.Time `json:"modified_on"`
	Status     int       `json:"status"`
	CallerIp   string    `json:"caller_ip"`
}

// GetSendLogs 获取所有日志记录
func GetSendLogs(pageNum int, pageSize int, name string, taskId string, maps map[string]interface{}) ([]LogsResult, error) {
	var logs []LogsResult
	logt := GetSchema(SendTasksLogs{})

	// 简化查询，只查询日志表
	query := db.Table(logt)

	dayVal, ok := maps["day_created_on"]
	if ok {
		delete(maps, "day_created_on")
		query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
	}

	query = query.Where(maps)

	// 按名称搜索（搜索日志表的 name 字段）
	if name != "" {
		query = query.Where(fmt.Sprintf("%s.name like ?", logt), fmt.Sprintf("%%%s%%", name))
	}
	if taskId != "" {
		query = query.Where(fmt.Sprintf("%s.task_id = ?", logt), taskId)
	}
	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	query.Scan(&logs)

	//v1 接口的历史日志数据兼容处理
	// 应用层处理：为历史数据（type=task 且 name 为空）补充任务名称
	fillTaskNamesForLogs(&logs)

	return logs, nil
}

// fillTaskNamesForLogs 为历史日志数据补充任务名称
func fillTaskNamesForLogs(logs *[]LogsResult) {
	if logs == nil || len(*logs) == 0 {
		return
	}

	// 收集需要查询的 task_id
	taskIdsMap := make(map[string]bool)
	for _, log := range *logs {
		// 只处理 type=task 且 name 为空的记录
		if (log.Type == "" || log.Type == "task") && log.Name == "" && log.TaskID != "" {
			taskIdsMap[log.TaskID] = true
		}
	}

	// 如果没有需要查询的任务，直接返回
	if len(taskIdsMap) == 0 {
		return
	}

	// 批量查询任务名称
	taskIds := make([]string, 0, len(taskIdsMap))
	for taskId := range taskIdsMap {
		taskIds = append(taskIds, taskId)
	}

	var tasks []SendTasks
	taskt := GetSchema(SendTasks{})
	db.Table(taskt).
		Select("id, name").
		Where("id IN ?", taskIds).
		Scan(&tasks)

	// 构建 taskId -> name 的映射
	taskNameMap := make(map[string]string)
	for _, task := range tasks {
		taskNameMap[task.ID] = task.Name
	}

	// 填充日志的 name 字段
	for i := range *logs {
		log := &(*logs)[i]
		if (log.Type == "" || log.Type == "task") && log.Name == "" && log.TaskID != "" {
			if taskName, exists := taskNameMap[log.TaskID]; exists {
				log.Name = taskName
			}
		}
	}
}

// GetSendLogsTotal 获取所有日志总数
func GetSendLogsTotal(name string, taskId string, maps map[string]interface{}) (int64, error) {
	var total int64
	logt := GetSchema(SendTasksLogs{})

	// 简化查询，只查询日志表
	query := db.Table(logt)

	dayVal, ok := maps["day_created_on"]
	if ok {
		delete(maps, "day_created_on")
		query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
	}

	query = query.Where(maps)

	// 按名称搜索（搜索日志表的 name 字段）
	if name != "" {
		query = query.Where(fmt.Sprintf("%s.name like ?", logt), fmt.Sprintf("%%%s%%", name))
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

	// 优化方案：使用GORM的Offset和Limit找到临界ID，兼容多种数据库
	// 1. 获取第 keepNum 条记录的ID作为临界值
	var threshold SendTasksLogs
	result := db.Model(&SendTasksLogs{}).
		Select("id").
		Order("created_on DESC").
		Offset(keepNum - 1).
		Limit(1).
		First(&threshold)

	// 如果记录总数不足keepNum条，则不需要删除
	if result.Error != nil {
		return 0, nil
	}

	// 2. 删除ID小于临界值的记录
	deleteResult := db.Where("id < ?", threshold.ID).Delete(&SendTasksLogs{})
	if deleteResult.Error != nil {
		return affectedRows, deleteResult.Error
	}

	affectedRows = int(deleteResult.RowsAffected)
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

// BasicStatisticData 基础统计数据
type BasicStatisticData struct {
	TodaySuccNum          int `json:"today_succ_num"`
	TodayFailedNum        int `json:"today_failed_num"`
	TodayTotalNum         int `json:"today_total_num"`
	MessageTotalNum       int `json:"message_total_num"`
	HostedMessageTotalNum int `json:"hosted_message_total_num"`
}

// TrendStatisticData 趋势统计数据
type TrendStatisticData struct {
	LatestSendData []LatestSendData `json:"latest_send_data"`
}

// ChannelStatisticData 渠道统计数据
type ChannelStatisticData struct {
	WayCateData []WayCateData `json:"way_cate_data"`
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

// GetBasicStatisticData 获取基础统计数据
func GetBasicStatisticData() (BasicStatisticData, error) {
	var statistic BasicStatisticData
	logt := GetSchema(SendTasksLogs{})
	hostedt := GetSchema(HostedMessage{})
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

	// 全部消息统计数据
	totalQuery := db.Table(logt).Select(`COUNT(*) AS message_total_num`)
	totalQuery.Take(&statistic)

	// 托管消息统计数据
	hostedMessageTotalQuery := db.Table(hostedt).Select(`COUNT(*) AS hosted_message_total_num`)
	hostedMessageTotalQuery.Take(&statistic)

	return statistic, nil
}

// GetTrendStatisticData 获取趋势统计数据（使用 send_stats 表）
func GetTrendStatisticData() (TrendStatisticData, error) {
	var statistic TrendStatisticData
	var latestData []LatestSendData
	statsTable := GetSchema(SendStats{})

	// 最近30天数据
	days := 30
	now := util.GetNowTime()
	past := now.AddDate(0, 0, -days)
	pastDate := past.Format("2006-01-02")

	queryData := db.
		Table(statsTable).
		Select(`
			day,
			SUM(CASE WHEN status = 'success' THEN num ELSE 0 END) AS day_succ_num,
			SUM(CASE WHEN status = 'failed' THEN num ELSE 0 END) AS day_failed_num,
			SUM(num) AS num
		`).
		Where("day >= ?", pastDate).
		Group("day").
		Order("day")

	queryData.Scan(&latestData)
	statistic.LatestSendData = latestData
	return statistic, nil
}

// GetChannelStatisticData 获取渠道统计数据（基于 send_stats 表统计任务，再查询渠道）
func GetChannelStatisticData() (ChannelStatisticData, error) {
	var statistic ChannelStatisticData
	var wayCateData []WayCateData
	statsTable := GetSchema(SendStats{})
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})

	// 第一步：从 send_stats 表统计每个任务的执行次数
	var taskStats []struct {
		TaskID string
		Num    int64
	}

	db.
		Table(statsTable).
		Select("task_id, SUM(num) AS num").
		Where("task_id != ''"). // 排除全局统计
		Group("task_id").
		Scan(&taskStats)

	if len(taskStats) == 0 {
		statistic.WayCateData = wayCateData
		return statistic, nil
	}

	// 第二步：收集所有任务ID
	taskIDs := make([]string, 0, len(taskStats))
	taskNumMap := make(map[string]int64)
	for _, stat := range taskStats {
		taskIDs = append(taskIDs, stat.TaskID)
		taskNumMap[stat.TaskID] = stat.Num
	}

	// 第三步：查询这些任务关联的渠道信息
	var taskWays []struct {
		TaskID  string
		WayID   string
		WayName string
	}

	db.
		Table(insTable).
		Select(fmt.Sprintf("%s.task_id, %s.way_id, %s.name as way_name", insTable, insTable, waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Where(fmt.Sprintf("%s.task_id IN ?", insTable), taskIDs).
		Group(fmt.Sprintf("%s.task_id, %s.way_id, %s.name", insTable, insTable, waysTable)).
		Scan(&taskWays)

	// 第四步：按渠道聚合统计
	wayCountMap := make(map[string]int64)
	for _, tw := range taskWays {
		if num, exists := taskNumMap[tw.TaskID]; exists {
			wayCountMap[tw.WayName] += num
		}
	}

	// 第五步：转换为返回格式
	for wayName, count := range wayCountMap {
		wayCateData = append(wayCateData, WayCateData{
			WayName:  wayName,
			CountNum: int(count),
		})
	}

	statistic.WayCateData = wayCateData
	return statistic, nil
}
