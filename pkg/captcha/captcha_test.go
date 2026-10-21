package captcha

import (
	"testing"
	"time"
)

func TestGenerateAndVerifyCaptcha(t *testing.T) {
	// 1. 测试数字验证码生成
	digitRes, err := GenerateDigitCaptcha()
	if err != nil {
		t.Fatalf("GenerateDigitCaptcha failed: %v", err)
	}
	if digitRes.CaptchaID == "" || digitRes.ImageBase64 == "" {
		t.Fatalf("GenerateDigitCaptcha returned empty id or image")
	}

	// 2. 测试算术验证码生成
	mathRes, err := GenerateMathCaptcha()
	if err != nil {
		t.Fatalf("GenerateMathCaptcha failed: %v", err)
	}
	if mathRes.CaptchaID == "" || mathRes.ImageBase64 == "" {
		t.Fatalf("GenerateMathCaptcha returned empty id or image")
	}

	// 3. 错误验证码测试
	if VerifyDirect(digitRes.CaptchaID, "wrong_code") {
		t.Fatalf("VerifyDirect should return false for incorrect answer")
	}
	// 4. 重放测试 (VerifyDirect 内部是一次性清除)
	if VerifyDirect(digitRes.CaptchaID, "wrong_code") {
		t.Fatalf("VerifyDirect should return false for already deleted captcha")
	}
}

func TestVerificationToken(t *testing.T) {
	ip := "192.168.1.100"
	token := GenerateVerificationToken(ip)
	if token == "" {
		t.Fatalf("GenerateVerificationToken returned empty token")
	}

	// 1. 正常校验通过
	ok, reason := VerifyToken(token, ip)
	if !ok {
		t.Fatalf("VerifyToken failed on valid token: %s", reason)
	}

	// 2. 异 IP 盗用测试
	ok, reason = VerifyToken(token, "10.0.0.1")
	if ok {
		t.Fatalf("VerifyToken should fail when IP does not match")
	}

	// 3. 篡改测试
	tamperedToken := token + "tampered"
	ok, _ = VerifyToken(tamperedToken, ip)
	if ok {
		t.Fatalf("VerifyToken should fail on tampered token")
	}

	// 4. 空 Token 测试
	ok, _ = VerifyToken("", ip)
	if ok {
		t.Fatalf("VerifyToken should fail on empty token")
	}
}

func TestCleanExpiredCaptchas(t *testing.T) {
	testStore := NewMemoryStoreWithCron(10*time.Millisecond, 100)
	_ = testStore.Set("expired_1", "1234")
	_ = testStore.Set("expired_2", "5678")

	time.Sleep(20 * time.Millisecond)

	cleaned := testStore.CleanExpired()
	if cleaned != 2 {
		t.Fatalf("expected 2 expired captchas cleaned, got %d", cleaned)
	}

	if testStore.Get("expired_1", false) != "" || testStore.Get("expired_2", false) != "" {
		t.Fatalf("expired captchas should be purged from store")
	}
}

