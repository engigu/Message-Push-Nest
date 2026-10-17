package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"message-nest/models"
	"message-nest/pkg/constant"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/auth/sso/status", GetSSOStatus)
		apiV1.GET("/auth/oidc/login", OidcLogin)
		apiV1.GET("/auth/oidc/callback", OidcCallback)
		apiV1.GET("/auth/oauth2/login", OAuth2Login)
		apiV1.GET("/auth/oauth2/callback", OAuth2Callback)
		apiV1.POST("/auth/oidc/test", OidcTest)
		apiV1.POST("/auth/oauth2/test", OAuth2Test)
	}
	return r
}

func TestGetSSOStatusDisabled(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/sso/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Enabled   bool  `json:"enabled"`
			Providers []any `json:"providers"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
}

func TestOidcLoginWithoutConfig(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/oidc/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 when OIDC not configured, got %d", w.Code)
	}
}

func TestOidcCallbackStateMismatch(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/oidc/callback?state=wrong_state&code=test_code", nil)
	req.AddCookie(&http.Cookie{
		Name:  oidcStateCookie,
		Value: "correct_state.oidc./",
	})
	req.AddCookie(&http.Cookie{
		Name:  oidcVerifierCookie,
		Value: "test_verifier",
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected 302 redirect, got %d", w.Code)
	}
	location := w.Header().Get("Location")
	if !strings.Contains(location, "error=") {
		t.Fatalf("expected error in redirect Location, got %s", location)
	}
}

func TestOidcTestEmptyIssuer(t *testing.T) {
	r := setupTestRouter()

	payload := `{"type": "oidc", "issuer": ""}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/oidc/test", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected code 400 for empty issuer, got %d", resp.Code)
	}
}

func TestOAuth2TestEmptyEndpoints(t *testing.T) {
	r := setupTestRouter()

	payload := `{"type": "oauth2", "oauth2_auth_url": ""}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/oauth2/test", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected code 400 for empty endpoints, got %d", resp.Code)
	}
}

var _ = models.Settings{}
var _ = constant.SecuritySectionName
