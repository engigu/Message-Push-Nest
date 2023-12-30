package util

import (
	"time"
)

func TimeStampToYmdMHD(stamp int) string {
	timestamp := int64(stamp)
	timeObj := time.Unix(timestamp, 0)
	formattedTime := timeObj.Format("2006-01-02 15:04:05")
	return formattedTime
}

func GetNowTimeStampToYmdMHD() string {
	now := time.Now().Unix()
	timeObj := time.Unix(now, 0)
	formattedTime := timeObj.Format("2006-01-02 15:04:05")
	return formattedTime
}
