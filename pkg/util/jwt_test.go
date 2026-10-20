package util

import (
	"message-nest/pkg/setting"
	"testing"
)

func TestJWTTokenGenerationAndParse(t *testing.T) {
	// 初始化 JWT 密钥配置
	setting.AppSetting.JwtSecret = "TestSecretKey123456"

	userID := "UR1234567890"
	username := "test_operator"
	password := "SecretPass123"
	role := "admin"

	// 1. 生成 Token
	token, err := GenerateToken(userID, username, password, role, 7)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatalf("expected non-empty token string")
	}

	// 2. 解析 Token
	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Fatalf("expected UserID %s, got %s", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Fatalf("expected Username %s, got %s", username, claims.Username)
	}
	if claims.Role != role {
		t.Fatalf("expected Role %s, got %s", role, claims.Role)
	}

	// 3. 验证伪造或错误密钥
	invalidToken := token + "tampered"
	_, err = ParseToken(invalidToken)
	if err == nil {
		t.Fatalf("expected error for tampered token, got nil")
	}
}

func TestSSOTicketGenerationAndParse(t *testing.T) {
	setting.AppSetting.JwtSecret = "TestSecretKey123456"

	provider := "oidc"
	subject := "sub_corp_999"
	email := "corp@example.com"
	preferredUsername := "corp_user"

	ticket, err := GenerateSSOTicket(provider, subject, email, preferredUsername)
	if err != nil {
		t.Fatalf("GenerateSSOTicket failed: %v", err)
	}
	if ticket == "" {
		t.Fatalf("expected non-empty ticket string")
	}

	claims, err := ParseSSOTicket(ticket)
	if err != nil {
		t.Fatalf("ParseSSOTicket failed: %v", err)
	}

	if claims.Provider != provider || claims.Subject != subject || claims.Email != email || claims.PreferredUsername != preferredUsername {
		t.Fatalf("parsed claims mismatch: %+v", claims)
	}

	// 验证防伪造
	tamperedTicket := ticket + "abc"
	_, err = ParseSSOTicket(tamperedTicket)
	if err == nil {
		t.Fatalf("expected error for tampered ticket, got nil")
	}
}

