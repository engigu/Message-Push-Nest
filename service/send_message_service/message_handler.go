package send_message_service

import (
	"message-nest/models"
	"message-nest/service/send_way_service"
)

// 消息类型常量定义
const (
	MessageTypeEmail          = "Email"
	MessageTypeDtalk          = "Dtalk"
	MessageTypeQyWeiXin       = "QyWeiXin"
	MessageTypeCustom         = "Custom"
	MessageTypeWeChatOFAccount = "WeChatOFAccount"
	MessageTypeMessageNest    = "MessageNest"
)

// MessageHandler 消息处理器接口
type MessageHandler interface {
	// Send 发送消息
	// 返回：响应内容，错误信息
	Send(msgObj interface{}, ins models.SendTasksIns, typeC string, title string, content string, url string) (string, string)
	
	// GetType 返回该处理器支持的消息类型字符串
	GetType() string
}

// MessageHandlerRegistry 消息处理器注册表
type MessageHandlerRegistry struct {
	handlers map[string]MessageHandler
}

// NewMessageHandlerRegistry 创建新的消息处理器注册表
func NewMessageHandlerRegistry() *MessageHandlerRegistry {
	return &MessageHandlerRegistry{
		handlers: make(map[string]MessageHandler),
	}
}

// Register 注册消息处理器
func (r *MessageHandlerRegistry) Register(handler MessageHandler) {
	r.handlers[handler.GetType()] = handler
}

// GetHandler 根据类型字符串获取对应的处理器
func (r *MessageHandlerRegistry) GetHandler(wayType string) (MessageHandler, bool) {
	handler, ok := r.handlers[wayType]
	return handler, ok
}

// 全局消息处理器注册表
var globalRegistry = NewMessageHandlerRegistry()

// init 初始化时注册所有处理器
func init() {
	globalRegistry.Register(&EmailHandler{})
	globalRegistry.Register(&DtalkHandler{})
	globalRegistry.Register(&QyWeiXinHandler{})
	globalRegistry.Register(&CustomHandler{})
	globalRegistry.Register(&WeChatOfAccountHandler{})
	globalRegistry.Register(&HostMessageHandler{})
}

// GetGlobalRegistry 获取全局注册表
func GetGlobalRegistry() *MessageHandlerRegistry {
	return globalRegistry
}

// EmailHandler 邮箱消息处理器
type EmailHandler struct{}

func (h *EmailHandler) GetType() string {
	return MessageTypeEmail
}

func (h *EmailHandler) Send(msgObj interface{}, ins models.SendTasksIns, typeC string, title string, content string, url string) (string, string) {
	auth, ok := msgObj.(send_way_service.WayDetailEmail)
	if !ok {
		return "", "类型转换失败"
	}
	es := EmailService{}
	errMsg := es.SendTaskEmail(auth, ins, typeC, title, content)
	return "", errMsg
}

// DtalkHandler 钉钉消息处理器
type DtalkHandler struct{}

func (h *DtalkHandler) GetType() string {
	return MessageTypeDtalk
}

func (h *DtalkHandler) Send(msgObj interface{}, ins models.SendTasksIns, typeC string, title string, content string, url string) (string, string) {
	auth, ok := msgObj.(send_way_service.WayDetailDTalk)
	if !ok {
		return "", "类型转换失败"
	}
	es := DtalkService{}
	return es.SendDtalkMessage(auth, ins, typeC, title, content)
}

// QyWeiXinHandler 企业微信消息处理器
type QyWeiXinHandler struct{}

func (h *QyWeiXinHandler) GetType() string {
	return MessageTypeQyWeiXin
}

func (h *QyWeiXinHandler) Send(msgObj interface{}, ins models.SendTasksIns, typeC string, title string, content string, url string) (string, string) {
	auth, ok := msgObj.(send_way_service.WayDetailQyWeiXin)
	if !ok {
		return "", "类型转换失败"
	}
	es := QyWeiXinService{}
	return es.SendQyWeiXinMessage(auth, ins, typeC, title, content)
}

// CustomHandler 自定义webhook消息处理器
type CustomHandler struct{}

func (h *CustomHandler) GetType() string {
	return MessageTypeCustom
}

func (h *CustomHandler) Send(msgObj interface{}, ins models.SendTasksIns, typeC string, title string, content string, url string) (string, string) {
	auth, ok := msgObj.(send_way_service.WayDetailCustom)
	if !ok {
		return "", "类型转换失败"
	}
	cs := CustomService{}
	return cs.SendCustomMessage(auth, ins, typeC, title, content)
}

// WeChatOfAccountHandler 微信公众号消息处理器
type WeChatOfAccountHandler struct{}

func (h *WeChatOfAccountHandler) GetType() string {
	return MessageTypeWeChatOFAccount
}

func (h *WeChatOfAccountHandler) Send(msgObj interface{}, ins models.SendTasksIns, typeC string, title string, content string, url string) (string, string) {
	auth, ok := msgObj.(send_way_service.WeChatOFAccount)
	if !ok {
		return "", "类型转换失败"
	}
	cs := WeChatOfAccountService{}
	return cs.SendWeChatOfAccountMessage(auth, ins, typeC, title, content, url)
}

// HostMessageHandler 托管消息处理器
type HostMessageHandler struct{}

func (h *HostMessageHandler) GetType() string {
	return MessageTypeMessageNest
}

func (h *HostMessageHandler) Send(msgObj interface{}, ins models.SendTasksIns, typeC string, title string, content string, url string) (string, string) {
	auth, ok := msgObj.(send_way_service.MessageNest)
	if !ok {
		return "", "类型转换失败"
	}
	cs := HostMessageService{}
	return cs.SendHostMessage(auth, ins, typeC, title, content)
}
