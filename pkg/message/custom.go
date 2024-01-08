package message

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

type CustomWebhook struct {
	Webhook string
	Body    string
}

var Client = &http.Client{
	Timeout: 8 * time.Second,
}

func (cw *CustomWebhook) Request(url string, msg string) ([]byte, error) {
	resp, err := Client.Post(url, "application/json", bytes.NewBuffer([]byte(msg)))
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
