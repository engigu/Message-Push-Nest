package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	"message-nest/service/send_message_service"
	"net/http"
)

type SendMessageReq struct {
	TaskID   string `json:"task_id" validate:"required,len=12" label:"任务id"`
	Text     string `json:"text" validate:"required" label:"文本内容"`
	Title    string `json:"title"  label:"消息标题"`
	HTML     string `json:"html"  label:"html内容"`
	URL      string `json:"url"  label:"消息详情url地址"`
	MarkDown string `json:"markdown" label:"markdown内容"`
	Mode     string `json:"mode" label:"是否异步发送"`
}

// DoSendMassage 外部调用发信接口
func DoSendMassage(c *gin.Context) {
	var (
		appG = app.Gin{C: c}
		req  SendMessageReq
	)

	errCode, errMsg := app.BindJsonAndPlayValid(c, &req)
	if errCode != e.SUCCESS {
		appG.CResponse(errCode, errMsg, nil)
		return
	}

	msgService := send_message_service.SendMessageService{
		TaskID:        req.TaskID,
		Title:         req.Title,
		Text:          req.Text,
		HTML:          req.HTML,
		URL:           req.URL,
		MarkDown:      req.MarkDown,
		CallerIp:      c.ClientIP(),
		DefaultLogger: logrus.WithFields(logrus.Fields{
			//"prefix": "[Message Instance]",
		}),
	}
	task, err := msgService.SendPreCheck()
	if err != nil {
		appG.CResponse(http.StatusBadRequest, fmt.Sprintf("发送检查不通过：%s", err), nil)
		return
	}
	if req.Mode == "sync" {
		// 同步发送
		_, err := msgService.Send(task)
		if err != nil {
			appG.CResponse(http.StatusBadRequest, "发送失败！", nil)
			return
		}
		appG.CResponse(http.StatusOK, "发送成功！", nil)
	} else {
		// 异步发送
		msgService.AsyncSend(task)
		appG.CResponse(http.StatusOK, "提交成功！", nil)
	}
}
