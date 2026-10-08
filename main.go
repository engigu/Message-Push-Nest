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

	"github.com/gin-gonic/gin"
)

var (
	//go:embed web/dist/*
	f embed.FS

	//go:embed .release*
	rf embed.FS
)

func init() {
	logging.Setup()
	constant.InitReleaseInfo(rf)
	setting.Setup()
	logging.SetLevel(setting.AppSetting.LogLevel)
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
	gin.SetMode(gin.ReleaseMode)
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

	logging.Server.Infof("服务启动地址: http://0.0.0.0%s", endPoint)

	err := server.ListenAndServe()
	if err != nil {
		logging.Server.Errorf("服务监听异常: %v", err)
	}
}
