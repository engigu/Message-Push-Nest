package statistic_service

import (
	"message-nest/models"
)

type StatisticService struct {
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
	return models.GetTrendStatisticData()
}

// GetChannelStatisticData 获取渠道统计数据
func (sw *StatisticService) GetChannelStatisticData() (models.ChannelStatisticData, error) {
	return models.GetChannelStatisticData()
}
