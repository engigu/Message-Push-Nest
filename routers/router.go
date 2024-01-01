package routers

import (
	"github.com/gin-gonic/gin"
	"message-nest/middleware"
	"message-nest/routers/api"
	"message-nest/routers/api/v1"
)

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(middleware.LogMiddleware())
	//gin.DefaultWriter = logging.Logger.Out
	//gin.DefaultErrorWriter = logging.Logger.Out
	r.Use(gin.Recovery())
	r.Use(middleware.Cors())

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
