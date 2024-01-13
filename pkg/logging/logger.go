package logging

import (
	"fmt"
	"github.com/engigu/logrus-prefixed-formatter"
	"github.com/sirupsen/logrus"
	"message-nest/pkg/setting"
	"os"
	"path/filepath"
	"strings"
)

var Logger = logrus.New()

func getFuncOutStr(funcStr string) string {
	parts := strings.Split(funcStr, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	} else {
		return ""
	}
}

func CallerFormatter(funcStr string, fileStr string) string {
	fName := filepath.Base(fileStr)
	println(fileStr, funcStr, fName)
	funcOut := getFuncOutStr(funcStr)
	return fmt.Sprintf(" [%s %s]", fName, funcOut)
}

func Setup() {

	formatter := new(prefixed.TextFormatter)
	formatter.DisableTimestamp = false
	formatter.ForceFormatting = false
	formatter.FullTimestamp = true
	formatter.CallerFormatter = CallerFormatter
	formatter.TimestampFormat = "2006-01-02 15:04:05.999"
	formatter.SetColorScheme(&prefixed.ColorScheme{
		TimestampStyle: "cyan",
		CallerStyle:    "cyan",
		PrefixStyle:    "green",
	})

	Logger.Formatter = formatter
	Logger.SetReportCaller(true)
	Logger.SetOutput(os.Stdout)

	level := strings.ToLower(setting.AppSetting.LogLevel)
	switch level {
	case "debug":
		Logger.SetLevel(logrus.DebugLevel)
	case "info":
		Logger.SetLevel(logrus.InfoLevel)
	case "warn":
		Logger.SetLevel(logrus.WarnLevel)
	case "error":
		Logger.SetLevel(logrus.ErrorLevel)
	default:
		Logger.SetLevel(logrus.DebugLevel)
	}
}
