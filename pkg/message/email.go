package message

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type EmailMessage struct {
	Server   string
	Port     int
	Account  string
	Passwd   string
	FromName string
	GM       *gomail.Dialer
}

func (e *EmailMessage) Init(host string, port int, account string, passwd string, fromName string) {
	e.Server = host
	e.Port = port
	e.Account = account
	e.Passwd = passwd
	e.FromName = fromName
	e.GM = gomail.NewDialer(host, port, account, passwd)
}

func (e *EmailMessage) SendTextMessage(toEmail string, title string, content string) string {
	m := gomail.NewMessage()
	if e.FromName != "" {
		m.SetAddressHeader("From", e.Account, e.FromName)
	} else {
		m.SetHeader("From", e.Account)
	}
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", title)
	m.SetBody("text/html", content)

	if err := e.GM.DialAndSend(m); err != nil {
		return fmt.Sprintf("邮件发送失败: %s", err)
	}
	return ""
}

func (e *EmailMessage) SendHtmlMessage(toEmail string, title string, content string) string {
	return e.SendTextMessage(toEmail, title, content)
}
