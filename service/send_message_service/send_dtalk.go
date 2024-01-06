package send_message_service

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type DtalkService struct {
}

// SendDtalkMessage 执行发送钉钉
func (s *DtalkService) SendDtalkMessage(auth send_way_service.WayDetailDTalk, ins models.SendTasksIns, typeC string, title string, content string) string {
	insService := send_ins_service.SendTaskInsService{}
	errStr, c := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr
	}
	_, ok := c.(models.InsDtalkConfig)
	if !ok {
		return "钉钉config校验失败"
	}

	errMsg := ""
	cli := message.Dtalk{
		AccessToken: auth.AccessToken,
		Secret:      auth.Secret,
	}
	if typeC == "text" {
		_, err := cli.SendMessageText(content)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", ins.ContentType)
		}
	} else if typeC == "markdown" {
		_, err := cli.SendMessageMarkdown(title, content)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", ins.ContentType)
		}
	} else {
		errMsg = fmt.Sprintf("未知的钉钉发送内容类型：%s", ins.ContentType)
	}
	return errMsg
}
