package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type WeChatOFAccountChannel struct{ *BaseChannel }

func NewWeChatOFAccountChannel() *WeChatOFAccountChannel {
	return &WeChatOFAccountChannel{BaseChannel: NewBaseChannel(MessageTypeWeChatOFAccount, []string{FormatTypeText})}
}

func (c *WeChatOFAccountChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WeChatOFAccount)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	config, ok := configInterface.(models.InsWeChatAccountConfig)
	if !ok {
		return "微信公众号模板消息config校验失败", ""
	}
	_, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}
	cli := message.WeChatOFAccount{
		AppID:      auth.AppID,
		AppSecret:  auth.APPSecret,
		TemplateID: auth.TempID,
		ToUser:     config.ToAccount,
		URL:        content.URL,
	}
	res, err := cli.Send(content.Title, formattedContent)
	var errMsg string
	if err != nil {
		errMsg = fmt.Sprintf("发送失败：%s", err.Error())
	}
	return res, errMsg
}
