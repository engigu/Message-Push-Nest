package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test-sec", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test-sec", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("expected X-Content-Type-Options: nosniff, got %q", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Errorf("expected X-Frame-Options: SAMEORIGIN, got %q", got)
	}
	if got := w.Header().Get("X-XSS-Protection"); got != "1; mode=block" {
		t.Errorf("expected X-XSS-Protection: 1; mode=block, got %q", got)
	}
	if got := w.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Errorf("expected Referrer-Policy: strict-origin-when-cross-origin, got %q", got)
	}
}

func TestRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 限制 3 次 / 2 秒
	r.GET("/test-limit", IPRateLimiter(3, 2*time.Second), func(c *gin.Context) {
		c.String(http.StatusOK, "allowed")
	})

	// 前 3 次应该正常通过
	for i := 1; i <= 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test-limit", nil)
		req.RemoteAddr = "192.168.1.10:54321"
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("request #%d expected 200, got %d", i, w.Code)
		}
	}

	// 第 4 次应该被拦截 429 Too Many Requests
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test-limit", nil)
	req.RemoteAddr = "192.168.1.10:54321"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("request #4 expected 429, got %d", w.Code)
	}
	if retryAfter := w.Header().Get("Retry-After"); retryAfter == "" {
		t.Errorf("expected Retry-After header on 429 response")
	}
}
