package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type EmailChannel struct{ *BaseChannel }

func NewEmailChannel() *EmailChannel {
	return &EmailChannel{BaseChannel: NewBaseChannel(MessageTypeEmail, []string{FormatTypeHTML, FormatTypeText})}
}

func (c *EmailChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailEmail)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return "", errStr
	}
	config, ok := configInterface.(models.InsEmailConfig)
	if !ok {
		return "", "邮箱config校验失败"
	}

	if config.ToAccount == "" {
		return "", "收件邮箱地址为空，请检查实例配置或启用动态接收模式"
	}

	contentType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}
	var emailer message.EmailMessage
	emailer.Init(auth.Server, auth.Port, auth.Account, auth.Passwd, auth.FromName)
	var errMsg string
	if contentType == FormatTypeText {
		errMsg = emailer.SendTextMessage(config.ToAccount, content.Title, formattedContent)
	} else if contentType == FormatTypeHTML {
		errMsg = emailer.SendHtmlMessage(config.ToAccount, content.Title, formattedContent)
	} else {
		errMsg = fmt.Sprintf("未知的邮件发送内容类型：%s", contentType)
	}
	return "", errMsg
}
