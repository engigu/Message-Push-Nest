package message

import (
	"github.com/silenceper/wechat/v2"
	"github.com/silenceper/wechat/v2/cache"
	offConfig "github.com/silenceper/wechat/v2/officialaccount/config"
	"github.com/silenceper/wechat/v2/officialaccount/message"
	"github.com/sirupsen/logrus"
)

type WeChatOFAccount struct {
	AppID      string
	AppSecret  string
	ToUser     string
	TemplateID string
	URL        string
}

// 使用内存缓存进行token的存储
var memory = cache.NewMemory()

func (cw *WeChatOFAccount) Send(title string, content string) (string, error) {
	wc := wechat.NewWechat()
	cfg := &offConfig.Config{
		AppID:     cw.AppID,
		AppSecret: cw.AppSecret,
		Cache:     memory,
	}
	officialAccount := wc.GetOfficialAccount(cfg)

	// 获取 Access Token
	_, err := officialAccount.GetAccessToken()
	if err != nil {
		logrus.Errorf("获取access token失败:%s", err)
		return "", err
	}

	msgData := make(map[string]*message.TemplateDataItem)
	msgData["content"] = &message.TemplateDataItem{
		Value: content,
	}
	msgData["title"] = &message.TemplateDataItem{
		Value: title,
		//Color: "#173177",
	}

	// 创建模板消息
	templateMessage := &message.TemplateMessage{
		ToUser:     cw.ToUser,
		TemplateID: cw.TemplateID,
		URL:        cw.URL,
		Data:       msgData,
	}

	// 发送模板消息
	_, err = officialAccount.GetTemplate().Send(templateMessage)
	if err != nil {
		logrus.Errorf("发送模板消息失败: %s", err)
		return "", err
	}
	//logrus.Infof("模板消息发送成功。 消息ID: %d", msgID)
	return "", nil
}
