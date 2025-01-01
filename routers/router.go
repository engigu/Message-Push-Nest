package routers

import (
	"embed"
	"github.com/gin-gonic/gin"
	"io/fs"
	"message-nest/middleware"
	"message-nest/pkg/setting"
	"message-nest/routers/api"
	"message-nest/routers/api/v1"
	"net/http"
)

// AppendCors 添加是否跨域（debug模式开启）
func AppendCors(app *gin.Engine) {
	if setting.ServerSetting.RunMode == "debug" {
		app.Use(middleware.Cors())
	}
}

// AppendServerStaticHtml 启用返回打包的静态文件
func AppendServerStaticHtml(app *gin.Engine, f embed.FS) {
	if setting.ServerSetting.EmbedHtml == "disable" {
		return
	}

	app.Use(middleware.StaticCacheMiddleware())

	assets, _ := fs.Sub(f, "web/dist/assets")
	dist, _ := fs.Sub(f, "web/dist")

	app.StaticFS("assets/", http.FS(assets))
	app.GET("/", func(ctx *gin.Context) {
		ctx.FileFromFS("/", http.FS(dist))
	})

}

// InitRouter 初始化路由
func InitRouter(f embed.FS) *gin.Engine {
	app := gin.New()
	app.Use(middleware.LogMiddleware())
	app.Use(gin.Recovery())

	AppendCors(app)
	AppendServerStaticHtml(app, f)

	app.POST("/auth", api.GetAuth)
	apiV1 := app.Group("/api/v1")
	apiV1.Use(middleware.JWT())
	{
		// sendways
		apiV1.POST("/sendways/add", v1.AddMsgSendWay)
		apiV1.POST("/sendways/delete", v1.DeleteMsgSendWay)
		apiV1.POST("/sendways/edit", v1.EditSendWay)
		apiV1.POST("/sendways/test", v1.TestSendWay)
		apiV1.GET("/sendways/list", v1.GetMsgSendWayList)
		apiV1.GET("/sendways/get", v1.GetMsgSendWay)

		// sendtasks
		apiV1.GET("/sendtasks/list", v1.GetMsgSendTaskList)
		apiV1.POST("/sendtasks/add", v1.AddMsgSendTask)
		apiV1.POST("/sendtasks/delete", v1.DeleteMsgSendTask)
		apiV1.POST("/sendtasks/edit", v1.EditMsgSendTask)
		apiV1.GET("/sendtasks/get", v1.GetMsgSendTask)

		// sendtasks/ins
		apiV1.POST("/sendtasks/ins/addmany", v1.AddManyTasksIns)
		apiV1.POST("/sendtasks/ins/addone", v1.AddTasksIns)
		apiV1.GET("/sendtasks/ins/gettask", v1.GetMsgSendWayIns)
		apiV1.POST("/sendtasks/ins/delete", v1.DeleteMsgTaskIns)
		apiV1.POST("/sendtasks/ins/update_enable", v1.UpdateMsgTaskInsEnable)

		// message/send
		apiV1.POST("/message/send", v1.DoSendMassage)

		apiV1.GET("/sendlogs/list", v1.GetTaskSendLogsList)

		// settings
		apiV1.POST("/settings/setpasswd", v1.EditPasswd)
		apiV1.POST("/settings/set", v1.EditSettings)
		apiV1.POST("/settings/reset", v1.RestDefaultSettings)
		apiV1.GET("/settings/getsetting", v1.GetUserSetting)

		// statistic
		apiV1.GET("/statistic", v1.GetStatisticData)

		// cronMessage
		apiV1.POST("/cronmessages/addone", v1.AddCronMsgTask)
		apiV1.GET("/cronmessages/list", v1.GetCronMsgList)
		apiV1.POST("/cronmessages/delete", v1.DeleteCronMsgTask)
		apiV1.POST("/cronmessages/edit", v1.EditCronMsgTask)

		// hostedMessage
		apiV1.GET("/hostedmessages/list", v1.GetHostMessageList)

	}

	return app
}
