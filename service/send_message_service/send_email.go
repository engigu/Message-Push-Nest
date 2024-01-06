package send_message_service

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type EmailService struct {
}

// SendTaskEmail 执行发送邮件
func (s *EmailService) SendTaskEmail(auth send_way_service.WayDetailEmail, ins models.SendTasksIns, typeC string, title string, content string) string {
	insService := send_ins_service.SendTaskInsService{}
	errStr, c := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr
	}
	config, ok := c.(models.InsEmailConfig)
	if !ok {
		return "邮箱config校验失败"
	}

	var emailer message.EmailMessage
	errMsg := ""
	emailer.Init(auth.Server, auth.Port, auth.Account, auth.Passwd)
	if typeC == "text" {
		errMsg = emailer.SendTextMessage(config.ToAccount, title, content)
	} else if typeC == "html" {
		errMsg = emailer.SendHtmlMessage(config.ToAccount, title, content)
	} else {
		errMsg = fmt.Sprintf("未知的邮件发送内容类型：%s", ins.ContentType)
	}
	return errMsg
}
