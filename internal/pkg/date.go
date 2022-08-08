package pkg

import "time"

// GetTodayDate 获得今天的日期字符串 20220801
func GetTodayDate() string {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	return time.Now().In(loc).Format("20060102")
}
