package send_message_service

import (
	"message-nest/models"
	"message-nest/service/hosted_message_service"
	"message-nest/service/send_way_service"
)

type HostMessageService struct {
}

// SendHostMessage 执行托管消息记录
func (s *HostMessageService) SendHostMessage(
	auth send_way_service.MessageNest,
	ins models.SendTasksIns,
	typeC string,
	title string,
	content string) (string, string) {

	errMsg := ""
	var res string
	var err error
	messageService := hosted_message_service.HostMessageService{
		Title:   title,
		Content: content,
		Type:    typeC,
	}
	err = messageService.Add()
	if err != nil {
		errMsg = err.Error()
		res = "托管消息创建失败！"
	} else {
		res = "托管消息创建成功！"
	}
	return string(res), errMsg
}
