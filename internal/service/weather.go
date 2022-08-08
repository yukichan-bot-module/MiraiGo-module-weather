package service

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Mrs4s/MiraiGo/message"
	"github.com/yukichan-bot-module/MiraiGo-module-weather/internal/database"
	"github.com/yukichan-bot-module/MiraiGo-module-weather/internal/weather"
	"gorm.io/gorm"
)

// DatabaseErrorMessage 数据库错误信息
const DatabaseErrorMessage string = "数据库错误，请联系开发者修 bug。开源地址：https://github.com/yukichan-bot-module/MiraiGo-module-weather"

var limit int = 3
var apiKey string
var whitelist []int64
var adminList []int64

// UpdateLimit 更新 limit
func UpdateLimit(l int) {
	limit = l
}

// UpdateAPIKey 更新 api key
func UpdateAPIKey(key string) {
	apiKey = key
}

// UpdateWhiteList 更新白名单
func UpdateWhiteList(list []int64) {
	whitelist = list
}

// UpdateAdminList 更新管理员名单
func UpdateAdminList(list []int64) {
	adminList = list
}

// PrivateWeatherService 私聊服务
func PrivateWeatherService(sender *message.Sender, msg string) string {
	if strings.HasPrefix(msg, "修改地址 ") {
		return updateLocation(sender, msg)
	}
	switch msg {
	case "实时天气":
		return getRealTimeWeather(sender.Uin)
	case "出门建议":
		return "这个功能还没有开发呢。"
	case "明天天气":
		return "这个功能还没有开发呢。"
	}
	return ""
}

// GroupWeatherService 群聊服务
func GroupWeatherService(sender *message.Sender, msg string) string {
	if strings.HasPrefix(msg, ".weather.") && isAdmin(sender.Uin) {
		switch {
		case strings.HasPrefix(msg, ".weather.clear.times "):
			return clearUserTimes(sender.Uin, msg)
		case strings.HasPrefix(msg, ".weather.blacklist.add "):
			return addUserToBlacklist(sender.Uin, msg)
		case strings.HasPrefix(msg, ".weather.whitelist.add "):
			return addUserToWhitelist(sender.Uin, msg)
		case strings.HasPrefix(msg, ".weather.allowed "):
			return addGroupToAllowed(sender.Uin, msg)
		default:
			return ""
		}
	}
	if strings.HasPrefix(msg, "修改地址 ") {
		return updateLocation(sender, msg)
	}
	switch msg {
	case "实时天气":
		return getRealTimeWeather(sender.Uin)
	case "出门建议":
		return "这个功能还没有开发呢。"
	case "明天天气":
		return "这个功能还没有开发呢。"
	}
	return ""
}

func updateLocation(sender *message.Sender, msg string) string {
	parts := strings.Split(msg, " ")
	if len(parts) != 3 {
		return "解析失败，请检查格式。正确的格式：「修改地址 经度 纬度」，示例：「修改地址 101.6656 39.2072」。"
	}
	var longitude float64
	var latitude float64
	var err error
	longitude, err = strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Sprintf("解析失败，经度「%s」不是正确的数字。", parts[1])
	}
	latitude, err = strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return fmt.Sprintf("解析失败，纬度「%s」不是正确的数字。", parts[2])
	}
	if longitude < -180.0 || longitude > 180.0 {
		return fmt.Sprintf("解析失败，「%.4f」不是正确的经度。", longitude)
	}
	if latitude < -90.0 || latitude > 90.0 {
		return fmt.Sprintf("解析失败，「%.4f」不是正确的纬度。", latitude)
	}
	log.Println("经纬度", longitude, latitude)
	dbService := NewDBService(database.GetDB())
	_, err = dbService.GetUser(sender.Uin)
	if err == gorm.ErrRecordNotFound {
		if err = dbService.CreateUser(sender.Uin, sender.Nickname, longitude, latitude); err != nil {
			log.Println("Fail to create user.", err)
			return DatabaseErrorMessage
		}
		return "保存成功。"
	}
	if err != nil {
		log.Println("Fail to get user.")
		return DatabaseErrorMessage
	}
	if err := dbService.UpdateUserInfo(sender.Uin, sender.Nickname, longitude, latitude); err != nil {
		log.Print("Fail to update user.")
		return DatabaseErrorMessage
	}
	return "保存成功。"
}

func getRealTimeWeather(uin int64) string {
	dbService := NewDBService(database.GetDB())
	longitude, latitude, err := dbService.GetUserLocation(uin)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "未查询到地址信息，可通过群聊或私聊发送「修改地址 经度 纬度」来添加地址信息，经纬度信息需保留四位小数以上。示例：「修改地址 101.6656 39.2072」。发送后数据会被保存，如需修改使用同样的指令即可。"
		}
		log.Println("Fail to get user location.")
		return DatabaseErrorMessage
	}
	if !inWhitelist(uin) {
		times, err := dbService.GetUserTimes(uin)
		if err != nil {
			log.Println("Fail to get user times.")
			return DatabaseErrorMessage
		}
		if times >= limit {
			return fmt.Sprintf("彩云天气为付费 api，万次 8 元，为防止滥用，当前每人每日使用次数上限为 %d 次。", limit)
		}
	}
	if err := dbService.IncreaseUserTimes(uin); err != nil {
		log.Println("Fail to increase user times")
		return DatabaseErrorMessage
	}
	caiyun := weather.NewCaiyun(apiKey)
	rtWeather, err := caiyun.RealTime(longitude, latitude)
	if err != nil {
		return "调用天气 api 时发生错误。可能是网络问题或 api 使用次数耗尽。"
	}
	return rtWeather
}

func clearUserTimes(uin int64, msg string) string {
	return "这个功能还没有开发呢。"
}

func addUserToBlacklist(uin int64, msg string) string {
	return "这个功能还没有开发呢。"
}

func addUserToWhitelist(uin int64, msg string) string {
	return "这个功能还没有开发呢。"
}

func addGroupToAllowed(uin int64, msg string) string {
	return "这个功能还没有开发呢。"
}

func isAdmin(uin int64) bool {
	for _, v := range adminList {
		if v == uin {
			return true
		}
	}
	return false
}

func inWhitelist(uin int64) bool {
	for _, v := range whitelist {
		if v == uin {
			return true
		}
	}
	return false
}
