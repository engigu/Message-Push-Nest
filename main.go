package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"message-nest/migrate"
	"message-nest/models"
	"message-nest/pkg/logging"

	"message-nest/pkg/setting"
	"message-nest/pkg/table"
	"message-nest/routers"
	"message-nest/service/cron_service"
	"message-nest/service/env_service"
	"net/http"
)

var (
	//go:embed web/dist/*
	f embed.FS
)

func init() {
	setting.Setup()
	logging.Setup()
	migrate.Setup()
	models.Setup()
	table.Setup()
	env_service.Setup()
	cron_service.Setup()
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

	logrus.Infof("start message server listening http://0.0.0.0%s", endPoint)

	err := server.ListenAndServe()
	if err != nil {
		logrus.Error("Server err: ", err)
	}
}
