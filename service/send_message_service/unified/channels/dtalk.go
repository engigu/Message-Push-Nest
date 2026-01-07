package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type DtalkChannel struct{ *BaseChannel }

func NewDtalkChannel() *DtalkChannel {
	return &DtalkChannel{BaseChannel: NewBaseChannel(MessageTypeDtalk, []string{FormatTypeMarkdown, FormatTypeText})}
}

func (c *DtalkChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailDTalk)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	_, ok = configInterface.(models.InsDtalkConfig)
	if !ok {
		return "钉钉config校验失败", ""
	}
	contentType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}
	atMobiles := content.GetAtMobiles()
	if content.IsAtAll() {
		atMobiles = append(atMobiles, "all")
	}
	cli := message.Dtalk{AccessToken: auth.AccessToken, Secret: auth.Secret}
	var res []byte
	var errMsg string
	if contentType == FormatTypeText {
		res, err = cli.SendMessageText(formattedContent, atMobiles...)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else if contentType == FormatTypeMarkdown {
		res, err = cli.SendMessageMarkdown(content.Title, formattedContent, atMobiles...)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else {
		errMsg = fmt.Sprintf("未知的钉钉发送内容类型：%s", contentType)
	}
	return string(res), errMsg
}
