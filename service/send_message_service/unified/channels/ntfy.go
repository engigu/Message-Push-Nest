package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type NtfyChannel struct{ *BaseChannel }

func NewNtfyChannel() *NtfyChannel {
	return &NtfyChannel{BaseChannel: NewBaseChannel(MessageTypeNtfy, []string{FormatTypeText})}
}

func (c *NtfyChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailNtfy)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	_, ok = configInterface.(models.InsNtfyConfig)
	if !ok {
		return "Ntfy config校验失败", ""
	}

	cli := message.Ntfy{
		Url:      auth.Url,
		Topic:    auth.Topic,
		Priority: auth.Priority,
		Icon:     auth.Icon,
		Token:    auth.Token,
		Username: auth.Username,
		Password: auth.Password,
		Actions:  auth.Actions,
	}

	res, err := cli.Request(content.Title, content.Text)
	if err != nil {
		return string(res), fmt.Sprintf("发送失败：%s", err.Error())
	}
	return string(res), ""
}
