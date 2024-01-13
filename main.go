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
	"message-nest/service/send_message_service"
	"net/http"
	"sync"
)

var (
	//go:embed web/dist/*
	f embed.FS

	wg sync.WaitGroup
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

func GinServerUp() {
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

	logrus.Info("start http server listening http://0.0.0.0", endPoint)

	err := server.ListenAndServe()
	if err != nil {
		logrus.Error("Server err: ", err)
	}
}

func main() {
	wg.Add(1)

	go GinServerUp()
	go send_message_service.MessageConsumer(&wg)

	wg.Wait()
	fmt.Println("Server exit...")
}
