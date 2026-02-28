package util

import (
	"database/sql/driver"
	"fmt"
	"time"
	_ "time/tzdata"
)

const timeFormat = "2006-01-02 15:04:05"
const timezone = "Asia/Shanghai"

// 全局定义
type Time time.Time

func (t Time) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(timeFormat)+2)
	b = append(b, '"')
	// 转换到东八区时间再格式化
	loc, _ := time.LoadLocation(timezone)
	b = time.Time(t).In(loc).AppendFormat(b, timeFormat)
	b = append(b, '"')

	return b, nil
}

func (t *Time) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+timeFormat+`"`, string(data), time.Local)
	*t = Time(now)
	return
}

func (t Time) String() string {
	// 转换到东八区时间再格式化
	loc, _ := time.LoadLocation(timezone)
	return time.Time(t).In(loc).Format(timeFormat)
}

func (t Time) local() time.Time {
	loc, _ := time.LoadLocation(timezone)
	return time.Time(t).In(loc)
}

func (t Time) Value() (driver.Value, error) {
	var zeroTime time.Time
	var ti = time.Time(t)
	if ti.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return ti, nil
}

func (t *Time) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = Time(value)
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}

// GetNowTimeStr 获取当前的时间字符串
func GetNowTimeStr() string {
	chinaTime := GetNowTime()
	formattedTime := chinaTime.Format("2006-01-02 15:04:05")
	return formattedTime
}

// GetNowTimeStr 获取当前的时间字符串
func GetNowTime() time.Time {
	currentTimeUTC := time.Now().UTC()
	chinaTime := currentTimeUTC.Add(8 * time.Hour)
	return chinaTime
}
