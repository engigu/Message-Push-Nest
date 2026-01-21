package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type barkResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Bark struct {
	PushKey string
	Archive string
	Group   string
	Sound   string
	Icon    string
	Level   string
	URL     string
}

func (b *Bark) Request(title, content string) ([]byte, error) {
	url := b.getURL()
	data := map[string]interface{}{
		"title": title,
		"body":  content,
	}
	if b.Archive != "" {
		data["isArchive"] = b.Archive
	}
	if b.Group != "" {
		data["group"] = b.Group
	}
	if b.Sound != "" {
		data["sound"] = b.Sound
	}
	if b.Icon != "" {
		data["icon"] = b.Icon
	}
	if b.Level != "" {
		data["level"] = b.Level
	}
	if b.URL != "" {
		data["url"] = b.URL
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json;charset=utf-8", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r barkResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return body, err
	}

	if r.Code != 200 {
		return body, fmt.Errorf("bark response error: %s", string(body))
	}
	return body, nil
}

func (b *Bark) getURL() string {
	pushKey := b.PushKey
	if strings.HasPrefix(pushKey, "http") {
		return pushKey
	}
	return fmt.Sprintf("https://api.day.app/%s", pushKey)
}
