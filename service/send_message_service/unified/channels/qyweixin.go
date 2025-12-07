package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type QyWeiXinChannel struct{ *BaseChannel }

func NewQyWeiXinChannel() *QyWeiXinChannel {
	return &QyWeiXinChannel{BaseChannel: NewBaseChannel(MessageTypeQyWeiXin, []string{FormatTypeMarkdown, FormatTypeText})}
}

func (c *QyWeiXinChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(send_way_service.WayDetailQyWeiXin)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	_, ok = configInterface.(models.InsQyWeiXinConfig)
	if !ok {
		return "企业微信config校验失败", ""
	}
	contentType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}
	atList := []string{}
	atList = append(atList, content.GetAtUserIds()...)
	atList = append(atList, content.GetAtMobiles()...)
	if content.IsAtAll() {
		atList = append(atList, "@all")
	}
	cli := message.QyWeiXin{AccessToken: auth.AccessToken}
	var res []byte
	var errMsg string
	if contentType == FormatTypeText {
		res, err = cli.SendMessageText(formattedContent, atList...)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else if contentType == FormatTypeMarkdown {
		res, err = cli.SendMessageMarkdown(content.Title, formattedContent, atList...)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else {
		errMsg = fmt.Sprintf("未知的企业微信发送内容类型：%s", contentType)
	}
	return string(res), errMsg
}
