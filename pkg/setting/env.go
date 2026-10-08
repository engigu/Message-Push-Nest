package setting

import (
	"os"
	"strings"

	"message-nest/pkg/logging"
	"github.com/unknwon/com"
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
		logging.Config.Warnf("必须配置环境变量: %s", key)
		return ""
	} else {
		return value
	}
}

// maskSecret 对敏感字符串进行前后明文脱敏处理
func maskSecret(val string) string {
	length := len(val)
	if length == 0 {
		return ""
	}
	if length <= 4 {
		return "******"
	}
	keep := 2
	if length <= 8 {
		keep = 1
	}
	return val[:keep] + "******" + val[length-keep:]
}

// printOptionValue 打印可选环境变量值
func printOptionValue() {
	for key, val := range optionValueMap {
		upperKey := strings.ToUpper(key)
		if strings.Contains(upperKey, "SECRET") || strings.Contains(upperKey, "PASSWORD") || strings.Contains(upperKey, "TOKEN") {
			logging.Config.Infof("环境变量配置项: %s, 值: %s", key, maskSecret(val))
		} else {
			logging.Config.Infof("环境变量配置项: %s, 值: %s", key, val)
		}
	}
}

// loadConfigFromEnv 从环境变量加载配置
func loadConfigFromEnv() {
	AppSetting.JwtSecret = getOptionEnvValue("JWT_SECRET", "")
	AppSetting.LogLevel = getOptionEnvValue("LOG_LEVEL", "INFO")

	ServerSetting.HttpPort = 8000
	ServerSetting.ReadTimeout = 60
	ServerSetting.WriteTimeout = 60
	ServerSetting.UrlPrefix = getOptionEnvValue("URL_PREFIX", "")

	DatabaseSetting.Type = getOptionEnvValue("DB_TYPE", "sqlite")
	DatabaseSetting.Ssl = getOptionEnvValue("SSL", "false")

	if DatabaseSetting.Type == "mysql" {
		DatabaseSetting.Host = getMustEnvValue("MYSQL_HOST")
		DatabaseSetting.Port = com.StrTo(getMustEnvValue("MYSQL_PORT")).MustInt()
		DatabaseSetting.User = getMustEnvValue("MYSQL_USER")
		DatabaseSetting.Password = getMustEnvValue("MYSQL_PASSWORD")
		DatabaseSetting.Name = getMustEnvValue("MYSQL_DB")
	}

	if DatabaseSetting.Type == "postgres" {
		DatabaseSetting.Host = getMustEnvValue("POSTGRES_HOST")
		DatabaseSetting.Port = com.StrTo(getMustEnvValue("POSTGRES_PORT")).MustInt()
		DatabaseSetting.User = getMustEnvValue("POSTGRES_USER")
		DatabaseSetting.Password = getMustEnvValue("POSTGRES_PASSWORD")
		DatabaseSetting.Name = getMustEnvValue("POSTGRES_DB")
	}

	DatabaseSetting.TablePrefix = getOptionEnvValue("MYSQL_TABLE_PREFIX", "message_")
	DatabaseSetting.SqlDebug = getOptionEnvValue("SQL_DEBUG", "disable")
	printOptionValue()
}
