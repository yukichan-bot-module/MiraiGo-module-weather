package weather

import (
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/yukichan-bot-module/MiraiGo-module-weather/internal/database"
	"github.com/yukichan-bot-module/MiraiGo-module-weather/internal/service"
	"gopkg.in/yaml.v3"
)

var instance *weather
var logger = utils.GetModuleLogger("com.aimerneige.weather")

var weatherConfig struct {
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
}

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
	// update config
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
		if inBlacklist(msg.Sender.Uin) {
			return
		}
		if !isAllowedGroup(msg.GroupCode) {
			return
		}
		replyMsgString := service.GroupWeatherService(msg.Sender.Uin, msg.ToString())
		if replyMsgString == "" {
			return
		}
		replyMsg := message.NewSendingMessage().Append(message.NewText(replyMsgString))
		c.SendGroupMessage(msg.GroupCode, replyMsg)
	})
	b.OnPrivateMessage(func(c *client.QQClient, msg *message.PrivateMessage) {
		if inBlacklist(msg.Sender.Uin) {
			return
		}
		replyMsgString := service.PrivateWeatherService(msg.Sender.Uin, msg.ToString())
		if replyMsgString == "" {
			return
		}
		replyMsg := message.NewSendingMessage().Append(message.NewText(replyMsgString))
		c.SendPrivateMessage(msg.Sender.Uin, replyMsg)
	})
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
