package send_message_service

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type WeChatOfAccountService struct {
}

// SendWeChatOfAccountMessage 执行发送微信公众号模板消息
func (s *WeChatOfAccountService) SendWeChatOfAccountMessage(auth send_way_service.WeChatOFAccount, ins models.SendTasksIns, typeC string, title string, content string, url string) (string, string) {
	insService := send_ins_service.SendTaskInsService{}
	errStr, c := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	config, ok := c.(models.InsWeChatAccountConfig)
	if !ok {
		return "微信公众号模板消息config校验失败", ""
	}
	var (
		err    error
		res    string
		errMsg string
	)
	cli := message.WeChatOFAccount{
		AppID:      auth.AppID,
		AppSecret:  auth.APPSecret,
		TemplateID: auth.TempID,
		ToUser:     config.ToAccount,
		URL:        url,
	}
	res, err = cli.Send(title, content)
	if err != nil {
		errMsg = fmt.Sprintf("发送失败：%s", ins.ContentType)
	}
	return res, errMsg
}
