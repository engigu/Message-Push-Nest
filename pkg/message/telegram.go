package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type telegramResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

type Telegram struct {
	BotToken string
	ChatID   string
	ApiHost  string // 可选的自定义API地址（优先级最高）
	ProxyURL string // 可选的代理地址，支持 http://、https://、socks5:// 格式
}

func (t *Telegram) Request(params map[string]interface{}) ([]byte, error) {
	apiURL := t.getAPIURL()

	// 构建请求体
	data := url.Values{}
	for key, value := range params {
		data.Set(key, fmt.Sprintf("%v", value))
	}

	// 创建 HTTP 客户端
	client := t.getHTTPClient()

	resp, err := client.Post(apiURL, "application/x-www-form-urlencoded", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// 忽略关闭错误
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r telegramResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return body, err
	}

	if !r.Ok {
		return body, fmt.Errorf("telegram api error: %s", r.Description)
	}

	return body, nil
}

// SendMessageText 发送文本消息
func (t *Telegram) SendMessageText(text string) ([]byte, error) {
	params := map[string]interface{}{
		"chat_id":                  t.ChatID,
		"text":                     text,
		"disable_web_page_preview": "true",
	}

	return t.Request(params)
}

// SendMessageMarkdown 发送Markdown格式消息
func (t *Telegram) SendMessageMarkdown(text string) ([]byte, error) {
	params := map[string]interface{}{
		"chat_id":                  t.ChatID,
		"text":                     text,
		"parse_mode":               "Markdown",
		"disable_web_page_preview": "true",
	}

	return t.Request(params)
}

// SendMessageHTML 发送HTML格式消息
func (t *Telegram) SendMessageHTML(text string) ([]byte, error) {
	params := map[string]interface{}{
		"chat_id":                  t.ChatID,
		"text":                     text,
		"parse_mode":               "HTML",
		"disable_web_page_preview": "true",
	}

	return t.Request(params)
}

func (t *Telegram) getAPIURL() string {
	// 自定义 API 地址优先级最高
	if t.ApiHost != "" {
		return fmt.Sprintf("%s/bot%s/sendMessage", t.ApiHost, t.BotToken)
	}
	return fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.BotToken)
}

// getHTTPClient 获取配置了代理的 HTTP 客户端
func (t *Telegram) getHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 如果配置了代理且没有自定义 API 地址，则使用代理
	// 自定义 API 地址优先级更高，通常用于自建代理服务器
	if t.ProxyURL != "" && t.ApiHost == "" {
		proxyURL, err := url.Parse(t.ProxyURL)
		if err == nil {
			// 判断是否为 SOCKS5 代理
			if strings.HasPrefix(strings.ToLower(t.ProxyURL), "socks5://") {
				// 使用 SOCKS5 代理
				dialer, err := t.createSOCKS5Dialer(proxyURL)
				if err == nil {
					client.Transport = &http.Transport{
						DialContext: dialer.DialContext,
					}
				}
			} else {
				// 使用 HTTP/HTTPS 代理
				client.Transport = &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
			}
		}
	}

	return client
}

// createSOCKS5Dialer 创建 SOCKS5 代理拨号器
func (t *Telegram) createSOCKS5Dialer(proxyURL *url.URL) (proxy.ContextDialer, error) {
	// 解析代理地址
	host := proxyURL.Host

	// 检查是否有认证信息
	var auth *proxy.Auth
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{
			User:     proxyURL.User.Username(),
			Password: password,
		}
	}

	// 创建基础拨号器
	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// 创建 SOCKS5 拨号器
	dialer, err := proxy.SOCKS5("tcp", host, auth, baseDialer)
	if err != nil {
		return nil, err
	}

	// 转换为 ContextDialer
	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, errors.New("failed to convert to ContextDialer")
	}

	return contextDialer, nil
}
