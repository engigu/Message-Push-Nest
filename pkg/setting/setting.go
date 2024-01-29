package setting

import (
	"log"
	"os"
	"time"

	"github.com/go-ini/ini"
)

type App struct {
	JwtSecret string

	RuntimeRootPath string
	LogLevel        string
	InitData        string
}

var AppSetting = &App{}

type Server struct {
	RunMode      string
	HttpPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	EmbedHtml string
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
}

var DatabaseSetting = &Database{}

var cfg *ini.File

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// Setup initialize the configuration instance
func Setup() {
	var err error
	intPath := "conf/app.ini"

	if fileExists(intPath) {
		log.Printf("[message-nest] start server from %s.", intPath)
		cfg, err = ini.Load(intPath)
		if err != nil {
			log.Fatalf("[message-nest] setting.Setup, fail to parse 'conf/app.ini': %v", err)
		}

		mapTo("app", AppSetting)
		mapTo("server", ServerSetting)
		mapTo("database", DatabaseSetting)
	} else {
		log.Printf("[message-nest] %s is not exists, start server from env vars.", intPath)
		loadConfigFromEnv()
	}

	ServerSetting.ReadTimeout = ServerSetting.ReadTimeout * time.Second
	ServerSetting.WriteTimeout = ServerSetting.WriteTimeout * time.Second

}

// mapTo map section
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("[message-nest] Cfg.MapTo %s err: %v", section, err)
	}
}
