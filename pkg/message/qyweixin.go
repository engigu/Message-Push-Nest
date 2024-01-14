package message

import (
	"bytes"
	"encoding/json"
	"errors"
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
		return body, errors.New(fmt.Sprintf("response error: %s", string(body)))
	}
	return body, err
}

// SendMessageText Function to send message
func (t *QyWeiXin) SendMessageText(text string, at ...string) ([]byte, error) {
	msg := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": text,
		},
	}
	resp, err := t.Request(msg)
	return resp, err
}

func (t *QyWeiXin) SendMessageMarkdown(title, text string, at ...string) ([]byte, error) {
	msg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title":   title,
			"content": text,
		},
	}
	resp, err := t.Request(msg)
	return resp, err
}

func (t *QyWeiXin) getURL() string {
	url := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=" + t.AccessToken
	return url
}
