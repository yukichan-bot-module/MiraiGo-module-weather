package service

import (
	"github.com/yukichan-bot-module/MiraiGo-module-weather/internal/database/model"
	"gorm.io/gorm"
)

// DBService 数据库服务
type DBService struct {
	db *gorm.DB
}

// NewDBService 创建数据库服务
func NewDBService(db *gorm.DB) *DBService {
	return &DBService{
		db: db,
	}
}

// CreateUser 新建用户
func (d *DBService) CreateUser(uin int64, name string, longitude float64, latitude float64) error {
	return d.db.Create(&model.User{
		Uin:       uin,
		Name:      name,
		Longitude: longitude,
		Latitude:  latitude,
		Times:     0,
	}).Error
}

// GetUser 获取用户
func (d *DBService) GetUser(uin int64) (model.User, error) {
	var user model.User
	err := d.db.Where("uin = ?", uin).First(&user).Error
	return user, err
}

// GetUserLocation 获取用户位置
func (d *DBService) GetUserLocation(uin int64) (float64, float64, error) {
	var user model.User
	err := d.db.Select("longitude", "latitude").Where("uin = ?", uin).First(&user).Error
	return user.Longitude, user.Latitude, err
}

// GetUserTimes 获得用户的调用次数
func (d *DBService) GetUserTimes(uin int64) (int, error) {
	var user model.User
	err := d.db.Select("times").Where("uin = ?", uin).First(&user).Error
	return user.Times, err
}

// UpdateUserInfo 更新用户信息
func (d *DBService) UpdateUserInfo(uin int64, name string, longitude float64, latitude float64) error {
	return d.db.Model(&model.User{}).Where("uin = ?", uin).Update("name", name).Update("longitude", longitude).Update("latitude", latitude).Error
}

// UpdateUserTimes 更新用户调用次数信息
func (d *DBService) UpdateUserTimes(uin int64, times int) error {
	return d.db.Model(&model.User{}).Where("uin = ?", uin).Update("times", times).Error
}

// IncreaseUserTimes 增加用户调用次数
func (d *DBService) IncreaseUserTimes(uin int64) error {
	return d.db.Model(&model.User{}).Where("uin = ?", uin).Update("times", gorm.Expr("times + 1")).Error
}

// ClearUserTimes 清空用户调用次数
func (d *DBService) ClearUserTimes(uin int64) error {
	return d.db.Model(&model.User{}).Where("uin = ?", uin).Update("times", 0).Error
}

// ClearAllUserTimes 清空全部用户调用次数
// view https://gorm.io/docs/update.html#Block-Global-Updates
// for more detail about `.Where("1 = 1")`
func (d *DBService) ClearAllUserTimes() error {
	return d.db.Model(&model.User{}).Where("1 = 1").Update("times", 0).Error
}
