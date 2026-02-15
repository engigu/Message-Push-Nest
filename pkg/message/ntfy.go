package message

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
)

type Ntfy struct {
	Url      string
	Topic    string
	Priority string
	Icon     string
	Token    string
	Username string
	Password string
	Actions  string
}

func encodeRFC2047(text string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("=?utf-8?B?%s?=", encoded)
}

func (n *Ntfy) Request(title, content string) ([]byte, error) {
	if n.Url == "" {
		n.Url = "https://ntfy.sh"
	}
	url := fmt.Sprintf("%s/%s", n.Url, n.Topic)

	req, err := http.NewRequest("POST", url, bytes.NewBufferString(content))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Title", encodeRFC2047(title))
	priority := n.Priority
	if priority == "" {
		priority = "3"
	}
	req.Header.Set("Priority", priority)
	if n.Icon != "" {
		req.Header.Set("Icon", n.Icon)
	}
	if n.Actions != "" {
		req.Header.Set("Actions", encodeRFC2047(n.Actions))
	}

	if n.Token != "" {
		req.Header.Set("Authorization", "Bearer "+n.Token)
	} else if n.Username != "" && n.Password != "" {
		authStr := n.Username + ":" + n.Password
		encodedAuth := base64.StdEncoding.EncodeToString([]byte(authStr))
		req.Header.Set("Authorization", "Basic "+encodedAuth)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("ntfy response error: %s", string(body))
	}

	return body, nil
}
