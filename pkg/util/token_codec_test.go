package util

import (
	"message-nest/pkg/setting"
	"strings"
	"testing"
)

func TestTokenCodecSecureAndLegacyCompatibility(t *testing.T) {
	setting.AppSetting.JwtSecret = "TestSecretForTokenCodec98765"
	originalTaskID := "TK_SECURE_999"

	// 1. 测试安全加密与认证解密
	secToken, err := EncryptTokenSecure(originalTaskID)
	if err != nil {
		t.Fatalf("EncryptTokenSecure failed: %v", err)
	}
	if !strings.HasPrefix(secToken, "v2_") {
		t.Fatalf("expected v2_ prefix for secure token, got: %s", secToken)
	}

	decrypted, err := DecryptToken(secToken)
	if err != nil {
		t.Fatalf("DecryptToken failed for secure token: %v", err)
	}
	if decrypted != originalTaskID {
		t.Fatalf("expected decrypted %s, got %s", originalTaskID, decrypted)
	}

	// 2. 防篡改测试：修改密文中任意字符应导致验签解密失败
	tampered := secToken[:len(secToken)-2] + "ff"
	_, err = DecryptToken(tampered)
	if err == nil {
		t.Fatalf("expected tamper detection error, but DecryptToken succeeded!")
	}

	// 3. 存量兼容测试：旧版 71 异或 Token 平滑解析
	legacyToken := EncryptTokenHex(originalTaskID, 71)
	legacyDecrypted, err := DecryptToken(legacyToken)
	if err != nil {
		t.Fatalf("DecryptToken failed for legacy token: %v", err)
	}
	if legacyDecrypted != originalTaskID {
		t.Fatalf("expected legacy decrypted %s, got %s", originalTaskID, legacyDecrypted)
	}
}

