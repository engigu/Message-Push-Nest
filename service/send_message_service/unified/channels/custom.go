package channels

import (
	"encoding/json"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_way_service"
	"regexp"
	"strings"
)

type CustomChannel struct{ *BaseChannel }

func NewCustomChannel() *CustomChannel {
	return &CustomChannel{BaseChannel: NewBaseChannel(MessageTypeCustom, []string{FormatTypeText})}
}

func (c *CustomChannel) FormatContent(content *UnifiedMessageContent) (string, string, error) {
	if content.HasText() {
		return FormatTypeText, content.Text, nil
	}
	if content.HasMarkdown() {
		return FormatTypeText, markdownToText(content.Markdown), nil
	}
	if content.HasHTML() {
		return FormatTypeText, htmlToText(content.HTML), nil
	}
	return FormatTypeText, "", nil
}

func (c *CustomChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(send_way_service.WayDetailCustom)
	if !ok {
		return "", "类型转换失败"
	}
	_, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}
	cli := message.CustomWebhook{}
	data, _ := json.Marshal(formattedContent)
	dataStr := strings.Trim(string(data), "\"")
	bodyStr := strings.Replace(auth.Body, "TEXT", dataStr, -1)
	res, err := cli.Request(auth.Webhook, bodyStr)
	var errMsg string
	if err != nil {
		errMsg = fmt.Sprintf("发送失败：%s", err.Error())
	}
	return string(res), errMsg
}

func markdownToText(md string) string {
	t := md
	t = regexp.MustCompile(`#{1,6}\s+`).ReplaceAllString(t, "")
	t = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(t, "$1")
	t = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(t, "$1")
	t = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(t, "$1")
	t = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(t, "$1")
	t = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(t, "$1")
	t = regexp.MustCompile("```[^`]*```").ReplaceAllString(t, "")
	t = regexp.MustCompile("`([^`]+)`").ReplaceAllString(t, "$1")
	t = regexp.MustCompile(`(?m)^>\s+`).ReplaceAllString(t, "")
	t = regexp.MustCompile(`(?m)^[\*\-\+]\s+`).ReplaceAllString(t, "")
	t = regexp.MustCompile(`(?m)^\d+\.\s+`).ReplaceAllString(t, "")
	return strings.TrimSpace(t)
}

func htmlToText(html string) string {
	t := html
	t = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(t, "")
	t = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`).ReplaceAllString(t, "")
	t = regexp.MustCompile(`(?i)<br\s*/?>`).ReplaceAllString(t, "\n")
	t = regexp.MustCompile(`(?i)</p>`).ReplaceAllString(t, "\n")
	t = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(t, "")
	t = strings.ReplaceAll(t, "&nbsp;", " ")
	t = strings.ReplaceAll(t, "&lt;", "<")
	t = strings.ReplaceAll(t, "&gt;", ">")
	t = strings.ReplaceAll(t, "&amp;", "&")
	t = strings.ReplaceAll(t, "&quot;", "\"")
	t = regexp.MustCompile(`\n{3,}`).ReplaceAllString(t, "\n\n")
	return strings.TrimSpace(t)
}
