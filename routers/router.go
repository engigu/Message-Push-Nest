package routers

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/fs"
	"message-nest/middleware"
	"message-nest/pkg/setting"
	"message-nest/routers/api"
	"message-nest/routers/api/v1"
	"message-nest/routers/api/v2"
	"net/http"
	"strings"
)

// AppendCors 添加是否跨域（debug模式开启）
func AppendCors(app *gin.Engine) {
	if setting.ServerSetting.RunMode == "debug" {
		app.Use(middleware.Cors())
	}
}

// AppendServerStaticHtmlWithPrefix 启用返回打包的静态文件（支持路径前缀）
func AppendServerStaticHtmlWithPrefix(router gin.IRouter, f embed.FS, pathPrefix string) {
	if setting.ServerSetting.EmbedHtml == "disable" {
		return
	}

	assets, _ := fs.Sub(f, "web/dist/assets")
	dist, _ := fs.Sub(f, "web/dist")

	// 根据是否有路径前缀来设置静态文件路由
	if pathPrefix != "" {
		// 有路径前缀时，使用相对路径
		if r, ok := router.(*gin.RouterGroup); ok {
			r.Use(middleware.StaticCacheMiddleware())
			r.StaticFS("/assets", http.FS(assets))
			r.GET("/", func(ctx *gin.Context) {
				// 读取 index.html
				indexFile, err := dist.Open("index.html")
				if err != nil {
					ctx.String(http.StatusInternalServerError, "Failed to load index.html")
					return
				}
				defer indexFile.Close()

				// 读取文件内容
				content, err := io.ReadAll(indexFile)
				if err != nil {
					ctx.String(http.StatusInternalServerError, "Failed to read index.html")
					return
				}

				// 注入配置脚本
				configScript := fmt.Sprintf(`<script>window.__URL_PATH_PREFIX__ = '%s';</script>`, pathPrefix)
				htmlContent := string(content)
				// 在 </head> 标签前注入配置
				htmlContent = strings.Replace(htmlContent, "</head>", configScript+"</head>", 1)

				ctx.Header("Content-Type", "text/html; charset=utf-8")
				ctx.String(http.StatusOK, htmlContent)
			})
		}
	} else {
		// 无路径前缀时，使用原有逻辑
		if r, ok := router.(*gin.Engine); ok {
			r.Use(middleware.StaticCacheMiddleware())
			r.StaticFS("assets/", http.FS(assets))
			r.GET("/", func(ctx *gin.Context) {
				ctx.FileFromFS("/", http.FS(dist))
			})
		}
	}
}

// AppendServerStaticHtml 启用返回打包的静态文件（保留向后兼容）
func AppendServerStaticHtml(app *gin.Engine, f embed.FS) {
	AppendServerStaticHtmlWithPrefix(app, f, "")
}

// InitRouter 初始化路由
func InitRouter(f embed.FS) *gin.Engine {
	app := gin.New()
	app.Use(middleware.LogMiddleware())
	app.Use(gin.Recovery())

	AppendCors(app)
	
	// 获取 URL 前缀
	pathPrefix := setting.ServerSetting.UrlPrefix
	if pathPrefix != "" && pathPrefix[0] != '/' {
		pathPrefix = "/" + pathPrefix
	}
	
	// 如果有路径前缀，创建路由组
	var router gin.IRouter
	if pathPrefix != "" {
		router = app.Group(pathPrefix)
	} else {
		router = app
	}
	
	AppendServerStaticHtmlWithPrefix(router, f, pathPrefix)

	router.POST("/auth", api.GetAuth)
	apiV1 := router.Group("/api/v1")
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

		// login logs
		apiV1.GET("/loginlogs/recent", v1.GetRecentLoginLogs)

		// statistic
		apiV1.GET("/statistic", v1.GetStatisticData)
		apiV1.GET("/statistic/task", v1.GetSendStatsByTask)

		// cronMessage
		apiV1.POST("/cronmessages/addone", v1.AddCronMsgTask)
		apiV1.GET("/cronmessages/list", v1.GetCronMsgList)
		apiV1.POST("/cronmessages/delete", v1.DeleteCronMsgTask)
		apiV1.POST("/cronmessages/edit", v1.EditCronMsgTask)
		apiV1.POST("/cronmessages/sendnow", v1.SendNowCronMsg)

		// hostedMessage
		apiV1.GET("/hostedmessages/list", v1.GetHostMessageList)

		// messageTemplate
		apiV1.GET("/templates/list", v1.GetMessageTemplateList)
		apiV1.GET("/templates/get", v1.GetMessageTemplate)
		apiV1.POST("/templates/add", v1.AddMessageTemplate)
		apiV1.POST("/templates/edit", v1.EditMessageTemplate)
		apiV1.POST("/templates/delete", v1.DeleteMessageTemplate)
		apiV1.POST("/templates/preview", v1.PreviewMessageTemplate)
		
		// messageTemplate instances
		apiV1.GET("/templates/ins/get", v1.GetTemplateWithIns)
		apiV1.POST("/templates/ins/addone", v1.AddTemplateIns)

	}

	// API v2
	apiV2 := router.Group("/api/v2")
	apiV2.Use(middleware.JWT())
	{
		// message/send - 使用模板发送消息
		apiV2.POST("/message/send", v2.DoSendMessageByTemplate)
	}

	return app
}
