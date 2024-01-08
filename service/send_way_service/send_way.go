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
	AccessToken string `json:"access_token" validate:"required,max=100" label:"钉钉access_token"`
	Keys        string `json:"keys" validate:"max=200" label:"钉钉关键字"`
	Secret      string `json:"secret" validate:"max=100" label:"钉钉加签秘钥"`
}

// WayDetailCustom 自定义渠道
type WayDetailCustom struct {
	Webhook string `json:"webhook" validate:"required,max=200" label:"自定义的webhook地址"`
	Body    string `json:"body" validate:"max=2000" label:"自定义的请求体"`
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
	} else if sw.Type == "Custom" {
		var custom WayDetailCustom
		err := json.Unmarshal([]byte(sw.Auth), &custom)
		if err != nil {
			return "自定义参数反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(custom)
		return Msg, custom
	}
	return fmt.Sprintf("未知的发信渠道校验: %s", sw.Type), empty
}

// TestSendWay 尝试带发信测试连通性
func (sw *SendWay) TestSendWay(msgObj interface{}) (string, string) {
	testMsg := "This is a test message from message-nest."
	emailAuth, ok := msgObj.(WayDetailEmail)
	if ok {
		var emailer message.EmailMessage
		emailer.Init(emailAuth.Server, 465, emailAuth.Account, emailAuth.Passwd)
		errMsg := emailer.SendTextMessage(emailAuth.Account, "test email", testMsg)
		return errMsg, ""
	}
	dtalkAuth, ok := msgObj.(WayDetailDTalk)
	if ok {
		var cli = message.Dtalk{
			AccessToken: dtalkAuth.AccessToken,
			Secret:      dtalkAuth.Secret,
		}
		res, err := cli.SendMessageText(testMsg + dtalkAuth.Keys)
		if err != nil {
			return fmt.Sprintf("发送失败：%s", err), string(res)
		}
		return "", string(res)
	}
	customAuth, ok := msgObj.(WayDetailCustom)
	if ok {
		var cli = message.CustomWebhook{}
		data, _ := json.Marshal(testMsg)
		dataStr := string(data)
		dataStr = strings.Trim(dataStr, "\"")
		bodyStr := strings.Replace(customAuth.Body, "TEXT", dataStr, -1)
		res, err := cli.Request(customAuth.Webhook, bodyStr)
		if err != nil {
			return fmt.Sprintf("发送失败：%s", err), string(res)
		}
		return "", string(res)
	}
	return fmt.Sprintf("未知的发信渠道校验: %s", sw.Type), ""
}
