package send_message_service

import (
	"encoding/json"
	"fmt"

	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_way_service"
	"strings"
)

type CustomService struct {
}

// SendCustomMessage 执行发送钉钉
func (s *CustomService) SendCustomMessage(auth send_way_service.WayDetailCustom, ins models.SendTasksIns, typeC string, title string, content string) (string, string) {
	errMsg := ""
	cli := message.CustomWebhook{}
	data, _ := json.Marshal(content)
	dataStr := string(data)
	dataStr = strings.Trim(dataStr, "\"")
	bodyStr := strings.Replace(auth.Body, "TEXT", dataStr, -1)
	res, err := cli.Request(auth.Webhook, bodyStr)
	if err != nil {
		errMsg = fmt.Sprintf("发送失败：%s", err)
	}
	return string(res), errMsg
}
