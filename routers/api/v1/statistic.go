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

	// 获取类型参数
	statisticType := c.Query("type")
	
	msgService := statistic_service.StatisticService{}
	
	switch statisticType {
	case "basic":
		data, err := msgService.GetBasicStatisticData()
		if err != nil {
			appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取基础统计失败！原因：%s", err), nil)
			return
		}
		appG.CResponse(http.StatusOK, "获取基础统计成功", data)
	case "trend":
		data, err := msgService.GetTrendStatisticData()
		if err != nil {
			appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取趋势统计失败！原因：%s", err), nil)
			return
		}
		appG.CResponse(http.StatusOK, "获取趋势统计成功", data)
	case "channels":
		data, err := msgService.GetChannelStatisticData()
		if err != nil {
			appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取渠道统计失败！原因：%s", err), nil)
			return
		}
		appG.CResponse(http.StatusOK, "获取渠道统计成功", data)
	default:
		// 默认返回完整统计数据（保持向后兼容）
		data, err := msgService.GetStatisticData()
		if err != nil {
			appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取失败！原因：%s", err), nil)
			return
		}
		appG.CResponse(http.StatusOK, "获取成功", data)
	}
}

