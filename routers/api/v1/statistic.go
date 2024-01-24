package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"message-nest/pkg/app"
	"message-nest/service/statistic_service"
	"net/http"
)

// GetStatisticData 获取发送统计数据
func GetStatisticData(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	msgService := statistic_service.StatisticService{}
	data, err := msgService.GetStatisticData()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取失败！原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取成功", data)
}
