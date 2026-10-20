package bootstrap

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"message-nest/middleware"
	"message-nest/migrate"
	"message-nest/models"
	"message-nest/pkg/constant"
	"message-nest/pkg/logging"
	"message-nest/pkg/setting"
	"message-nest/routers"
	"message-nest/routers/api"
	"message-nest/service/cron_msg_service"
	"message-nest/service/cron_service"
)

// App 应用启动引导器 (参考白狐架构)
type App struct {
	WebFS     embed.FS
	ReleaseFS embed.FS
	Router    *gin.Engine
	Server    *http.Server
}

// New 初始化核心基础组件
func New(webFS, releaseFS embed.FS) *App {
	app := &App{
		WebFS:     webFS,
		ReleaseFS: releaseFS,
	}

	app.initConfig()
	app.initDatabase()
	app.initScheduler()
	app.initRouter()

	return app
}

// initConfig 初始化全局配置、日志与版本信息
func (a *App) initConfig() {
	logging.Setup()
	constant.InitReleaseInfo(a.ReleaseFS)
	setting.Setup()
	logging.SetLevel(setting.AppSetting.LogLevel)
}

// initDatabase 初始化数据库连接池
func (a *App) initDatabase() {
	models.Setup()
}

// initScheduler 初始化后台异步任务与定时调度器
func (a *App) initScheduler() {
	go func() {
		// 完成 model 的迁移与管理员账号初始化
		migrate.Setup()

		// 启动系统级清理定时任务 (日志与托管消息)
		cron_service.StartTasksRunOnStartup()

		// 加载用户自定义的定时消息任务
		cron_msg_service.StartUpUserSetupMsgCronTask()

		// 统一使用系统内置 cron 调度系统级内存防护清理任务 (每10分钟清理一次过期缓存，严格控制内存开销)
		cron_service.AddTask(cron_service.ScheduledTask{
			Schedule: "*/10 * * * *",
			Job: func() {
				middleware.CleanAllRateLimiters()
				api.CleanExpiredLoginAttempts()
			},
		})
	}()
}

// initRouter 初始化 Gin 路由与 HTTP Server 实例
func (a *App) initRouter() {
	gin.SetMode(gin.ReleaseMode)
	a.Router = routers.InitRouter(a.WebFS)

	readTimeout := setting.ServerSetting.ReadTimeout
	writeTimeout := setting.ServerSetting.WriteTimeout
	endPoint := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)
	maxHeaderBytes := 1 << 20

	a.Server = &http.Server{
		Addr:           endPoint,
		Handler:        a.Router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
}

// Run 启动 HTTP 监听服务
func (a *App) Run() {
	endPoint := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)
	logging.Server.Infof("服务启动地址: http://0.0.0.0%s", endPoint)

	err := a.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logging.Server.Errorf("服务监听异常: %v", err)
	}
}
