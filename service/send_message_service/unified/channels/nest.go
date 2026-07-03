package channels

import (
	"encoding/json"
	"message-nest/models"
	"message-nest/service/hosted_message_service"
	"message-nest/service/send_way_service"
	"regexp"
)

type MessageNestChannel struct{ *BaseChannel }

func NewMessageNestChannel() *MessageNestChannel {
	return &MessageNestChannel{BaseChannel: NewBaseChannel(MessageTypeMessageNest, []string{FormatTypeMarkdown, FormatTypeHTML, FormatTypeText})}
}

func (c *MessageNestChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	_, ok := msgObj.(*send_way_service.MessageNest)
	if !ok {
		return "", "类型转换失败"
	}
	contentType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}

	// 解析配置
	var nestConfig models.InsMessageNestConfig
	if ins.Config != "" {
		if err := json.Unmarshal([]byte(ins.Config), &nestConfig); err != nil {
			return "", "解析自托管消息配置失败: " + err.Error()
		}
	}

	// 匹配黑名单正则，满足则不记录
	if nestConfig.BlacklistRegex != "" {
		matched, err := regexp.MatchString(nestConfig.BlacklistRegex, formattedContent)
		if err == nil && matched {
			return "消息匹配黑名单正则[" + nestConfig.BlacklistRegex + "]，未记录", ""
		}
	}

	messageService := hosted_message_service.HostMessageService{
		Title:   content.Title,
		Content: formattedContent,
		Type:    contentType,
	}
	
	if content.Extra != nil {
		if key, ok := content.Extra["HostedMsgKey"].(string); ok {
			messageService.UniqueKey = key
		}
	}
	err = messageService.Add()
	var res, errMsg string
	if err != nil {
		errMsg = err.Error()
		res = "托管消息创建失败！"
	} else {
		res = "托管消息创建成功！"
	}
	return res, errMsg
}
