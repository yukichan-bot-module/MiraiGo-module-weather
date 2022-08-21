package weather

import (
	"fmt"
	"os"
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
	Key       string  `yaml:"key"`
	Limit     int     `yaml:"limit"`
	Admin     []int64 `yaml:"admin"`
	Allowed   []int64 `yaml:"allowed"`
	BlackList []int64 `yaml:"blacklist"`
	WhiteList []int64 `yaml:"whitelist"`
	DB        struct {
		Type  string `yaml:"type"`
		MySQL struct {
			Username string `yaml:"username"`
			Password string `yaml:"password"`
			Host     string `yaml:"host"`
			Port     string `yaml:"port"`
			Database string `yaml:"database"`
			Charset  string `yaml:"charset"`
		} `yaml:"mysql"`
		SQLite struct {
			Path string `yaml:"path"`
		} `yaml:"sqlite"`
	} `yaml:"db"`
	Daily []struct {
		GroupCode int64   `yaml:"group"`
		Longitude float64 `yaml:"longitude"`
		Latitude  float64 `yaml:"latitude"`
		Time      string  `yaml:"time"`
		Type      string  `yaml:"type"`
		Notify    string  `yaml:"notify"`
	} `yaml:"daily"`
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
		replyMsgString := groupWeatherService(msg)
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
		replyMsgString := privateWeatherService(msg)
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
		_groupCode := d.GroupCode
		_longitude := d.Longitude
		_latitude := d.Latitude
		_time := d.Time
		_type := d.Type
		_notify := d.Notify
		s.Every(1).Day().At(_time).Do(func() {
			caiyunAPI := service.NewCaiyun(weatherConfig.Key)
			weatherString := ""
			switch _type {
			case "today":
				weatherString, _ = caiyunAPI.Today(_longitude, _latitude)
			case "tomorrow":
				weatherString, _ = caiyunAPI.Tomorrow(_longitude, _latitude)
			default:
				weatherString = "配置文件错误，请检查"
			}
			if weatherString == "" {
				return
			}
			replyMsgString := _notify + "\n" + weatherString
			msg := message.NewSendingMessage().Append(message.NewText(replyMsgString))
			b.SendGroupMessage(_groupCode, msg)
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
func privateWeatherService(privateMsg *message.PrivateMessage) string {
	sender := privateMsg.Sender
	msg := privateMsg.ToString()
	if strings.HasPrefix(msg, "修改地址 ") {
		return updateLocation(sender, msg)
	}
	caiyunAPI := service.NewCaiyun(weatherConfig.Key)
	switch msg {
	case "实时天气":
		return callWeatherAPI(sender.Uin, caiyunAPI.RealTime)
	case "出门建议":
		return callWeatherAPI(sender.Uin, caiyunAPI.Rain)
	case "今天天气":
		return callWeatherAPI(sender.Uin, caiyunAPI.Today)
	case "明天天气":
		return callWeatherAPI(sender.Uin, caiyunAPI.Tomorrow)
	}
	return ""
}

// groupWeatherService 群聊服务
func groupWeatherService(groupMsg *message.GroupMessage) string {
	sender := groupMsg.Sender
	msg := groupMsg.ToString()
	// 检查管理员指令
	if strings.HasPrefix(msg, ".weather.") && isAdmin(sender.Uin) {
		switch {
		case strings.HasPrefix(msg, ".weather.clear.times "):
			return clearUserTimes(msg)
		case strings.HasPrefix(msg, ".weather.blacklist.add "):
			return addUserToBlacklist(msg)
		case strings.HasPrefix(msg, ".weather.blacklist.remove "):
			return removeUserFromBlacklist(msg)
		case strings.HasPrefix(msg, ".weather.whitelist.add "):
			return addUserToWhitelist(msg)
		case strings.HasPrefix(msg, ".weather.whitelist.remove "):
			return removeUserFromWhitelist(msg)
		case msg == ".weather.allowed":
			return addGroupToAllowed(groupMsg.GroupCode)
		case msg == ".weather.disallowed":
			return removeGroupFromAllowed(groupMsg.GroupCode)
		default:
			return ""
		}
	}
	// 忽略未开启功能的群组
	if !isAllowedGroup(groupMsg.GroupCode) {
		return ""
	}
	// 解析用户指令
	if strings.HasPrefix(msg, "修改地址 ") {
		return updateLocation(sender, msg)
	}
	caiyunAPI := service.NewCaiyun(weatherConfig.Key)
	switch msg {
	case "实时天气":
		return callWeatherAPI(sender.Uin, caiyunAPI.RealTime)
	case "出门建议":
		return callWeatherAPI(sender.Uin, caiyunAPI.Rain)
	case "今天天气":
		return callWeatherAPI(sender.Uin, caiyunAPI.Today)
	case "明天天气":
		return callWeatherAPI(sender.Uin, caiyunAPI.Tomorrow)
	}
	return ""
}

// updateLocation 更新用户地址
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
	dbService := service.NewDBService(database.GetDB())
	_, err = dbService.GetUser(sender.Uin)
	if err == gorm.ErrRecordNotFound {
		if err = dbService.CreateUser(sender.Uin, sender.Nickname, longitude, latitude); err != nil {
			logger.WithError(err).Errorf("Fail to create user.")
			return DatabaseErrorMessage
		}
		return "保存成功。"
	}
	if err != nil {
		logger.WithError(err).Errorf("Fail to get user.")
		return DatabaseErrorMessage
	}
	if err := dbService.UpdateUserInfo(sender.Uin, sender.Nickname, longitude, latitude); err != nil {
		logger.WithError(err).Errorf("Fail to update user.")
		return DatabaseErrorMessage
	}
	return "保存成功。"
}

func callWeatherAPI(uin int64, apiCalled func(float64, float64) (string, error)) string {
	dbService := service.NewDBService(database.GetDB())
	longitude, latitude, err := dbService.GetUserLocation(uin)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "未查询到地址信息，可通过群聊或私聊发送「修改地址 经度 纬度」来添加地址信息，经纬度信息需保留四位小数以上以保证精确度。示例：「修改地址 101.6656 39.2072」。发送后数据会被保存，如需修改使用同样的指令即可。\n\n注：不支持城市，因为城市准确度很差。本 bot 使用的 api 为付费 api，请勿滥用。"
		}
		logger.WithError(err).Errorf("Fail to get user location.")
		return DatabaseErrorMessage
	}
	if !inWhitelist(uin) {
		times, err := dbService.GetUserTimes(uin)
		if err != nil {
			logger.WithError(err).Errorf("Fail to get user times.")
			return DatabaseErrorMessage
		}
		if times >= weatherConfig.Limit {
			return fmt.Sprintf("彩云天气为付费 api，万次 8 元，为防止滥用，当前每人每日使用次数上限为 %d 次。", weatherConfig.Limit)
		}
	}
	if err := dbService.IncreaseUserTimes(uin); err != nil {
		logger.WithError(err).Errorf("Fail to increase user times.")
		return DatabaseErrorMessage
	}
	apiResponse, err := apiCalled(longitude, latitude)
	if err != nil {
		return "调用天气 api 时发生错误。可能是网络问题或 api 使用次数耗尽。"
	}
	return apiResponse
}

