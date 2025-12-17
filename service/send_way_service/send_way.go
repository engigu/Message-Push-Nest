package send_way_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/constant"
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

// WayValidator 渠道验证接口
type WayValidator interface {
	Validate(authJson string) (string, interface{})
}

// WayTester 渠道测试接口
type WayTester interface {
	Test() (string, string)
}

// 渠道注册表
var (
	validatorRegistry = map[string]func() WayValidator{
		constant.MessageTypeEmail:           func() WayValidator { return &WayDetailEmail{} },
		constant.MessageTypeDtalk:           func() WayValidator { return &WayDetailDTalk{} },
		constant.MessageTypeQyWeiXin:        func() WayValidator { return &WayDetailQyWeiXin{} },
		constant.MessageTypeFeishu:          func() WayValidator { return &WayDetailFeishu{} },
		constant.MessageTypeCustom:          func() WayValidator { return &WayDetailCustom{} },
		constant.MessageTypeWeChatOFAccount: func() WayValidator { return &WeChatOFAccount{} },
		constant.MessageTypeMessageNest:     func() WayValidator { return &MessageNest{} },
		constant.MessageTypeAliyunSMS:       func() WayValidator { return &WayDetailAliyunSMS{} },
	}
	testerRegistry = map[string]func(interface{}) WayTester{
		constant.MessageTypeEmail:           func(m interface{}) WayTester { return m.(*WayDetailEmail) },
		constant.MessageTypeDtalk:           func(m interface{}) WayTester { return m.(*WayDetailDTalk) },
		constant.MessageTypeQyWeiXin:        func(m interface{}) WayTester { return m.(*WayDetailQyWeiXin) },
		constant.MessageTypeFeishu:          func(m interface{}) WayTester { return m.(*WayDetailFeishu) },
		constant.MessageTypeCustom:          func(m interface{}) WayTester { return m.(*WayDetailCustom) },
		constant.MessageTypeWeChatOFAccount: func(m interface{}) WayTester { return m.(*WeChatOFAccount) },
		constant.MessageTypeMessageNest:     func(m interface{}) WayTester { return m.(*MessageNest) },
		constant.MessageTypeAliyunSMS:       func(m interface{}) WayTester { return m.(*WayDetailAliyunSMS) },
	}
)

// WayDetailEmail 邮箱渠道明细字段
type WayDetailEmail struct {
	Server  string `validate:"required,max=50" label:"SMTP服务地址"`
	Port    int    `validate:"required,max=65535" label:"SMTP服务端口"`
	Account string `validate:"required,email" label:"邮箱账号"`
	Passwd  string `validate:"required,max=50" label:"邮箱密码"`
}

func (w *WayDetailEmail) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "邮箱auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, *w
}

func (w *WayDetailEmail) Test() (string, string) {
	testMsg := "This is a test message from message-nest."
	var emailer message.EmailMessage
	emailer.Init(w.Server, 465, w.Account, w.Passwd)
	errMsg := emailer.SendTextMessage(w.Account, "test email", testMsg)
	return errMsg, ""
}

// WayDetailDTalk 钉钉渠道明细字段
type WayDetailDTalk struct {
	AccessToken string `json:"access_token" validate:"required,max=100" label:"钉钉access_token"`
	Keys        string `json:"keys" validate:"max=200" label:"钉钉关键字"`
	Secret      string `json:"secret" validate:"max=100" label:"钉钉加签秘钥"`
}

func (w *WayDetailDTalk) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "钉钉auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, *w
}

func (w *WayDetailDTalk) Test() (string, string) {
	testMsg := "This is a test message from message-nest."
	var cli = message.Dtalk{
		AccessToken: w.AccessToken,
		Secret:      w.Secret,
	}
	res, err := cli.SendMessageText(testMsg + w.Keys)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailQyWeiXin 企业微信渠道明细字段
type WayDetailQyWeiXin struct {
	AccessToken string `json:"access_token" validate:"required,max=100" label:"企业微信access_token"`
}

func (w *WayDetailQyWeiXin) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "企业微信auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, *w
}

