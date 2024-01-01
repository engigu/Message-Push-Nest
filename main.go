package main

import (
	"fmt"
	"message-nest/pkg/table"
	"message-nest/service/cron_service"
	"message-nest/service/env_service"
	"net/http"

	"github.com/gin-gonic/gin"

	"message-nest/models"
	"message-nest/pkg/logging"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"message-nest/routers"
)

func init() {
	setting.Setup()
	models.Setup()
	logging.Setup()
	util.Setup()
	table.Setup()
	env_service.Setup()
	cron_service.Setup()
}

func main() {
	gin.SetMode(setting.ServerSetting.RunMode)

	routersInit := routers.InitRouter()
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

	logging.Logger.Info("start http server listening http://0.0.0.0", endPoint)

	err := server.ListenAndServe()
	if err != nil {
		logging.Logger.Error("Server err: ", err)
	}
}
