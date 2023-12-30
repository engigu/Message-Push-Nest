package util

import (
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"

	"message-nest/pkg/setting"
)

// GetPage get page parameters
func GetPage(c *gin.Context) int {
	result := 0
	page := com.StrTo(c.Query("page")).MustInt()
	if page > 0 {
		result = (page - 1) * setting.AppSetting.PageSize
	}

	return result
}

func GetPageSize(c *gin.Context) (int, int) {
	result := 0
	page := com.StrTo(c.Query("page")).MustInt()
	size := com.StrTo(c.Query("size")).MustInt()
	if page > 0 {
		result = (page - 1) * size
	}
	return result, size
}
