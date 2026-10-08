package setting

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"github.com/go-ini/ini"
	"message-nest/pkg/logging"
)

type App struct {
	JwtSecret string
	LogLevel  string
}

var AppSetting = &App{}

type Server struct {
	HttpPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	UrlPrefix string
}

var ServerSetting = &Server{}

type Database struct {
	Type        string
	User        string
	Password    string
	Host        string
	Port        int
	Name        string
	TablePrefix string
	SqlDebug    string
	Ssl         string
}

var DatabaseSetting = &Database{}

var cfg *ini.File

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func createConfFolder() {
	// 检查目录是否存在
	dir := "conf/"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return
		}
	}
}

// Setup initialize the configuration instance
func Setup() {
	var err error
	intPath := "conf/app.ini"
	createConfFolder()

	if fileExists(intPath) {
		logging.Config.Infof("正在从配置文件 %s 启动服务...", intPath)
		cfg, err = ini.Load(intPath)
		if err != nil {
			logging.Config.Fatalf("解析配置文件 '%s' 失败: %v", intPath, err)
		}

		mapTo("app", AppSetting)
		mapTo("server", ServerSetting)
		mapTo("database", DatabaseSetting)
	} else {
		logging.Config.Infof("配置文件 %s 不存在，从环境变量加载配置...", intPath)
		loadConfigFromEnv()
	}

	ServerSetting.ReadTimeout = ServerSetting.ReadTimeout * time.Second
	ServerSetting.WriteTimeout = ServerSetting.WriteTimeout * time.Second

	ensureJwtSecret()
}

func ensureJwtSecret() {
	if AppSetting.JwtSecret == "" || AppSetting.JwtSecret == "message-nest" {
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			logging.Security.Fatalf("动态生成安全随机 JWT 密钥失败: %v", err)
		}
		AppSetting.JwtSecret = hex.EncodeToString(randomBytes)
		logging.Security.Warnln("未配置 JWT_SECRET 或使用了不安全的默认值('message-nest')。")
		logging.Security.Warnln("已为本次运行动态生成高熵安全随机密钥。")
		logging.Security.Warnln("注意：服务重启后之前颁发的 Token 将失效，生产环境请务必显式配置持久化的 JWT_SECRET！")
	}
}

// mapTo map section
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		logging.Config.Fatalf("映射配置分区 [%s] 失败: %v", section, err)
	}
}
