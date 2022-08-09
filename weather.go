package weather

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/go-co-op/gocron"
	"github.com/yukichan-bot-module/MiraiGo-module-weather/internal/database"
	"github.com/yukichan-bot-module/MiraiGo-module-weather/internal/service"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// Config 模块配置
type Config struct {
	Key       string  `json:"key"`
	Limit     int     `json:"limit"`
	Admin     []int64 `json:"admin"`
	Allowed   []int64 `json:"allowed"`
	BlackList []int64 `json:"blacklist"`
	WhiteList []int64 `json:"whitelist"`
	DB        struct {
		Type  string `json:"type"`
		MySQL struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Host     string `json:"host"`
			Port     string `json:"port"`
			Database string `json:"database"`
			Charset  string `json:"charset"`
		} `json:"mysql"`
		SQLite struct {
			Path string `json:"path"`
		} `json:"sqlite"`
	} `json:"db"`
	Daily []struct {
		Group     string  `json:"group"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Time      string  `json:"time"`
	} `json:"daily"`
}

// DatabaseErrorMessage 数据库错误信息
const DatabaseErrorMessage string = "数据库错误，请联系开发者修 bug。开源地址：https://github.com/yukichan-bot-module/MiraiGo-module-weather"

var instance *weather
var logger = utils.GetModuleLogger("com.aimerneige.weather")
var weatherConfig Config

type weather struct {
}

func init() {
	instance = &weather{}
	bot.RegisterModule(instance)
}

func (w *weather) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       "com.aimerneige.weather",
		Instance: instance,
	}
}

// Init 初始化过程
// 在此处可以进行 Module 的初始化配置
// 如配置读取
func (w *weather) Init() {
	path := config.GlobalConfig.GetString("aimerneige.weather.path")
	if path == "" {
		path = "./weather.yaml"
	}
	bytes := utils.ReadFile(path)
	if err := yaml.Unmarshal(bytes, &weatherConfig); err != nil {
		logger.WithError(err).Errorf("Unable to read config file in %s", path)
	}
}

// PostInit 第二次初始化
// 再次过程中可以进行跨 Module 的动作
// 如通用数据库等等
func (w *weather) PostInit() {
	databaseType := weatherConfig.DB.Type
	switch databaseType {
	case "mysql":
		mysqlDatabase := database.MysqlDatabase{
			UserName: weatherConfig.DB.MySQL.Username,
			Password: weatherConfig.DB.MySQL.Password,
			Host:     weatherConfig.DB.MySQL.Host,
			Port:     weatherConfig.DB.MySQL.Port,
			Database: weatherConfig.DB.MySQL.Database,
			CharSet:  weatherConfig.DB.MySQL.Charset,
		}
		database.InitDatabase(mysqlDatabase)
		logger.Info("Init mysql database success ", mysqlDatabase)
	case "sqlite":
		sqliteDatabase := database.SqliteDatabase{
			FilePath: weatherConfig.DB.SQLite.Path,
		}
		database.InitDatabase(sqliteDatabase)
		logger.Info("Init sqlite database success ", sqliteDatabase)
	default:
		logger.Fatal("Unsupported database type: " + databaseType)
	}
}

// Serve 注册服务函数部分
func (w *weather) Serve(b *bot.Bot) {
	b.OnGroupMessage(func(c *client.QQClient, msg *message.GroupMessage) {
		// 忽略匿名消息
		if msg.Sender.IsAnonymous() {
			return
		}
		// 忽略黑名单用户
		if inBlacklist(msg.Sender.Uin) {
			return
		}
		// 忽略未开启功能的群组
		if !isAllowedGroup(msg.GroupCode) {
			return
		}
		replyMsgString := groupWeatherService(msg.Sender, msg.ToString())
		if replyMsgString == "" {
			return
		}
		replyMsg := message.NewSendingMessage().Append(message.NewText(replyMsgString))
		c.SendGroupMessage(msg.GroupCode, replyMsg)
	})
	b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {
		// 忽略黑名单用户
		if inBlacklist(msg.Sender.Uin) {
			return
		}
		replyMsgString := privateWeatherService(msg.Sender, msg.ToString())
		if replyMsgString == "" {
			return
		}
		replyMsg := message.NewSendingMessage().Append(message.NewText(replyMsgString))
		c.SendPrivateMessage(msg.Sender.Uin, replyMsg)
	})
	s := gocron.NewScheduler(time.UTC)
	s.Every(1).Day().At("00:00").Do(func() {
		dbService := service.NewDBService(database.GetDB())
		err := dbService.ClearAllUserTimes()
		if err != nil {
			logger.WithError(err).Errorf("Fail to clear user times.")
			for i := 0; i < 3 && err != nil; i++ {
				err = dbService.ClearAllUserTimes()
				logger.WithError(err).Errorf("Fail to clear user times. The %d times to retry.", i)
			}
		}
	})
	for _, d := range weatherConfig.Daily {
		s.Every(1).Day().At(d.Time).Do(func() {
			// TODO
		})
	}
	s.StartAsync()
}

// Start 此函数会新开携程进行调用
// ```go
//
//	go exampleModule.Start()
//
// ```
// 可以利用此部分进行后台操作
// 如 http 服务器等等
func (w *weather) Start(b *bot.Bot) {
}

// Stop 结束部分
// 一般调用此函数时，程序接收到 os.Interrupt 信号
// 即将退出
// 在此处应该释放相应的资源或者对状态进行保存
func (w *weather) Stop(b *bot.Bot, wg *sync.WaitGroup) {
	// 别忘了解锁
	defer wg.Done()
}

// privateWeatherService 私聊服务
func privateWeatherService(sender *message.Sender, msg string) string {
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

// groupWeatherService 群聊服务
func groupWeatherService(sender *message.Sender, msg string) string {
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
	dbService := service.NewDBService(database.GetDB())
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
	dbService := service.NewDBService(database.GetDB())
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
		if times >= weatherConfig.Limit {
			return fmt.Sprintf("彩云天气为付费 api，万次 8 元，为防止滥用，当前每人每日使用次数上限为 %d 次。", weatherConfig.Limit)
		}
	}
	if err := dbService.IncreaseUserTimes(uin); err != nil {
		log.Println("Fail to increase user times")
		return DatabaseErrorMessage
	}
	caiyun := service.NewCaiyun(weatherConfig.Key)
	rtWeather, err := caiyun.RealTime(longitude, latitude)
	if err != nil {
		return "调用天气 api 时发生错误。可能是网络问题或 api 使用次数耗尽。"
	}
	return rtWeather
}

// clearUserTimes 清除用户调用次数
func clearUserTimes(uin int64, msg string) string {
	return "这个功能还没有开发呢。"
}

// addUserToBlacklist 添加用户到黑名单
func addUserToBlacklist(uin int64, msg string) string {
	return "这个功能还没有开发呢。"
}

// addUserToWhitelist 添加用户到白名单
func addUserToWhitelist(uin int64, msg string) string {
	return "这个功能还没有开发呢。"
}

// addGroupToAllowed 添加群组到许可名单
func addGroupToAllowed(uin int64, msg string) string {
	return "这个功能还没有开发呢。"
}

func isAdmin(uin int64) bool {
	for _, v := range weatherConfig.Admin {
		if v == uin {
			return true
		}
	}
	return false
}

func isAllowedGroup(id int64) bool {
	for _, groupID := range weatherConfig.Allowed {
		if id == groupID {
			return true
		}
	}
	return false
}

func inBlacklist(userID int64) bool {
	for _, v := range weatherConfig.BlackList {
		if userID == v {
			return true
		}
	}
	return false
}

func inWhitelist(userID int64) bool {
	for _, v := range weatherConfig.WhiteList {
		if userID == v {
			return true
		}
	}
	return false
}
