package main

import (
	"embed"
	"fmt"
	"message-nest/migrate"
	"message-nest/models"
	"message-nest/pkg/constant"
	"message-nest/pkg/logging"
	"message-nest/pkg/setting"
	"message-nest/routers"
	"message-nest/service/cron_msg_service"
	"message-nest/service/cron_service"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	//go:embed web/dist/*
	f embed.FS

	//go:embed .release*
	rf embed.FS
)

func init() {
	constant.InitReleaseInfo(rf)
	setting.Setup()
	logging.Setup()
	models.Setup()
	go func() {
		// 完成model的迁移之后，需要加载异步任务
		migrate.Setup()
		cron_service.StartTasksRunOnStartup()
		// 加载用户自己设定的定时消息任务
		cron_msg_service.StartUpUserSetupMsgCronTask()
	}()
}

func main() {
	gin.SetMode(setting.ServerSetting.RunMode)
	routersInit := routers.InitRouter(f)
	readTimeout := setting.ServerSetting.ReadTimeout
	writeTimeout := setting.ServerSetting.WriteTimeout
	endPoint := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	startInfo := ""
	if setting.ServerSetting.RunMode == "debug" {
		startInfo = fmt.Sprintf("run mode: %s, start message server @ http://localhost%s", setting.ServerSetting.RunMode, endPoint)
	} else {
		startInfo = fmt.Sprintf("run mode: %s, start message server @ http://0.0.0.0%s", setting.ServerSetting.RunMode, endPoint)
	}

	logrus.WithFields(logrus.Fields{
		"prefix": fmt.Sprintf("[PID:%d]", os.Getpid()),
	}).Infof(startInfo)

	err := server.ListenAndServe()
	if err != nil {
		logrus.Errorf("Server err: ", err)
	}
}
