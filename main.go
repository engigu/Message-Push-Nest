package main

import (
	"fmt"
	"log"
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

// @title Golang Gin API
// @version 1.0
// @description An example of gin
// @termsOfService https://github.com/EDDYCJY/go-gin-example
// @license.name MIT
// @license.url https://message-nest/blob/master/LICENSE
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

	log.Printf("[info] start http server listening http://0.0.0.0%s", endPoint)

	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Server err: %v", err)
	}
}
