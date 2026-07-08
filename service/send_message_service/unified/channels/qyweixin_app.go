package channels

import (
	"fmt"
	"message-nest/models"
	"message-nest/pkg/message"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"
)

type QyWeiXinAppChannel struct{ *BaseChannel }

func NewQyWeiXinAppChannel() *QyWeiXinAppChannel {
	return &QyWeiXinAppChannel{BaseChannel: NewBaseChannel(MessageTypeQyWeiXinApp, []string{FormatTypeMarkdown, FormatTypeText})}
}

func (c *QyWeiXinAppChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailQyWeiXinApp)
	if !ok {
		return "", "类型转换失败"
	}
	
	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr, ""
	}
	
	config, ok := configInterface.(models.InsQyWeiXinAppConfig)
	if !ok {
		return "企业微信自建应用 config 校验失败", ""
	}
	
	contentType, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}

	cli := message.QyWeiXinApp{
		WayID:    ins.WayID,
		CorpID:   auth.CorpID,
		AgentID:  auth.AgentID,
		Secret:   auth.Secret,
		ApiHost:  auth.ApiHost,
		ProxyURL: auth.ProxyURL,
	}

	var res []byte
	var errMsg string
	if contentType == FormatTypeText {
		res, err = cli.SendTextMessage(config.ToUser, config.ToParty, config.ToTag, formattedContent)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else if contentType == FormatTypeMarkdown {
		res, err = cli.SendMarkdownMessage(config.ToUser, config.ToParty, config.ToTag, formattedContent)
		if err != nil {
			errMsg = fmt.Sprintf("发送失败：%s", err.Error())
		}
	} else {
		errMsg = fmt.Sprintf("未知的企业微信自建应用发送内容类型：%s", contentType)
	}

	return string(res), errMsg
}
