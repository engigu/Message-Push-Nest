package message

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type feishuResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Feishu struct {
	AccessToken string
	Secret      string
}

// genSign 生成飞书签名
func (f *Feishu) genSign(timestamp int64) string {
	if f.Secret == "" {
		return ""
	}
	stringToSign := fmt.Sprintf("%v\n%s", timestamp, f.Secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature
}

// SendMessageText 发送文本消息
func (f *Feishu) SendMessageText(content string, atMobiles ...string) ([]byte, error) {
	timestamp := time.Now().Unix()
	sign := f.genSign(timestamp)

	msg := map[string]interface{}{
		"timestamp": strconv.FormatInt(timestamp, 10),
		"sign":      sign,
		"msg_type":  "text",
		"content": map[string]interface{}{
			"text": content,
		},
	}

	return f.send(msg)
}

// SendMessageMarkdown 发送 Markdown 消息
func (f *Feishu) SendMessageMarkdown(title, content string, atMobiles ...string) ([]byte, error) {
	timestamp := time.Now().Unix()
	sign := f.genSign(timestamp)

	// 处理 @ 人员
	atContent := ""
	if len(atMobiles) > 0 {
		for _, mobile := range atMobiles {
			if mobile == "all" {
				atContent += "<at user_id=\"all\">所有人</at>"
			} else {
				atContent += fmt.Sprintf("<at user_id=\"%s\"></at>", mobile)
			}
		}
		content = atContent + "\n" + content
	}

	msg := map[string]interface{}{
		"timestamp": strconv.FormatInt(timestamp, 10),
		"sign":      sign,
		"msg_type":  "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": title,
				},
			},
			"elements": []map[string]interface{}{
				{
					"tag":     "markdown",
					"content": content,
				},
			},
		},
	}

	return f.send(msg)
}

// send 发送请求
func (f *Feishu) send(msg map[string]interface{}) ([]byte, error) {
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", f.AccessToken)

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result feishuResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return body, fmt.Errorf("解析响应失败: %v", err)
	}

	if result.Code != 0 {
		return body, fmt.Errorf("飞书返回错误: %s", result.Msg)
	}

	return body, nil
}
