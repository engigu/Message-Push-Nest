package models

import (
	"fmt"
	"message-nest/pkg/util"
	//"time"
)

type SendTasksLogs struct {
	ID       int    `gorm:"primaryKey" json:"id" `
	TaskID   string `json:"task_id" gorm:"type:varchar(12);default:'';index:task_id;index:idx_logs_task_time,priority:1"`
	Type     string `json:"type" gorm:"type:varchar(20);default:'task';comment:'类型：task-任务，template-模板'"`
	Name     string `json:"name" gorm:"type:varchar(256);default:'';comment:'任务或模板名称'"`
	Log      string `json:"log" gorm:"type:text;"`
	Status   *int   `json:"status" gorm:"type:int;default:0;"`
	CallerIp string `json:"caller_ip" gorm:"type:varchar(256);default:'';"`

	CreatedBy  string `json:"created_by" gorm:"type:varchar(36);default:'';index;index:idx_logs_user_time,priority:1"`
	ModifiedBy string `json:"modified_by" gorm:"type:varchar(36);default:'';"`

	CreatedAt util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime;index:idx_logs_created_on;index:idx_logs_user_time,priority:2;index:idx_logs_task_time,priority:2"`
	UpdatedAt util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime;"`
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
	CreatedBy  string    `json:"created_by"`
	ModifiedBy string    `json:"modified_by"`
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
		if dayStr, isStr := dayVal.(string); isStr && len(dayStr) >= 10 {
			dayPrefix := dayStr[:10]
			query = query.Where(fmt.Sprintf("%s.created_on >= ? AND %s.created_on <= ?", logt, logt),
				dayPrefix+" 00:00:00", dayPrefix+" 23:59:59.999")
		} else {
			query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
		}
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
		if dayStr, isStr := dayVal.(string); isStr && len(dayStr) >= 10 {
			dayPrefix := dayStr[:10]
			query = query.Where(fmt.Sprintf("%s.created_on >= ? AND %s.created_on <= ?", logt, logt),
				dayPrefix+" 00:00:00", dayPrefix+" 23:59:59.999")
		} else {
			query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
		}
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

// DeleteOutDateLogs 删除过期的任务日志，保留最新的 keepNum 条，支持按用户过滤
func DeleteOutDateLogs(keepNum int, createdBy ...string) (int, error) {
	var affectedRows int
	if keepNum <= 0 {
		return 0, nil
	}

	// 优化方案：使用GORM的Offset和Limit找到临界ID，兼容多种数据库
	// 1. 获取第 keepNum 条记录的ID作为临界值
	var threshold SendTasksLogs
	query := db.Model(&SendTasksLogs{}).Select("id")
	if len(createdBy) > 0 && createdBy[0] != "" {
		query = query.Where("created_by = ?", createdBy[0])
	} else if len(createdBy) > 0 && createdBy[0] == "" {
		query = query.Where("created_by = '' OR created_by IS NULL")
	}

	result := query.Order("created_on DESC, id DESC").
		Offset(keepNum - 1).
		Limit(1).
		Find(&threshold)

	// 如果记录总数不足keepNum条，则不需要删除
	if result.Error != nil || threshold.ID == 0 {
		return 0, nil
	}

	// 2. 删除ID小于临界值的记录（限定用户范围）
	deleteQuery := db.Where("id < ?", threshold.ID)
	if len(createdBy) > 0 && createdBy[0] != "" {
		deleteQuery = deleteQuery.Where("created_by = ?", createdBy[0])
	} else if len(createdBy) > 0 && createdBy[0] == "" {
		deleteQuery = deleteQuery.Where("created_by = '' OR created_by IS NULL")
	}

	deleteResult := deleteQuery.Delete(&SendTasksLogs{})
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
	WaysTotalNum          int `json:"ways_total_num"`

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
	WaysTotalNum          int `json:"ways_total_num"`
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
func GetStatisticData(createdBy ...string) (StatisticData, error) {
	var statistic StatisticData
	var latestData []LatestSendData
	var wayCateData []WayCateData
	logt := GetSchema(SendTasksLogs{})
	inst := GetSchema(SendTasksIns{})
	hostedt := GetSchema(HostedMessage{})
	wayst := GetSchema(SendWays{})
	currDay := util.GetNowTimeStr()[:10]

	userFilter := ""
	if len(createdBy) > 0 && createdBy[0] != "" {
		userFilter = createdBy[0]
	}

	// 今日统计数据
	var todayStats struct {
		TodayTotalNum  int `gorm:"column:today_total_num"`
		TodaySuccNum   int `gorm:"column:today_succ_num"`
		TodayFailedNum int `gorm:"column:today_failed_num"`
	}
	startOfToday := currDay + " 00:00:00"
	endOfToday := currDay + " 23:59:59.999"
	query := db.
		Table(logt).
		Select(`
	COUNT(*) AS today_total_num,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS today_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS today_failed_num`).
		Where("created_on >= ? AND created_on <= ?", startOfToday, endOfToday)
	if userFilter != "" {
		query = query.Where("created_by = ?", userFilter)
	}

	if err := query.Scan(&todayStats).Error; err == nil {
		statistic.TodayTotalNum = todayStats.TodayTotalNum
		statistic.TodaySuccNum = todayStats.TodaySuccNum
		statistic.TodayFailedNum = todayStats.TodayFailedNum
	}

	// 	全部消息统计数据
	var messageTotalNum int
	msgQuery := db.Table(logt).Select(`COUNT(*)`)
	if userFilter != "" {
		msgQuery = msgQuery.Where("created_by = ?", userFilter)
	}
	if err := msgQuery.Scan(&messageTotalNum).Error; err == nil {
		statistic.MessageTotalNum = messageTotalNum
	}

	// 	托管消息统计数据
	var hostedMessageTotalNum int
	hostedQuery := db.Table(hostedt).Select(`COUNT(*)`)
	if userFilter != "" {
		hostedQuery = hostedQuery.Where("created_by = ?", userFilter)
	}
	if err := hostedQuery.Scan(&hostedMessageTotalNum).Error; err == nil {
		statistic.HostedMessageTotalNum = hostedMessageTotalNum
	}

	// 发送渠道统计数据
	var waysTotalNum int
	waysQuery := db.Table(wayst).Select(`COUNT(*)`)
	if userFilter != "" {
		waysQuery = waysQuery.Where("created_by = ?", userFilter)
	}
	if err := waysQuery.Scan(&waysTotalNum).Error; err == nil {
		statistic.WaysTotalNum = waysTotalNum
	}

	// 最近30天数据
	days := 30
	now := util.GetNowTime()
	past := now.AddDate(0, 0, -days)
	pastDate := past.Format("2006-01-02")
	next := now.AddDate(0, 0, 1)
	nextDate := next.Format("2006-01-02")
	var dayExpr string
	if db.Dialector.Name() == "postgres" {
		dayExpr = "CAST(DATE(created_on) AS VARCHAR)"
	} else {
		dayExpr = "CAST(DATE(created_on) AS CHAR)"
	}
	queryData := db.
		Table(logt).
		Select(fmt.Sprintf(`
	%s AS day,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS day_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS day_failed_num,
	COUNT(*) AS num`, dayExpr)).
		Where(fmt.Sprintf(" created_on >= '%s' and created_on <= '%s' ", pastDate, nextDate))
	if userFilter != "" {
		queryData = queryData.Where("created_by = ?", userFilter)
	}
	queryData = queryData.Group("day").Order("day")

	queryData.Scan(&latestData)

	// 消息实例分类数目
	wayCateQuery := db.
		Table(inst).
		Select(fmt.Sprintf("%s.name as way_name, count(%s.id) as count_num", wayst, wayst)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", wayst, inst, wayst))
	if userFilter != "" {
		wayCateQuery = wayCateQuery.Where(fmt.Sprintf("%s.created_by = ?", inst), userFilter)
	}
	wayCateQuery.Group(fmt.Sprintf("%s.id", wayst)).
		Scan(&wayCateData)

	statistic.LatestSendData = latestData
	statistic.WayCateData = wayCateData
	return statistic, nil
}

// GetBasicStatisticData 获取基础统计数据
func GetBasicStatisticData(createdBy ...string) (BasicStatisticData, error) {
	var statistic BasicStatisticData
	logt := GetSchema(SendTasksLogs{})
	hostedt := GetSchema(HostedMessage{})
	currDay := util.GetNowTimeStr()[:10]

	userFilter := ""
	if len(createdBy) > 0 && createdBy[0] != "" {
		userFilter = createdBy[0]
	}

	// 今日统计数据
	var todayStats struct {
		TodayTotalNum  int `gorm:"column:today_total_num"`
		TodaySuccNum   int `gorm:"column:today_succ_num"`
		TodayFailedNum int `gorm:"column:today_failed_num"`
	}
	startOfToday := currDay + " 00:00:00"
	endOfToday := currDay + " 23:59:59.999"
	query := db.
		Table(logt).
		Select(`
	COUNT(*) AS today_total_num,
	SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) AS today_succ_num,
	SUM(CASE WHEN status != 1 or status is null THEN 1 ELSE 0 END) AS today_failed_num`).
		Where("created_on >= ? AND created_on <= ?", startOfToday, endOfToday)
	if userFilter != "" {
		query = query.Where("created_by = ?", userFilter)
	}

	if err := query.Scan(&todayStats).Error; err == nil {
		statistic.TodayTotalNum = todayStats.TodayTotalNum
		statistic.TodaySuccNum = todayStats.TodaySuccNum
		statistic.TodayFailedNum = todayStats.TodayFailedNum
	}

	// 全部消息统计数据
	var messageTotalNum int
	msgQuery := db.Table(logt).Select(`COUNT(*)`)
	if userFilter != "" {
		msgQuery = msgQuery.Where("created_by = ?", userFilter)
	}
	if err := msgQuery.Scan(&messageTotalNum).Error; err == nil {
		statistic.MessageTotalNum = messageTotalNum
	}

	// 托管消息统计数据
	var hostedMessageTotalNum int
	hostedQuery := db.Table(hostedt).Select(`COUNT(*)`)
	if userFilter != "" {
		hostedQuery = hostedQuery.Where("created_by = ?", userFilter)
	}
	if err := hostedQuery.Scan(&hostedMessageTotalNum).Error; err == nil {
		statistic.HostedMessageTotalNum = hostedMessageTotalNum
	}

	// 发送渠道统计数据
	var waysTotalNum int
	wayst := GetSchema(SendWays{})
	waysQuery := db.Table(wayst).Select(`COUNT(*)`)
	if userFilter != "" {
		waysQuery = waysQuery.Where("created_by = ?", userFilter)
	}
	if err := waysQuery.Scan(&waysTotalNum).Error; err == nil {
		statistic.WaysTotalNum = waysTotalNum
	}

	return statistic, nil
}

// GetTrendStatisticData 获取趋势统计数据（使用 send_stats 表）
func GetTrendStatisticData(days int, createdBy ...string) (TrendStatisticData, error) {
	var statistic TrendStatisticData
	var latestData []LatestSendData
	statsTable := GetSchema(SendStats{})

	// 默认30天，如果传入参数则使用参数值
	if days <= 0 {
		days = 30
	}

	userFilter := ""
	if len(createdBy) > 0 && createdBy[0] != "" {
		userFilter = createdBy[0]
	}

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
		Where("day >= ?", pastDate)
	if userFilter != "" {
		queryData = queryData.Where("created_by = ?", userFilter)
	}
	queryData = queryData.Group("day").Order("day")

	queryData.Scan(&latestData)
	statistic.LatestSendData = latestData
	return statistic, nil
}

// GetChannelStatisticData 获取渠道统计数据（基于 send_stats 表统计任务，再查询渠道）
func GetChannelStatisticData(createdBy ...string) (ChannelStatisticData, error) {
	var statistic ChannelStatisticData
	var wayCateData []WayCateData
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})

	userFilter := ""
	if len(createdBy) > 0 && createdBy[0] != "" {
		userFilter = createdBy[0]
	}

	// 统计所有任务和模板中关联的实例的渠道数据
	var wayStats []struct {
		WayName  string
		CountNum int
	}

	query := db.
		Table(insTable).
		Select(fmt.Sprintf("%s.name as way_name, COUNT(*) as count_num", waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable))
	if userFilter != "" {
		query = query.Where(fmt.Sprintf("%s.created_by = ?", insTable), userFilter)
	}
	err := query.Group(fmt.Sprintf("%s.name", waysTable)).
		Scan(&wayStats).Error

	if err != nil {
		return statistic, err
	}

	// 转换为返回格式
	for _, stat := range wayStats {
		wayCateData = append(wayCateData, WayCateData{
			WayName:  stat.WayName,
			CountNum: stat.CountNum,
		})
	}

	// 确保返回空数组而不是 nil
	if len(wayCateData) == 0 {
		wayCateData = []WayCateData{}
	}

	statistic.WayCateData = wayCateData
	return statistic, nil
}
