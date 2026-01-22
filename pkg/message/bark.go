package message

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
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
	Key     string
	IV      string
}

func (b *Bark) Request(title, content string) ([]byte, error) {
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

	var postData interface{}
	url := b.getURL()

	if b.Key != "" && b.IV != "" {
		// Use encryption
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		ciphertext, err := b.encryptPayload(string(jsonData))
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %v", err)
		}
		postData = map[string]interface{}{
			"ciphertext": ciphertext,
			"device_key": b.PushKey,
			"sound":      b.Sound,
		}
		// When using encryption, use the push endpoint if PushKey is just a key
		if !strings.HasPrefix(b.PushKey, "http") {
			url = "https://api.day.app/push"
		}
	} else {
		// Normal request
		postData = data
	}

	jsonData, err := json.Marshal(postData)
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

func (b *Bark) encryptPayload(payload string) (string, error) {
	key := []byte(b.Key)
	iv := []byte(b.IV)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	paddedPayload := b.pkcs7Pad([]byte(payload), aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(paddedPayload))
	mode.CryptBlocks(ciphertext, paddedPayload)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (b *Bark) pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}
