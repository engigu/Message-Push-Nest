package logging

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCustomFormatter(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetFormatter(&CustomFormatter{})
	logger.SetLevel(logrus.DebugLevel)

	tests := []struct {
		name     string
		logFunc  func()
		contains []string
	}{
		{
			name: "Default prefix App",
			logFunc: func() {
				logger.Info("系统服务启动")
			},
			contains: []string{"[INFO]", "[App]", "系统服务启动"},
		},
		{
			name: "Explicit field prefix",
			logFunc: func() {
				logger.WithField("prefix", "Scheduler").Infof("定时清理任务已启动")
			},
			contains: []string{"[INFO]", "[Scheduler]", "定时清理任务已启动"},
		},
		{
			name: "Strip brackets and colon in prefix field",
			logFunc: func() {
				logger.WithField("prefix", "[Init]:").Infof("数据表初始化完成")
			},
			contains: []string{"[INFO]", "[Init]", "数据表初始化完成"},
		},
		{
			name: "Auto parse message bracket prefix",
			logFunc: func() {
				logger.Infof("[Database] 数据库迁移成功")
			},
			contains: []string{"[INFO]", "[Database]", "数据库迁移成功"},
		},
		{
			name: "Auto parse message bracket prefix with colon",
			logFunc: func() {
				logger.Infof("[HTTP]: 服务监听端口 :8000")
			},
			contains: []string{"[INFO]", "[HTTP]", "服务监听端口 :8000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()
			for _, item := range tt.contains {
				if !strings.Contains(output, item) {
					t.Errorf("expected output to contain %q, but got:\n%s", item, output)
				}
			}
			// 确保没有多余的尾随冒号如 "[Init]:"
			if strings.Contains(output, "]: ") {
				t.Errorf("unexpected trailing colon after prefix tag in:\n%s", output)
			}
		})
	}
}
