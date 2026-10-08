package logging

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// ANSI 颜色码
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m" // error, fatal, panic
	colorYellow = "\033[33m" // warn
	colorBlue   = "\033[36m" // info
	colorGray   = "\033[37m" // debug
)

// CustomFormatter 统一实现紧凑彩色日志格式: [时间][LEVEL] [模块] 消息
type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	level := strings.ToUpper(entry.Level.String())

	var levelColor string
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = colorGray
	case logrus.InfoLevel:
		levelColor = colorBlue
	case logrus.WarnLevel:
		levelColor = colorYellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = colorRed
	default:
		levelColor = colorBlue
	}

	prefix := ""
	if p, ok := entry.Data["prefix"]; ok && p != nil {
		pStr := strings.TrimSpace(fmt.Sprintf("%v", p))
		if pStr != "" {
			pStr = strings.TrimPrefix(pStr, "[")
			pStr = strings.TrimSuffix(pStr, "]")
			pStr = strings.TrimSuffix(pStr, ":")
			pStr = strings.TrimSpace(pStr)
			if pStr != "" {
				prefix = "[" + pStr + "] "
			}
		}
	}

	msg := entry.Message
	// 如果 entry.Data 中没有 prefix，但 msg 本身是以 [xxx] 开头的，如 "[Init] 消息"
	if prefix == "" {
		if strings.HasPrefix(msg, "[") {
			idx := strings.Index(msg, "]")
			if idx > 1 {
				extractedPrefix := msg[1:idx]
				msg = strings.TrimSpace(msg[idx+1:])
				msg = strings.TrimPrefix(msg, ":")
				msg = strings.TrimSpace(msg)
				prefix = "[" + extractedPrefix + "] "
			}
		}
	}

	// 如果依然没有 prefix，默认赋予统一标识 [App]
	if prefix == "" {
		prefix = "[App] "
	}

	out := fmt.Sprintf("[%s]%s[%s]%s %s%s\n", timestamp, levelColor, level, colorReset, prefix, msg)
	return []byte(out), nil
}

// 预定义各功能模块的 Logger Entry，业务代码直接调用
var (
	App          *logrus.Entry
	Server       *logrus.Entry
	Config       *logrus.Entry
	Database     *logrus.Entry
	Init         *logrus.Entry
	Scheduler    *logrus.Entry
	HTTP         *logrus.Entry
	Security     *logrus.Entry
	Sender       *logrus.Entry
	PreCheck     *logrus.Entry
	CronMsg      *logrus.Entry
	SendInstance *logrus.Entry
	TemplateSend *logrus.Entry
)

func init() {
	initEntries()
}

func initEntries() {
	App = logrus.WithField("prefix", "App")
	Server = logrus.WithField("prefix", "Server")
	Config = logrus.WithField("prefix", "Config")
	Database = logrus.WithField("prefix", "Database")
	Init = logrus.WithField("prefix", "Init")
	Scheduler = logrus.WithField("prefix", "Scheduler")
	HTTP = logrus.WithField("prefix", "HTTP")
	Security = logrus.WithField("prefix", "Security")
	Sender = logrus.WithField("prefix", "Sender")
	PreCheck = logrus.WithField("prefix", "PreCheck")
	CronMsg = logrus.WithField("prefix", "CronMsg")
	SendInstance = logrus.WithField("prefix", "SendInstance")
	TemplateSend = logrus.WithField("prefix", "TemplateSend")
}

// WithPrefix 为动态模块创建带前缀标识的 Entry
func WithPrefix(prefix string) *logrus.Entry {
	return logrus.WithField("prefix", prefix)
}

func Setup() {
	logrus.SetFormatter(&CustomFormatter{})
	logrus.SetReportCaller(false)
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
	initEntries()
}

func SetLevel(levelStr string) {
	level := strings.ToLower(levelStr)
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


