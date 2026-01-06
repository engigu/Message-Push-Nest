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
	case "send_stats":
		// 新增：基于 send_stats 表的统计数据
		data, err := msgService.GetSendStatsData()
		if err != nil {
			appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取发送统计失败！原因：%s", err), nil)
			return
		}
		appG.CResponse(http.StatusOK, "获取发送统计成功", data)
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

// GetSendStatsByTask 获取指定任务的发送统计数据
func GetSendStatsByTask(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
	)

	taskID := c.Query("task_id")
	if taskID == "" {
		appG.CResponse(http.StatusBadRequest, "任务ID不能为空", nil)
		return
	}

	msgService := statistic_service.StatisticService{
		TaskID: taskID,
	}

	data, err := msgService.GetSendStatsByTask()
	if err != nil {
		appG.CResponse(http.StatusInternalServerError, fmt.Sprintf("获取任务统计失败！原因：%s", err), nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取任务统计成功", data)
}

