package channels

// 消息格式类型常量
const (
	FormatTypeText     = "text"
	FormatTypeHTML     = "html"
	FormatTypeMarkdown = "markdown"
)

// 消息类型常量
const (
	MessageTypeEmail           = "Email"
	MessageTypeDtalk           = "Dtalk"
	MessageTypeQyWeiXin        = "QyWeiXin"
	MessageTypeCustom          = "Custom"
	MessageTypeWeChatOFAccount = "WeChatOFAccount"
	MessageTypeMessageNest     = "MessageNest"
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
