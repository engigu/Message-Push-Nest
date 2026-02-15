package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type gotifyResponse struct {
	Id        int    `json:"id"`
	Message   string `json:"message"`
	ErrorCode int    `json:"errorCode"`
}

type Gotify struct {
	Url      string
	Token    string
	Priority int
}

func (g *Gotify) Request(title, content string) ([]byte, error) {
	// Construct the URL with token
	u, err := url.Parse(fmt.Sprintf("%s/message", g.Url))
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("token", g.Token)
	u.RawQuery = q.Encode()

	data := map[string]interface{}{
		"title":    title,
		"message":  content,
		"priority": g.Priority,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r gotifyResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return body, err
	}

	if r.Id == 0 {
		return body, fmt.Errorf("gotify response error: %s", string(body))
	}

	return body, nil
}
