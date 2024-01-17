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

//var Logger = logrus.New()

func getFuncOutStr(funcStr string) string {
	parts := strings.Split(funcStr, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	} else {
		return ""
	}
}

func CustomCallerFormatter(funcStr string, fileStr string) string {
	fName := filepath.Base(fileStr)
	funcOut := getFuncOutStr(funcStr)
	return fmt.Sprintf(" [%s %s]", fName, funcOut)
}

func Setup() {

	formatter := new(prefixed.TextFormatter)
	formatter.DisableTimestamp = false
	formatter.DisableColors = false
	formatter.ForceColors = true
	formatter.ForceFormatting = true
	formatter.FullTimestamp = true
	formatter.CallerFormatter = CustomCallerFormatter
	formatter.TimestampFormat = "2006-01-02 15:04:05.000"
	formatter.SetColorScheme(&prefixed.ColorScheme{
		TimestampStyle: "cyan",
		CallerStyle:    "cyan",
		PrefixStyle:    "green",
	})

	logrus.SetFormatter(formatter)
	logrus.SetReportCaller(true)
	logrus.SetOutput(os.Stdout)

	level := strings.ToLower(setting.AppSetting.LogLevel)
	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
}
