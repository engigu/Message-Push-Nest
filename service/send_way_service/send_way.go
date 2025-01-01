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

// WayDetailQyWeiXin 企业微信渠道明细字段
type WayDetailQyWeiXin struct {
	AccessToken string `json:"access_token" validate:"required,max=100" label:"企业微信access_token"`
}

// WayDetailCustom 自定义渠道
type WayDetailCustom struct {
	Webhook string `json:"webhook" validate:"required,max=200" label:"自定义的webhook地址"`
	Body    string `json:"body" validate:"max=2000" label:"自定义的请求体"`
}

// WeChatOFAccount 微信公众号
type WeChatOFAccount struct {
	AppID     string `json:"appID" validate:"required,max=200" label:"微信公众号id"`
	APPSecret string `json:"appsecret" validate:"max=2000" label:"微信公众号秘钥"`
	TempID    string `json:"tempid" validate:"max=2000" label:"模板消息id"`
}

// MessageNest 自托管消息
type MessageNest struct {
}

func (sw *SendWay) GetByID() (interface{}, error) {
	return models.GetWayByID(sw.ID)
}

func (sw *SendWay) NameIsExist(method string) error {
	way, err := models.GetWayByName(sw.Name)
	if err != nil {
		return err
	}
	if method == "add" {
		if len(way.ID) > 0 {
			return errors.New(fmt.Sprintf("已经存在同名的渠道名：%s", way.Name))
		}
	} else {
		// 只允许当前的重名
		if len(way.ID) > 0 && way.ID != sw.ID {
			return errors.New(fmt.Sprintf("已经存在同名的渠道名：%s", way.Name))
		}
	}
	return nil
}

func (sw *SendWay) Add() error {
	err := sw.NameIsExist("add")
	if err != nil {
		return err
	}
	return models.AddSendWay(sw.Name, sw.Auth, sw.Type, sw.CreatedBy, sw.ModifiedBy)
}

func (sw *SendWay) Edit() error {
	err := sw.NameIsExist("edit")
	if err != nil {
		return err
	}
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

func (sw *SendWay) Count() (int64, error) {
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
	} else if sw.Type == "QyWeiXin" {
		var config WayDetailQyWeiXin
		err := json.Unmarshal([]byte(sw.Auth), &config)
		if err != nil {
			return "企业微信auth反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(config)
		return Msg, config
	} else if sw.Type == "Custom" {
		var custom WayDetailCustom
		err := json.Unmarshal([]byte(sw.Auth), &custom)
		if err != nil {
			return "自定义参数反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(custom)
		return Msg, custom
	} else if sw.Type == "WeChatOFAccount" {
		var wca WeChatOFAccount
		err := json.Unmarshal([]byte(sw.Auth), &wca)
		if err != nil {
			return "微信公众号反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(wca)
		return Msg, wca
	} else if sw.Type == "MessageNest" {
		var wca MessageNest
		err := json.Unmarshal([]byte(sw.Auth), &wca)
		if err != nil {
			return "自托管消息反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(wca)
		return Msg, wca
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
	qywxAuth, ok := msgObj.(WayDetailQyWeiXin)
	if ok {
		var cli = message.QyWeiXin{
			AccessToken: qywxAuth.AccessToken,
		}
		res, err := cli.SendMessageText(testMsg)
		if err != nil {
			return fmt.Sprintf("发送失败：%s", err), string(res)
		}
		return "", string(res)
	}
	_, ok = msgObj.(WeChatOFAccount)
	if ok {
		return "微信公众号模板消息不用测试运行，请直接添加", ""
	}
	_, ok = msgObj.(MessageNest)
	if ok {
		return "自托管消息不用测试运行，请直接添加", ""
	}
	return fmt.Sprintf("未知的发信渠道校验: %s", sw.Type), ""
}
