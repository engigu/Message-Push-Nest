package setting

import (
	"github.com/unknwon/com"
	"log"
	"os"
)

var optionValueMap = map[string]string{}

// getOptionEnvValue 获取必须声明的环境变量
func getOptionEnvValue(key string, defaultV string) string {
	value := os.Getenv(key)
	result := ""
	if value == "" {
		result = defaultV
	} else {
		result = value
	}
	optionValueMap[key] = result
	return result
}

// getMustEnvValue 获取必须声明的环境变量
func getMustEnvValue(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Printf("[message-nest] you must assign env: %s", key)
		return ""
	} else {
		return value
	}
}

// printOptionValue 打印可选环境变量值
func printOptionValue() {
	for key, val := range optionValueMap {
		log.Printf("[message-nest] current option env: %s, value: %s", key, val)
	}
}

// loadConfigFromEnv 从环境变量加载配置
func loadConfigFromEnv() {
	AppSetting.JwtSecret = getOptionEnvValue("JWT_SECRET", "message-nest")
	AppSetting.LogLevel = getOptionEnvValue("LOG_LEVEL", "INFO")
	AppSetting.InitData = getOptionEnvValue("INIT_DATA", "")

	ServerSetting.RunMode = getOptionEnvValue("RUN_MODE", "release")
	ServerSetting.HttpPort = 8000
	ServerSetting.ReadTimeout = 60
	ServerSetting.WriteTimeout = 60

	DatabaseSetting.Type = "mysql"
	DatabaseSetting.Host = getMustEnvValue("MYSQL_HOST")
	DatabaseSetting.Port = com.StrTo(getMustEnvValue("MYSQL_PORT")).MustInt()
	DatabaseSetting.User = getMustEnvValue("MYSQL_USER")
	DatabaseSetting.Password = getMustEnvValue("MYSQL_PASSWORD")
	DatabaseSetting.Name = getMustEnvValue("MYSQL_DB")
	DatabaseSetting.TablePrefix = getMustEnvValue("MYSQL_TABLE_PREFIX")
	DatabaseSetting.SqlDebug = getOptionEnvValue("SQL_DEBUG", "disable")

	printOptionValue()
}
