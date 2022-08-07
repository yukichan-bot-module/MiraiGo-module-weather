package model

import "gorm.io/gorm"

// User 用户
type User struct {
	gorm.Model
	Name      string
	Uin       int64
	Longitude float64
	Latitude  float64
	Date      string // 最后调用日期 220801
	Times     int    // 调用次数
}