// clearUserTimes 清除用户调用次数
func clearUserTimes(msg string) string {
	if len(msg) <= 21 {
		return ""
	}
	uinString := msg[21:]
	uin, err := strconv.ParseInt(uinString, 10, 64)
	if err != nil {
		return fmt.Sprintf("解析失败，「%s」不是正确的 uin。", uinString)
	}
	dbService := service.NewDBService(database.GetDB())
	err = dbService.ClearUserTimes(uin)
	if err != nil {
		logger.WithError(err).Errorf("Fail to clear user times.")
		return "数据库错误，请检查后台日志。"
	}
	return "清除用户调用次数成功。"
}

// addUserToBlacklist 添加用户到黑名单
func addUserToBlacklist(msg string) string {
	if len(msg) <= 23 {
		return ""
	}
	uinString := msg[23:]
	uin, err := strconv.ParseInt(uinString, 10, 64)
	if err != nil {
		return fmt.Sprintf("解析失败，「%s」不是正确的 uin。", uinString)
	}
	if inBlacklist(uin) {
		return "该用户已在黑名单中。"
	}
	weatherConfig.BlackList = append(weatherConfig.BlackList, uin)
	err = updateWeatherConfigFile(weatherConfig)
	if err != nil {
		logger.WithError(err).Errorf("Fail to update config file.")
		return "在更新配置文件时出现了错误，请查看后台日志。"
	}
	return fmt.Sprintf("成功添加用户「%d」到黑名单。", uin)
}

