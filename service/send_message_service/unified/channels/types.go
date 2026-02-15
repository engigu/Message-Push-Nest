package channels

import "message-nest/pkg/constant"

// 重导出常量，保持向后兼容
const (
	FormatTypeText     = constant.FormatTypeText
	FormatTypeHTML     = constant.FormatTypeHTML
	FormatTypeMarkdown = constant.FormatTypeMarkdown
)

const (
	MessageTypeEmail           = constant.MessageTypeEmail
	MessageTypeDtalk           = constant.MessageTypeDtalk
	MessageTypeQyWeiXin        = constant.MessageTypeQyWeiXin
	MessageTypeFeishu          = constant.MessageTypeFeishu
	MessageTypeCustom          = constant.MessageTypeCustom
	MessageTypeWeChatOFAccount = constant.MessageTypeWeChatOFAccount
	MessageTypeMessageNest     = constant.MessageTypeMessageNest
	MessageTypeAliyunSMS       = constant.MessageTypeAliyunSMS
	MessageTypeTelegram        = constant.MessageTypeTelegram
	MessageTypeBark            = constant.MessageTypeBark
	MessageTypePushMe          = constant.MessageTypePushMe
	MessageTypeNtfy            = constant.MessageTypeNtfy
)

// UnifiedMessageContent 统一的消息内容结构
type UnifiedMessageContent struct {
	Title     string
	URL       string
	Text      string
	HTML      string
	Markdown  string
	AtMobiles []string
	AtUserIds []string
	AtAll     bool
	Summary   string
	ImageURL  string
	Extra     map[string]interface{}
}

func (m *UnifiedMessageContent) HasText() bool     { return m.Text != "" }
func (m *UnifiedMessageContent) HasHTML() bool     { return m.HTML != "" }
func (m *UnifiedMessageContent) HasMarkdown() bool { return m.Markdown != "" }

func (m *UnifiedMessageContent) GetAtMobiles() []string {
	if m.AtMobiles == nil {
		return []string{}
	}
	return m.AtMobiles
}

func (m *UnifiedMessageContent) GetAtUserIds() []string {
	if m.AtUserIds == nil {
		return []string{}
	}
	return m.AtUserIds
}

func (m *UnifiedMessageContent) IsAtAll() bool { return m.AtAll }
