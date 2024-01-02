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
func AppendCors(r *gin.Engine) {
	if setting.ServerSetting.RunMode == "debug" {
		r.Use(middleware.Cors())
	}
}

// ServerStaticHtml 启用返回打包的静态文件
func AppendServerStaticHtml(r *gin.Engine, f embed.FS) {
	if setting.ServerSetting.EmbedHtml == "disable" {
		return
	}

	assets, _ := fs.Sub(f, "web/dist/assets")
	dist, _ := fs.Sub(f, "web/dist")

	r.StaticFS("assets/", http.FS(assets))
	r.GET("/", func(ctx *gin.Context) {
		ctx.FileFromFS("/", http.FS(dist))
	})

}

// InitRouter initialize routing information
func InitRouter(f embed.FS) *gin.Engine {
	r := gin.New()
	r.Use(middleware.LogMiddleware())
	r.Use(gin.Recovery())

	AppendCors(r)
	AppendServerStaticHtml(r, f)

	r.POST("/auth", api.GetAuth)
	apiv1 := r.Group("/api/v1")
	apiv1.Use(middleware.JWT())
	{
		// sendways
		apiv1.POST("/sendways/add", v1.AddMsgSendWay)
		apiv1.POST("/sendways/delete", v1.DeleteMsgSendWay)
		apiv1.POST("/sendways/edit", v1.EditSendWay)
		apiv1.POST("/sendways/test", v1.TestSendWay)
		apiv1.GET("/sendways/list", v1.GetMsgSendWayList)
		apiv1.GET("/sendways/get", v1.GetMsgSendWay)

		// sendtasks
		apiv1.GET("/sendtasks/list", v1.GetMsgSendTaskList)
		apiv1.POST("/sendtasks/add", v1.AddMsgSendTask)
		apiv1.POST("/sendtasks/delete", v1.DeleteMsgSendTask)
		apiv1.POST("/sendtasks/edit", v1.EditMsgSendTask)

		// sendtasks/ins
		apiv1.POST("/sendtasks/ins/addmany", v1.AddManyTasksIns)
		apiv1.POST("/sendtasks/ins/addone", v1.AddTasksIns)
		apiv1.GET("/sendtasks/ins/gettask", v1.GetMsgSendWayIns)
		apiv1.POST("/sendtasks/ins/delete", v1.DeleteMsgTaskIns)

		// message/send
		apiv1.POST("/message/send", v1.DoSendMassage)

		apiv1.GET("/sendlogs/list", v1.GetTaskSendLogsList)

		// settings
		apiv1.POST("/settings/setpasswd", v1.EditPasswd)
		apiv1.POST("/settings/set", v1.EditSettings)
		apiv1.GET("/settings/getsetting", v1.GetUserSetting)

	}

	return r
}