// removeUserFromBlacklist 将用户从黑名单中移除
func removeUserFromBlacklist(msg string) string {
	if len(msg) <= 26 {
		return ""
	}
	uinString := msg[26:]
	uin, err := strconv.ParseInt(uinString, 10, 64)
	if err != nil {
		return fmt.Sprintf("解析失败，「%s」不是正确的 uin。", uinString)
	}
	if !inBlacklist(uin) {
		return "该用户不在黑名单中。"
	}
	for i, v := range weatherConfig.BlackList {
		if v == uin {
			weatherConfig.BlackList = append(weatherConfig.BlackList[:i], weatherConfig.BlackList[i+1:]...)
			break
		}
	}
	err = updateWeatherConfigFile(weatherConfig)
	if err != nil {
		logger.WithError(err).Errorf("Fail to update config file.")
		return "在更新配置文件时出现了错误，请查看后台日志。"
	}
	return fmt.Sprintf("成功将用户「%d」从黑名单中移除。", uin)
}

// addUserToWhitelist 添加用户到白名单
func addUserToWhitelist(msg string) string {
	if len(msg) <= 23 {
		return ""
	}
	uinString := msg[23:]
	uin, err := strconv.ParseInt(uinString, 10, 64)
	if err != nil {
		return fmt.Sprintf("解析失败，「%s」不是正确的 uin。", uinString)
	}
	if inWhitelist(uin) {
		return "该用户已在白名单中。"
	}
	weatherConfig.WhiteList = append(weatherConfig.WhiteList, uin)
	err = updateWeatherConfigFile(weatherConfig)
	if err != nil {
		logger.WithError(err).Errorf("Fail to update config file.")
		return "在更新配置文件时出现了错误，请查看后台日志。"
	}
	return fmt.Sprintf("成功添加用户「%d」到白名单。", uin)
}

// removeUserFromWhitelist 将用户从白名单中移除
func removeUserFromWhitelist(msg string) string {
	if len(msg) <= 26 {
		return ""
	}
	uinString := msg[26:]
	uin, err := strconv.ParseInt(uinString, 10, 64)
	if err != nil {
		return fmt.Sprintf("解析失败，「%s」不是正确的 uin。", uinString)
	}
	if !inWhitelist(uin) {
		return "该用户不在白名单中。"
	}
	for i, v := range weatherConfig.WhiteList {
		if v == uin {
			weatherConfig.WhiteList = append(weatherConfig.WhiteList[:i], weatherConfig.WhiteList[i+1:]...)
			break
		}
	}
	err = updateWeatherConfigFile(weatherConfig)
	if err != nil {
		logger.WithError(err).Errorf("Fail to update config file.")
		return "在更新配置文件时出现了错误，请查看后台日志。"
	}
	return fmt.Sprintf("成功将用户「%d」从白名单中移除。", uin)
}

// addGroupToAllowed 添加群组到许可名单
func addGroupToAllowed(groupCode int64) string {
	if isAllowedGroup(groupCode) {
		return "本群已加入许可名单。"
	}
	weatherConfig.Allowed = append(weatherConfig.Allowed, groupCode)
	err := updateWeatherConfigFile(weatherConfig)
	if err != nil {
		logger.WithError(err).Errorf("Fail to update config file.")
		return "在更新配置文件时出现了错误，请查看后台日志。"
	}
	return fmt.Sprintf("成功添加群「%d」到许可名单。", groupCode)
}

// removeGroupFromAllowed 将群组从许可名单中移除
func removeGroupFromAllowed(groupCode int64) string {
	if !isAllowedGroup(groupCode) {
		return "本群不在许可名单中。"
	}
	for i, v := range weatherConfig.Allowed {
		if v == groupCode {
			weatherConfig.Allowed = append(weatherConfig.Allowed[:i], weatherConfig.Allowed[i+1:]...)
			break
		}
	}
	err := updateWeatherConfigFile(weatherConfig)
	if err != nil {
		logger.WithError(err).Errorf("Fail to update config file.")
		return "在更新配置文件时出现了错误，请查看后台日志。"
	}
	return fmt.Sprintf("成功将群「%d」从许可名单中移除。", groupCode)
}

func updateWeatherConfigFile(newConfig Config) error {
	path := config.GlobalConfig.GetString("aimerneige.weather.path")
	if path == "" {
		path = "./weather.yaml"
	}
	data, err := yaml.Marshal(newConfig)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	weatherConfig = newConfig
	return nil
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
