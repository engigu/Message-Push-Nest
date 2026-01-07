package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type FeishuChannel struct{ *BaseChannel }

func NewFeishuChannel() *FeishuChannel {
	return &FeishuChannel{BaseChannel: NewBaseChannel(MessageTypeFeishu, []string{FormatTypeMarkdown, FormatTypeText})}
}

func (c *FeishuChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailFeishu)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	_, ok = configInterface.(models.InsFeishuConfig)
	if !ok {
		return "飞书config校验失败", ""
	}
	contentType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}
	atMobiles := content.GetAtMobiles()
	atUserIds := content.GetAtUserIds()
	// 合并 @ 人员列表
	atList := append(atMobiles, atUserIds...)
	if content.IsAtAll() {
		atList = append(atList, "all")
	}
	cli := message.Feishu{AccessToken: auth.AccessToken, Secret: auth.Secret}
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
		errMsg = fmt.Sprintf("未知的飞书发送内容类型：%s", contentType)
	}
	return string(res), errMsg
}
