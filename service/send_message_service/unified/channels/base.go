package channels

import "message-nest/models"

type Channel interface {
	GetType() string
	GetSupportedFormats() []string
	FormatContent(content *UnifiedMessageContent) (formatType string, formattedContent string, err error)
	SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string)
}

type BaseChannel struct {
	channelType      string
	supportedFormats []string
}

func NewBaseChannel(channelType string, supportedFormats []string) *BaseChannel {
	return &BaseChannel{channelType: channelType, supportedFormats: supportedFormats}
}

func (c *BaseChannel) GetType() string                { return c.channelType }
func (c *BaseChannel) GetSupportedFormats() []string { return c.supportedFormats }

func (c *BaseChannel) FormatContent(content *UnifiedMessageContent) (string, string, error) {
	for _, formatType := range c.supportedFormats {
		switch formatType {
		case FormatTypeMarkdown:
			if content.HasMarkdown() {
				return FormatTypeMarkdown, content.Markdown, nil
			}
		case FormatTypeHTML:
			if content.HasHTML() {
				return FormatTypeHTML, content.HTML, nil
			}
		case FormatTypeText:
			if content.HasText() {
				return FormatTypeText, content.Text, nil
			}
		}
	}
	if content.HasText() {
		return FormatTypeText, content.Text, nil
	}
	return FormatTypeText, "", nil
}
