package send_way_service

import (
	"encoding/json"
	"strings"
)

const MaskedSecretPlaceholder = "******"

var sensitiveKeys = []string{
	"passwd",
	"password",
	"secret",
	"access_token",
	"access_key_secret",
	"bot_token",
	"push_key",
	"appsecret",
	"token",
	"key",
	"iv",
}

func isSensitiveKey(key string) bool {
	lower := strings.ToLower(key)
	for _, k := range sensitiveKeys {
		if strings.Contains(lower, k) {
			return true
		}
	}
	return false
}

// MaskWayAuth 对 auth JSON 中的密码、密钥等敏感信息进行脱敏（替换为 ******）
func MaskWayAuth(authJson string) string {
	if strings.TrimSpace(authJson) == "" {
		return authJson
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(authJson), &data); err != nil {
		return authJson
	}

	for k, v := range data {
		if isSensitiveKey(k) {
			if strVal, ok := v.(string); ok && strVal != "" {
				data[k] = MaskedSecretPlaceholder
			}
		}
	}

	masked, err := json.Marshal(data)
	if err != nil {
		return authJson
	}
	return string(masked)
}

// MergeWayAuth 合并新提交的 auth 与原有 auth。若新提交中的敏感字段为掩码占位符，则还原保留原有未脱敏值
func MergeWayAuth(newAuthJson, oldAuthJson string) string {
	if strings.TrimSpace(oldAuthJson) == "" {
		return newAuthJson
	}
	var oldData map[string]interface{}
	if err := json.Unmarshal([]byte(oldAuthJson), &oldData); err != nil {
		return newAuthJson
	}

	var newData map[string]interface{}
	if err := json.Unmarshal([]byte(newAuthJson), &newData); err != nil {
		return newAuthJson
	}

	for k, v := range newData {
		if strVal, ok := v.(string); ok {
			if strVal == MaskedSecretPlaceholder {
				if origVal, exists := oldData[k]; exists {
					newData[k] = origVal
				}
			}
		}
	}

	merged, err := json.Marshal(newData)
	if err != nil {
		return newAuthJson
	}
	return string(merged)
}

