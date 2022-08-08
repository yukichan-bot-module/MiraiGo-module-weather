package model

import "gorm.io/gorm"

// User 用户
type User struct {
	gorm.Model
	Uin       int64 `gorm:"unique_index"`
	Name      string
	Longitude float64
	Latitude  float64
	Times     int // 调用次数
}
