package message

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type response struct {
	Code int    `json:"errcode"`
	Msg  string `json:"errmsg"`
}

type Dtalk struct {
	AccessToken string
	Secret      string
}

func (t *Dtalk) Request(msg interface{}) ([]byte, error) {
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
	var r response
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
func (t *Dtalk) SendMessageText(text string, at ...string) ([]byte, error) {
	msg := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": text,
		},
	}
	
	// 添加@功能
	if len(at) > 0 {
		atMobiles := []string{}
		isAtAll := false
		
		for _, mobile := range at {
			if mobile == "all" || mobile == "@all" {
				isAtAll = true
			} else {
				atMobiles = append(atMobiles, mobile)
			}
		}
		
		msg["at"] = map[string]interface{}{
			"atMobiles": atMobiles,
			"isAtAll":   isAtAll,
		}
	}
	
	resp, err := t.Request(msg)
	return resp, err
}

func (t *Dtalk) SendMessageMarkdown(title, text string, at ...string) ([]byte, error) {
	msg := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  text,
		},
	}
	
	// 添加@功能
	if len(at) > 0 {
		atMobiles := []string{}
		isAtAll := false
		
		for _, mobile := range at {
			if mobile == "all" || mobile == "@all" {
				isAtAll = true
			} else {
				atMobiles = append(atMobiles, mobile)
			}
		}
		
		msg["at"] = map[string]interface{}{
			"atMobiles": atMobiles,
			"isAtAll":   isAtAll,
		}
	}
	
	resp, err := t.Request(msg)
	return resp, err
}

func (t *Dtalk) hmacSha256(stringToSign string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func (t *Dtalk) getURL() string {
	wh := "https://oapi.dingtalk.com/robot/send?access_token=" + t.AccessToken
	timestamp := time.Now().UnixNano() / 1e6
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, t.Secret)
	sign := t.hmacSha256(stringToSign, t.Secret)
	url := fmt.Sprintf("%s&timestamp=%d&sign=%s", wh, timestamp, sign)
	return url
}
