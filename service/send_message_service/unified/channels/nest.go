package channels

import (
	"message-nest/models"
	"message-nest/service/hosted_message_service"
	"message-nest/service/send_way_service"
)

type MessageNestChannel struct{ *BaseChannel }

func NewMessageNestChannel() *MessageNestChannel {
	return &MessageNestChannel{BaseChannel: NewBaseChannel(MessageTypeMessageNest, []string{FormatTypeMarkdown, FormatTypeHTML, FormatTypeText})}
}

func (c *MessageNestChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	_, ok := msgObj.(*send_way_service.MessageNest)
	if !ok {
		return "", "类型转换失败"
	}
	contentType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}
	messageService := hosted_message_service.HostMessageService{
		Title:   content.Title,
		Content: formattedContent,
		Type:    contentType,
	}
	err = messageService.Add()
	var res, errMsg string
	if err != nil {
		errMsg = err.Error()
		res = "托管消息创建失败！"
	} else {
		res = "托管消息创建成功！"
	}
	return res, errMsg
}
