package message

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type PushMe struct {
	PushKey string
	URL     string
	Date    string
	Type    string
}

func (p *PushMe) Request(title, content string) (string, error) {
	apiURL := p.URL
	if apiURL == "" {
		apiURL = "https://push.i-i.me/"
	}

	data := url.Values{}
	data.Set("push_key", p.PushKey)
	data.Set("title", title)
	data.Set("content", content)
	if p.Date != "" {
		data.Set("date", p.Date)
	}
	if p.Type != "" {
		data.Set("type", p.Type)
	}

	resp, err := http.Post(apiURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == 200 && string(body) == "success" {
		return string(body), nil
	}

	return string(body), fmt.Errorf("PushMe response error: %s", string(body))
}
