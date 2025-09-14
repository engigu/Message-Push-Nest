package util

import (
	"fmt"
	"os"
	"time"
	"github.com/shirou/gopsutil/v3/process"
)

// GetMemoryUsage returns current process memory usage from OS perspective
func GetMemoryUsage() map[string]string {
	// Get current process
	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return map[string]string{
			"memory_usage": "获取失败",
			"memory_info":  "程序内存使用: 获取失败",
		}
	}
	
	// Get memory info
	memInfo, err := p.MemoryInfo()
	if err != nil {
		return map[string]string{
			"memory_usage": "获取失败",
			"memory_info":  "程序内存使用: 获取失败",
		}
	}
	
	// Convert bytes to MB
	rssMB := float64(memInfo.RSS) / 1024 / 1024
	vmsMB := float64(memInfo.VMS) / 1024 / 1024
	
	// Get process creation time
	createTime, err := p.CreateTime()
	var uptime string
	if err == nil {
		startTime := time.Unix(createTime/1000, 0)
		duration := time.Since(startTime)
		uptime = formatDuration(duration)
	} else {
		uptime = "获取失败"
	}
	
	return map[string]string{
		"memory_usage": fmt.Sprintf("%.1f MB", rssMB),
		"memory_info":  fmt.Sprintf("程序内存使用: %.1f MB", rssMB),
		"vms_usage":    fmt.Sprintf("%.1f MB", vmsMB),
		"uptime":       uptime,
		"uptime_info":  fmt.Sprintf("程序运行时间: %s", uptime),
	}
}

// formatDuration formats duration to human readable format
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分钟%d秒", minutes, seconds)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}
