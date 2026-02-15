package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type qywxResponse struct {
	Code int    `json:"errcode"`
	Msg  string `json:"errmsg"`
}

type QyWeiXin struct {
	AccessToken string
}

func (t *QyWeiXin) Request(msg interface{}) ([]byte, error) {
	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(t.getURL(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var r qywxResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return body, err
	}
	if r.Code != 0 {
		return body, fmt.Errorf("response error: %s", string(body))
	}
	return body, err
}

// SendMessageText Function to send message
func (t *QyWeiXin) SendMessageText(text string, at ...string) ([]byte, error) {
	msg := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": text,
		},
	}

	// 添加@功能
	// 企业微信支持两种@方式：
	// 1. mentioned_list: userid列表或"@all"
	// 2. mentioned_mobile_list: 手机号列表
	if len(at) > 0 {
		mentionedList := []string{}
		mentionedMobileList := []string{}

		for _, item := range at {
			if item == "@all" || item == "all" {
				mentionedList = append(mentionedList, "@all")
			} else if len(item) == 11 && item[0] == '1' {
				// 判断是否为手机号（简单判断：11位且以1开头）
				mentionedMobileList = append(mentionedMobileList, item)
			} else {
				// 否则当作userid处理
				mentionedList = append(mentionedList, item)
			}
		}

		textContent := msg["text"].(map[string]interface{})
		if len(mentionedList) > 0 {
			textContent["mentioned_list"] = mentionedList
		}
		if len(mentionedMobileList) > 0 {
			textContent["mentioned_mobile_list"] = mentionedMobileList
		}
	}

	resp, err := t.Request(msg)
	return resp, err
}

func (t *QyWeiXin) SendMessageMarkdown(title, text string, at ...string) ([]byte, error) {
	msg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": text,
		},
	}

	// 企业微信Markdown消息不支持@功能，但可以在内容中手动添加
	// 如果需要@功能，建议使用text类型

	resp, err := t.Request(msg)
	return resp, err
}

func (t *QyWeiXin) getURL() string {
	url := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + t.AccessToken
	return url
}
