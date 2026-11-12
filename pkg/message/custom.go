package message

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
	// "message-nest/pkg/util"
)

type CustomWebhook struct {
	Webhook string
	Header  map[string]string
	Body    string
}

func (cw *CustomWebhook) ParseHeaders(headerStr string) map[string]string {
	headers := make(map[string]string)
	if headerStr == "" {
		return headers
	}
	lines := strings.Split(headerStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}

// 暂时注释 SSRF 限制以支持内网/局域网 Webhook
// var Client = util.NewSafeHTTPClient(5 * time.Second)
var Client = &http.Client{
	Timeout: 5 * time.Second,
}

func (cw *CustomWebhook) Request(url string, msg string, headers map[string]string) ([]byte, error) {
	// 暂时注释 SSRF URL 校验
	// if err := util.ValidateSafeURL(url); err != nil {
	// 	return nil, err
	// }

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(msg)))
	if err != nil {
		return nil, err
	}

	// 设置默认 Content-Type 为 application/json，如果用户没有提供的话
	req.Header.Set("Content-Type", "application/json")

	// 设置自定义 Header
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := Client.Do(req)
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
	return body, err
}
