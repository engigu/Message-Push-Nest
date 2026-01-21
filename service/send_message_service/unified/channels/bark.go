package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type BarkChannel struct{ *BaseChannel }

func NewBarkChannel() *BarkChannel {
	return &BarkChannel{BaseChannel: NewBaseChannel(MessageTypeBark, []string{FormatTypeText})}
}

func (c *BarkChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailBark)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	_, ok = configInterface.(models.InsBarkConfig)
	if !ok {
		return "Bark config校验失败", ""
	}

	cli := message.Bark{
		PushKey: auth.PushKey,
		Archive: auth.Archive,
		Group:   auth.Group,
		Sound:   auth.Sound,
		Icon:    auth.Icon,
		Level:   auth.Level,
		URL:     auth.URL,
	}

	res, err := cli.Request(content.Title, content.Text)
	if err != nil {
		return string(res), fmt.Sprintf("发送失败：%s", err.Error())
	}
	return string(res), ""
}
