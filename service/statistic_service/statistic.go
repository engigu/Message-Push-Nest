package statistic_service

import (
	"message-nest/models"
)

type StatisticService struct {
	TaskID string
	Days   int
}

func (sw *StatisticService) GetStatisticData() (models.StatisticData, error) {
	return models.GetStatisticData()
}

// GetBasicStatisticData 获取基础统计数据
func (sw *StatisticService) GetBasicStatisticData() (models.BasicStatisticData, error) {
	return models.GetBasicStatisticData()
}

// GetTrendStatisticData 获取趋势统计数据
func (sw *StatisticService) GetTrendStatisticData() (models.TrendStatisticData, error) {
	days := sw.Days
	if days <= 0 {
		days = 30 // 默认30天
	}
	return models.GetTrendStatisticData(days)
}

// GetChannelStatisticData 获取渠道统计数据
func (sw *StatisticService) GetChannelStatisticData() (models.ChannelStatisticData, error) {
	return models.GetChannelStatisticData()
}

// GetSendStatsData 获取发送统计数据（基于 send_stats 表）
func (sw *StatisticService) GetSendStatsData() (models.SendStatsData, error) {
	days := sw.Days
	if days <= 0 {
		days = 30 // 默认30天
	}
	return models.GetSendStatsData(days)
}

// GetSendStatsByTask 获取指定任务的发送统计数据
func (sw *StatisticService) GetSendStatsByTask() (models.SendStatsData, error) {
	days := sw.Days
	if days <= 0 {
		days = 30 // 默认30天
	}

	return models.GetSendStatsByTask(sw.TaskID, days)
}
