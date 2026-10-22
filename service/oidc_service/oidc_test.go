package oidc_service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNormalizeIssuer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://id.example.com", "https://id.example.com"},
		{"https://id.example.com/", "https://id.example.com"},
		{"https://id.example.com///", "https://id.example.com"},
		{"https://id.example.com/.well-known/openid-configuration", "https://id.example.com"},
		{"https://id.example.com/.well-known/openid-configuration/", "https://id.example.com"},
		{" https://auth.example.com/app/ ", "https://auth.example.com/app"},
	}

	for _, tc := range tests {
		got := NormalizeIssuer(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeIssuer(%q) = %q, expected %q", tc.input, got, tc.expected)
		}
	}
}

func TestSecurityReaderScopes(t *testing.T) {
	r := NewSecurityReader()

	scopes := r.Scopes("email, profile   groups", "openid")
	expectedMap := map[string]bool{
		"openid":  true,
		"email":   true,
		"profile": true,
		"groups":  true,
	}

	if len(scopes) != 4 {
		t.Fatalf("expected 4 scopes, got %d: %v", len(scopes), scopes)
	}

	for _, s := range scopes {
		if !expectedMap[s] {
			t.Errorf("unexpected scope: %s", s)
		}
	}
}

func TestParseOAuth2Userinfo(t *testing.T) {
	// 1. 标准 GitHub 风格
	ghJSON := []byte(`{"id": 123456, "login": "octocat", "email": "octocat@github.com", "name": "The Octocat"}`)
	claims, err := ParseOAuth2Userinfo(ghJSON, "")
	if err != nil {
		t.Fatalf("ParseOAuth2Userinfo failed: %v", err)
	}
	if claims.Sub != "123456" {
		t.Errorf("expected Sub '123456', got %q", claims.Sub)
	}
	if claims.PreferredUsername != "octocat" {
		t.Errorf("expected PreferredUsername 'octocat', got %q", claims.PreferredUsername)
	}
	if claims.Email != "octocat@github.com" {
		t.Errorf("expected Email 'octocat@github.com', got %q", claims.Email)
	}

	// 2. 自定义 user_key
	customJSON := []byte(`{"custom_uid": "user_9988", "username": "alice", "mail": "alice@corp.internal"}`)
	claims2, err := ParseOAuth2Userinfo(customJSON, "custom_uid")
	if err != nil {
		t.Fatalf("ParseOAuth2Userinfo with custom key failed: %v", err)
	}
	if claims2.Sub != "user_9988" {
		t.Errorf("expected Sub 'user_9988', got %q", claims2.Sub)
	}
	if claims2.PreferredUsername != "alice" {
		t.Errorf("expected PreferredUsername 'alice', got %q", claims2.PreferredUsername)
	}
	if claims2.Email != "alice@corp.internal" {
		t.Errorf("expected Email 'alice@corp.internal', got %q", claims2.Email)
	}
}

func TestResolveRedirectURL(t *testing.T) {
	// 自定义覆盖
	custom := "https://my-domain.com/callback"
	got := ResolveRedirectURL(custom, "/api/v1/auth/oidc/callback", nil)
	if got != custom {
		t.Errorf("expected %q, got %q", custom, got)
	}

	// 默认路径（无 context）
	gotDefault := ResolveRedirectURL("", "/api/v1/auth/oidc/callback", nil)
	if gotDefault != "/api/v1/auth/oidc/callback" {
		t.Errorf("expected default /api/v1/auth/oidc/callback, got %q", gotDefault)
	}

	// 模拟带反向代理头及请求头的请求
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/sso/oidc/callback", nil)
	req.Host = "nest.corp.internal:8000"
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Host", "push.example.com")
	// 关键场景：第三方 IdP 登录完成跳回时的 Referer 头
	req.Header.Set("Referer", "https://sso.idp-provider.com/auth/realms/master/login-actions/authenticate")
	c.Request = req

	gotProxied := ResolveRedirectURL("", "/api/v1/auth/sso/oidc/callback", c)
	expectedProxied := "https://push.example.com/api/v1/auth/sso/oidc/callback"
	if gotProxied != expectedProxied {
		t.Errorf("expected %q, got %q", expectedProxied, gotProxied)
	}

	// 仅有 Host 没有 X-Forwarded-Host，且 Referer 来自外部 IdP 时，必须取本机 Host，严禁被 Referer 污染
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/sso/oidc/callback", nil)
	req2.Host = "my-nest.local:8080"
	req2.Header.Set("Referer", "https://keycloak.company.com/auth/login")
	c.Request = req2

	gotDirect := ResolveRedirectURL("", "/api/v1/auth/sso/oidc/callback", c)
	expectedDirect := "http://my-nest.local:8080/api/v1/auth/sso/oidc/callback"
	if gotDirect != expectedDirect {
		t.Errorf("expected %q, got %q", expectedDirect, gotDirect)
	}
}
