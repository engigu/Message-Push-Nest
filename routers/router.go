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

// AppendCors 启用跨域中间件支持
func AppendCors(app *gin.Engine) {
	// app.Use(middleware.Cors())
}

// AppendServerStaticHtmlWithPrefix 启用返回打包的静态文件（支持路径前缀）
func AppendServerStaticHtmlWithPrefix(router gin.IRouter, f embed.FS, pathPrefix string) {
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

				htmlContent := string(content)
				
				// 注入 base 标签和配置脚本
				// base 标签必须在 head 的最前面，确保所有相对路径都基于这个 base
				baseTag := fmt.Sprintf(`<base href="%s/">`, pathPrefix)
				configScript := fmt.Sprintf(`<script>window.__URL_PATH_PREFIX__ = '%s';</script>`, pathPrefix)
				
				// 在 <head> 标签后立即插入 base 标签
				htmlContent = strings.Replace(htmlContent, "<head>", "<head>"+baseTag, 1)
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

	// API v1
	apiV1 := router.Group("/api/v1")
	{
		// 公开免鉴权接口
		apiV1.POST("/auth", api.GetAuth)
		apiV1.GET("/hostedmessages/preview", v1.GetHostMessagePreview)

		// 业务接口（需要 JWT 鉴权）
		authGroup := apiV1.Group("")
		authGroup.Use(middleware.JWT())
		{
			// sendways
			authGroup.POST("/sendways/add", v1.AddMsgSendWay)
			authGroup.POST("/sendways/delete", v1.DeleteMsgSendWay)
			authGroup.POST("/sendways/edit", v1.EditSendWay)
			authGroup.POST("/sendways/test", v1.TestSendWay)
			authGroup.GET("/sendways/list", v1.GetMsgSendWayList)
			authGroup.GET("/sendways/get", v1.GetMsgSendWay)

			// sendtasks
			authGroup.GET("/sendtasks/list", v1.GetMsgSendTaskList)
			authGroup.POST("/sendtasks/add", v1.AddMsgSendTask)
			authGroup.POST("/sendtasks/delete", v1.DeleteMsgSendTask)
			authGroup.POST("/sendtasks/edit", v1.EditMsgSendTask)
			authGroup.GET("/sendtasks/get", v1.GetMsgSendTask)

			// sendtasks/ins
			authGroup.POST("/sendtasks/ins/addmany", v1.AddManyTasksIns)
			authGroup.POST("/sendtasks/ins/addone", v1.AddTasksIns)
			authGroup.GET("/sendtasks/ins/gettask", v1.GetMsgSendWayIns)
			authGroup.POST("/sendtasks/ins/delete", v1.DeleteMsgTaskIns)
			authGroup.POST("/sendtasks/ins/update_enable", v1.UpdateMsgTaskInsEnable)

			// message/send
			authGroup.POST("/message/send", v1.DoSendMassage)

			authGroup.GET("/sendlogs/list", v1.GetTaskSendLogsList)

			// settings
			authGroup.POST("/settings/setpasswd", v1.EditPasswd)
			authGroup.POST("/settings/set", v1.EditSettings)
			authGroup.POST("/settings/reset", v1.RestDefaultSettings)
			authGroup.GET("/settings/getsetting", v1.GetUserSetting)

			// login logs
			authGroup.GET("/loginlogs/recent", v1.GetRecentLoginLogs)

			// statistic
			authGroup.GET("/statistic", v1.GetStatisticData)
			authGroup.GET("/statistic/task", v1.GetSendStatsByTask)

			// cronMessage
			authGroup.POST("/cronmessages/addone", v1.AddCronMsgTask)
			authGroup.GET("/cronmessages/list", v1.GetCronMsgList)
			authGroup.POST("/cronmessages/delete", v1.DeleteCronMsgTask)
			authGroup.POST("/cronmessages/edit", v1.EditCronMsgTask)
			authGroup.POST("/cronmessages/sendnow", v1.SendNowCronMsg)

			// hostedMessage
			authGroup.GET("/hostedmessages/list", v1.GetHostMessageList)

			// messageTemplate
			authGroup.GET("/templates/list", v1.GetMessageTemplateList)
			authGroup.GET("/templates/get", v1.GetMessageTemplate)
			authGroup.POST("/templates/add", v1.AddMessageTemplate)
			authGroup.POST("/templates/edit", v1.EditMessageTemplate)
			authGroup.POST("/templates/delete", v1.DeleteMessageTemplate)
			authGroup.POST("/templates/preview", v1.PreviewMessageTemplate)
			
			// messageTemplate instances
			authGroup.GET("/templates/ins/get", v1.GetTemplateWithIns)
			authGroup.POST("/templates/ins/addone", v1.AddTemplateIns)
		}
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
