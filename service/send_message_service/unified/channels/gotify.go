package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type GotifyChannel struct{ *BaseChannel }

func NewGotifyChannel() *GotifyChannel {
	return &GotifyChannel{BaseChannel: NewBaseChannel(MessageTypeGotify, []string{FormatTypeText})}
}

func (c *GotifyChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailGotify)
	if !ok {
		return "", "类型转换失败"
	}
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	_, ok = configInterface.(models.InsGotifyConfig)
	if !ok {
		return "Gotify config校验失败", ""
	}

	cli := message.Gotify{
		Url:      auth.Url,
		Token:    auth.Token,
		Priority: auth.Priority,
	}

	res, err := cli.Request(content.Title, content.Text)
	if err != nil {
		return string(res), fmt.Sprintf("发送失败：%s", err.Error())
	}
	return string(res), ""
}
