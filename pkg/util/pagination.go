package util

import (
	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"
)

func GetPageSize(c *gin.Context) (int, int) {
	result := 0
	page := com.StrTo(c.Query("page")).MustInt()
	size := com.StrTo(c.Query("size")).MustInt()
	if page > 0 {
		result = (page - 1) * size
	}
	return result, size
}
