package send_way_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/message"
	"strings"
)

type SendWay struct {
	ID         string
	Name       string
	Type       string
	CreatedBy  string
	ModifiedBy string
	Auth       string
	CreatedOn  string

	PageNum  int
	PageSize int
}

// WayDetailEmail 邮箱渠道明细字段
type WayDetailEmail struct {
	Server  string `validate:"required,max=50" label:"SMTP服务地址"`
	Port    int    `validate:"required,max=65535" label:"SMTP服务端口"`
	Account string `validate:"required,email" label:"邮箱账号"`
	Passwd  string `validate:"required,max=50" label:"邮箱密码"`
}

// WayDetailDTalk 钉钉渠道明细字段
type WayDetailDTalk struct {
	WebhookUrl string `validate:"required,url" label:"钉钉webhookUrl地址"`
}

func (sw *SendWay) GetByID() (interface{}, error) {
	return models.GetWayByID(sw.ID)
}

func (sw *SendWay) Add() error {
	return models.AddSendWay(sw.Name, sw.Auth, sw.Type, sw.CreatedBy, sw.ModifiedBy)
}

func (sw *SendWay) Edit() error {
	data := make(map[string]interface{})
	data["modified_by"] = sw.ModifiedBy
	data["name"] = sw.Name
	data["auth"] = sw.Auth
	return models.EditSendWay(sw.ID, data)
}

func (sw *SendWay) Delete() error {
	tasks := models.FindTaskByWayId(sw.ID)
	if len(tasks) > 0 {
		var names []string
		for _, task := range tasks {
			names = append(names, task.Name)
		}
		return errors.New(fmt.Sprintf("已经存在使用的任务，删除失败！任务名：%s", strings.Join(names, ", ")))
	}
	return models.DeleteMsgWay(sw.ID)
}

func (sw *SendWay) Count() (int, error) {
	return models.GetSendWaysTotal(sw.Name, sw.Type, sw.getMaps())
}

func (sw *SendWay) GetAll() ([]models.SendWays, error) {
	tags, err := models.GetSendWays(sw.PageNum, sw.PageSize, sw.Name, sw.Type, sw.getMaps())
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (sw *SendWay) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	return maps
}

// ValidateDiffWay 各种发信渠道具体字段校验
func (sw *SendWay) ValidateDiffWay() (string, interface{}) {
	var empty interface{}
	if sw.Type == "Email" {
		var email WayDetailEmail
		err := json.Unmarshal([]byte(sw.Auth), &email)
		if err != nil {
			return "邮箱auth反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(email)
		return Msg, email
	} else if sw.Type == "Dtalk" {
		var dtalk WayDetailDTalk
		err := json.Unmarshal([]byte(sw.Auth), &dtalk)
		if err != nil {
			return "钉钉auth反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(dtalk)
		return Msg, dtalk
	}
	return fmt.Sprintf("未知的发信渠道校验: %s", sw.Type), empty
}

// TestSendWay 尝试带发信测试连通性
func (sw *SendWay) TestSendWay(msgObj interface{}) string {
	emailAuth, ok := msgObj.(WayDetailEmail)
	if ok {
		var emailer message.EmailMessage
		emailer.Init(emailAuth.Server, 465, emailAuth.Account, emailAuth.Passwd)
		errMsg := emailer.SendTextMessage(emailAuth.Account, "test", "This is a test email from message-nest.")
		return errMsg
	}
	return fmt.Sprintf("未知的发信渠道校验: %s", sw.Type)
}
