package message

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"message-nest/models"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

type qywxAppTokenResponse struct {
	Code        int    `json:"errcode"`
	Msg         string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type qywxAppSendResponse struct {
	Code int    `json:"errcode"`
	Msg  string `json:"errmsg"`
}

type QyWeiXinApp struct {
	WayID    string
	CorpID   string
	AgentID  string
	Secret   string
	ApiHost  string
	ProxyURL string
}

type tokenCache struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetToken 获取 access_token (优先从数据库 settings 表缓存获取)
func (q *QyWeiXinApp) GetToken() (string, error) {
	// 如果有缓存且未过期，直接返回
	if token, ok := q.getCachedToken(); ok {
		return token, nil
	}

	// 否则，请求微信 API 获取
	token, expiresIn, err := q.requestNewToken()
	if err != nil {
		return "", err
	}

	// 保存到数据库缓存
	q.saveCachedToken(token, expiresIn)
	return token, nil
}

// getCachedToken 从数据库中获取缓存的 token
func (q *QyWeiXinApp) getCachedToken() (string, bool) {
	if q.WayID == "" {
		return "", false
	}
	s, err := models.GetSettingByKey("qyweixin_app_token", q.WayID)
	if err != nil || s.Value == "" {
		return "", false
	}
	var cache tokenCache
	if err := json.Unmarshal([]byte(s.Value), &cache); err != nil {
		return "", false
	}
	// 提前 5 分钟失效
	if time.Now().Add(5 * time.Minute).After(cache.ExpiresAt) {
		return "", false
	}
	return cache.Token, true
}

// saveCachedToken 将 token 保存到数据库缓存
func (q *QyWeiXinApp) saveCachedToken(token string, expiresIn int) {
	if q.WayID == "" {
		return
	}
	cache := tokenCache{
		Token:     token,
		ExpiresAt: time.Now().Add(time.Duration(expiresIn) * time.Second),
	}
	bytes, err := json.Marshal(cache)
	if err != nil {
		return
	}

	models.DeleteSettingByKey("qyweixin_app_token", q.WayID)
	models.AddOneSetting(models.Settings{
		Section: "qyweixin_app_token",
		Key:     q.WayID,
		Value:   string(bytes),
	})
}

// requestNewToken 请求微信接口获取新 Token
func (q *QyWeiXinApp) requestNewToken() (string, int, error) {
	apiHost := "https://qyapi.weixin.qq.com"
	if q.ApiHost != "" {
		apiHost = q.ApiHost
	}
	apiURL := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", apiHost, q.CorpID, q.Secret)

	client := q.getHTTPClient()
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}

	var r qywxAppTokenResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return "", 0, err
	}

	if r.Code != 0 {
		return "", 0, fmt.Errorf("qyweixin app gettoken err: %d - %s", r.Code, r.Msg)
	}

	if r.AccessToken == "" {
		return "", 0, errors.New("qyweixin app gettoken returned empty token")
	}

	return r.AccessToken, r.ExpiresIn, nil
}

// SendMessage 发送统一包装消息
func (q *QyWeiXinApp) SendMessage(toUser, toParty, toTag, msgType string, contentMap map[string]interface{}) ([]byte, error) {
	token, err := q.GetToken()
	if err != nil {
		return nil, err
	}

	apiHost := "https://qyapi.weixin.qq.com"
	if q.ApiHost != "" {
		apiHost = q.ApiHost
	}
	apiURL := fmt.Sprintf("%s/cgi-bin/message/send?access_token=%s", apiHost, token)

	agentID, _ := strconv.Atoi(q.AgentID)
	msg := map[string]interface{}{
		"touser":   toUser,
		"toparty":  toParty,
		"totag":    toTag,
		"msgtype":  msgType,
		"agentid":  agentID,
		"safe":     0,
		msgType:    contentMap,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	client := q.getHTTPClient()
	resp, err := client.Post(apiURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r qywxAppSendResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return body, err
	}

	if r.Code != 0 {
		return body, fmt.Errorf("qyweixin app send err: %d - %s", r.Code, r.Msg)
	}

	return body, nil
}

// SendTextMessage 发送文本消息
func (q *QyWeiXinApp) SendTextMessage(toUser, toParty, toTag, text string) ([]byte, error) {
	content := map[string]interface{}{
		"content": text,
	}
	return q.SendMessage(toUser, toParty, toTag, "text", content)
}

// SendMarkdownMessage 发送 Markdown 消息
func (q *QyWeiXinApp) SendMarkdownMessage(toUser, toParty, toTag, text string) ([]byte, error) {
	content := map[string]interface{}{
		"content": text,
	}
	return q.SendMessage(toUser, toParty, toTag, "markdown", content)
}

// getHTTPClient 获取配置了代理的 HTTP 客户端
func (q *QyWeiXinApp) getHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 自定义 API 地址优先级更高，通常用于自建代理服务器，此时不用本地代理
	if q.ProxyURL != "" && q.ApiHost == "" {
		proxyURL, err := url.Parse(q.ProxyURL)
		if err == nil {
			if strings.HasPrefix(strings.ToLower(q.ProxyURL), "socks5://") {
				dialer, err := q.createSOCKS5Dialer(proxyURL)
				if err == nil {
					client.Transport = &http.Transport{
						DialContext: dialer.DialContext,
					}
				}
			} else {
				client.Transport = &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				}
			}
		}
	}

	return client
}

// createSOCKS5Dialer 创建 SOCKS5 代理拨号器
func (q *QyWeiXinApp) createSOCKS5Dialer(proxyURL *url.URL) (proxy.ContextDialer, error) {
	host := proxyURL.Host
	var auth *proxy.Auth
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{
			User:     proxyURL.User.Username(),
			Password: password,
		}
	}

	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	dialer, err := proxy.SOCKS5("tcp", host, auth, baseDialer)
	if err != nil {
		return nil, err
	}

	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, errors.New("failed to convert to ContextDialer")
	}

	return contextDialer, nil
}
