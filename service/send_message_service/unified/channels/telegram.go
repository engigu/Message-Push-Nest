package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type TelegramChannel struct{ *BaseChannel }

func NewTelegramChannel() *TelegramChannel {
	return &TelegramChannel{BaseChannel: NewBaseChannel(MessageTypeTelegram, []string{FormatTypeMarkdown, FormatTypeHTML, FormatTypeText})}
}

func (c *TelegramChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(send_way_service.WayDetailTelegram)
	if !ok {
		return "", "类型转换失败"
	}
	
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	
	_, ok = configInterface.(models.InsTelegramConfig)
	if !ok {
		return "Telegram config校验失败", ""
	}
	
	contentType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}
	
	cli := message.Telegram{
		BotToken: auth.BotToken,
		ChatID:   auth.ChatID,
		ApiHost:  auth.ApiHost,
		ProxyURL: auth.ProxyURL,
	}
	
	var res []byte
	var errMsg string
	
	if contentType == FormatTypeText {
		res, err = cli.SendMessageText(formattedContent)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else if contentType == FormatTypeMarkdown {
		res, err = cli.SendMessageMarkdown(formattedContent)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else if contentType == FormatTypeHTML {
		res, err = cli.SendMessageHTML(formattedContent)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else {
		errMsg = fmt.Sprintf("未知的Telegram发送内容类型：%s", contentType)
	}
	
	return string(res), errMsg
}
