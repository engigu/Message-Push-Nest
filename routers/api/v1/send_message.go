package v1

import (
	"fmt"
	"message-nest/pkg/app"
	"message-nest/pkg/e"
	utilpkg "message-nest/pkg/util"
	"message-nest/service/send_message_service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SendMessageReq struct {
	// 兼容旧参数：task_id；不再必填
	TaskID string `json:"task_id" validate:"omitempty,len=12" label:"任务id"`
	// 新参数：token（与task_id互斥，优先使用task_id）
	Token    string `json:"token" label:"任务token"`
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

	// 解析token为task_id（如提供）
	taskID := req.TaskID
	// 如果使用了token就使用token解析出task_id
	if req.Token != "" {
		dec, err := utilpkg.DecryptTokenHex(req.Token, 71) // 71 为简单对称密钥
		if err != nil {
			appG.CResponse(http.StatusBadRequest, fmt.Sprintf("token解析失败：%v", err), nil)
			return
		}
		taskID = dec
	}

	if taskID == "" {
		appG.CResponse(http.StatusBadRequest, "参数缺失：token 或 task_id 必须提供其一", nil)
		return
	}

	msgService := send_message_service.SendMessageService{
		TaskID:   taskID,
		Title:    req.Title,
		Text:     req.Text,
		HTML:     req.HTML,
		URL:      req.URL,
		MarkDown: req.MarkDown,
		CallerIp: c.ClientIP(),
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
