package util

import (
	"strings"
	"testing"
)

func TestValidateSafeURL(t *testing.T) {
	blockedURLs := []string{
		"http://127.0.0.1:8080/test",
		"http://localhost:3000",
		"http://10.0.1.5/admin",
		"http://192.168.1.100/status",
		"http://172.16.1.1:9090",
		"http://169.254.169.254/latest/meta-data",
		"file:///etc/passwd",
		"ftp://evil.com/leak",
		"gopher://127.0.0.1:6379/_flushall",
	}

	for _, u := range blockedURLs {
		err := ValidateSafeURL(u)
		if err == nil {
			t.Fatalf("expected SSRF error for %s, but got nil", u)
		}
	}

	// 合法公网域名格式校验
	validURLs := []string{
		"https://example.com/webhook",
		"http://example.com:8080/api",
	}

	for _, u := range validURLs {
		err := ValidateSafeURL(u)
		if err != nil && strings.Contains(err.Error(), "SSRF Protection") {
			t.Fatalf("unexpected SSRF false positive for %s: %v", u, err)
		}
	}
}

