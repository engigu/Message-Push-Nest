package statistic_service

import (
	"message-nest/models"
)

type StatisticService struct {
}

func (sw *StatisticService) GetStatisticData() (models.StatisticData, error) {
	return models.GetStatisticData()
}