func (w *WayDetailQyWeiXin) Test() (string, string) {
	testMsg := "This is a test message from message-nest."
	var cli = message.QyWeiXin{
		AccessToken: w.AccessToken,
	}
	res, err := cli.SendMessageText(testMsg)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailFeishu 飞书渠道明细字段
type WayDetailFeishu struct {
	AccessToken string `json:"access_token" validate:"required,max=100" label:"飞书access_token"`
	Keys        string `json:"keys" validate:"max=200" label:"飞书关键字"`
	Secret      string `json:"secret" validate:"max=100" label:"飞书加签秘钥"`
}

func (w *WayDetailFeishu) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "飞书auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, *w
}

func (w *WayDetailFeishu) Test() (string, string) {
	testMsg := "This is a test message from message-nest."
	var cli = message.Feishu{
		AccessToken: w.AccessToken,
		Secret:      w.Secret,
	}
	res, err := cli.SendMessageText(testMsg + w.Keys)
	if err != nil {
		return fmt.Sprintf("发送失败：%s", err), string(res)
	}
	return "", string(res)
}

// WayDetailCustom 自定义渠道
type WayDetailCustom struct {
	Webhook string `json:"webhook" validate:"required,max=200" label:"自定义的webhook地址"`
	Body    string `json:"body" validate:"max=2000" label:"自定义的请求体"`
}

func (w *WayDetailCustom) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "自定义参数反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, *w
}

func (w *WayDetailCustom) Test() (string, string) {
	return "自定义webhook不用测试运行，请直接添加", ""
}

// WeChatOFAccount 微信公众号
type WeChatOFAccount struct {
	AppID     string `json:"appID" validate:"required,max=200" label:"微信公众号id"`
	APPSecret string `json:"appsecret" validate:"max=2000" label:"微信公众号秘钥"`
	TempID    string `json:"tempid" validate:"max=2000" label:"模板消息id"`
}

func (w *WeChatOFAccount) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "微信公众号反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, *w
}

func (w *WeChatOFAccount) Test() (string, string) {
	return "微信公众号模板消息不用测试运行，请直接添加", ""
}

// WayDetailAliyunSMS 阿里云短信渠道明细字段
type WayDetailAliyunSMS struct {
	AccessKeyId     string `json:"access_key_id" validate:"required,max=100" label:"AccessKeyId"`
	AccessKeySecret string `json:"access_key_secret" validate:"required,max=100" label:"AccessKeySecret"`
	SignName        string `json:"sign_name" validate:"required,max=50" label:"短信签名"`
}

func (w *WayDetailAliyunSMS) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "阿里云短信auth反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, *w
}

func (w *WayDetailAliyunSMS) Test() (string, string) {
	return "阿里云短信不用测试运行，请直接添加", ""
}

// MessageNest 自托管消息
type MessageNest struct {
}

func (w *MessageNest) Validate(authJson string) (string, interface{}) {
	var empty interface{}
	err := json.Unmarshal([]byte(authJson), w)
	if err != nil {
		return "自托管消息反序列化失败！", empty
	}
	_, msg := app.CommonPlaygroundValid(*w)
	return msg, *w
}

func (w *MessageNest) Test() (string, string) {
	return "自托管消息不用测试运行，请直接添加", ""
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

// getValidator 根据渠道类型获取对应的验证器
func (sw *SendWay) getValidator() WayValidator {
	factory, exists := validatorRegistry[sw.Type]
	if !exists {
		return nil
	}
	return factory()
}

// ValidateDiffWay 各种发信渠道具体字段校验
func (sw *SendWay) ValidateDiffWay() (string, interface{}) {
	var empty interface{}
	validator := sw.getValidator()
	if validator == nil {
		return fmt.Sprintf("未知的发信渠道校验: %s", sw.Type), empty
	}
	return validator.Validate(sw.Auth)
}

// TestSendWay 尝试带发信测试连通性
func (sw *SendWay) TestSendWay(msgObj interface{}) (string, string) {
	factory, exists := testerRegistry[sw.Type]
	if !exists {
		return fmt.Sprintf("未知的发信渠道测试: %s", sw.Type), ""
	}
	tester := factory(msgObj)
	return tester.Test()
}
