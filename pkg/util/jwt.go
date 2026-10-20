package util

import (
	"fmt"
	"message-nest/pkg/setting"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Password string `json:"password"`

	jwt.RegisteredClaims
}

func GenerateToken(userID, username, password, role string, expDays int) (string, error) {
	// 如果传入的天数小于等于0，使用默认值1天
	if expDays <= 0 {
		expDays = 1
	}
	if role == "" {
		role = "user"
	}
	expHours := time.Duration(expDays) * 24 * time.Hour
	SetClaims := UserClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Password: EncodeMD5(password),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expHours)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "message-nest",
		},
	}

	tokenStruct := jwt.NewWithClaims(jwt.SigningMethodHS256, SetClaims)
	token, err := tokenStruct.SignedString([]byte(setting.AppSetting.JwtSecret))
	return token, err
}

func ParseToken(tokenString string) (*UserClaims, error) {
	tokenObj, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(setting.AppSetting.JwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tokenObj.Claims.(*UserClaims); ok && tokenObj.Valid {
		return claims, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

// SSOTicketClaims 用于第三方单点登录授权后未绑定内部账号时的临时安全凭据
type SSOTicketClaims struct {
	Provider          string `json:"provider"`
	Subject           string `json:"subject"`
	Email             string `json:"email"`
	PreferredUsername string `json:"preferred_username"`

	jwt.RegisteredClaims
}

// GenerateSSOTicket 生成单点登录绑定临时票据（15分钟有效）
func GenerateSSOTicket(provider, subject, email, preferredUsername string) (string, error) {
	claims := SSOTicketClaims{
		Provider:          provider,
		Subject:           subject,
		Email:             email,
		PreferredUsername: preferredUsername,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "message-nest-sso-ticket",
		},
	}
	tokenStruct := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tokenStruct.SignedString([]byte(setting.AppSetting.JwtSecret))
}

// ParseSSOTicket 解析并验证单点登录临时凭据
func ParseSSOTicket(ticketString string) (*SSOTicketClaims, error) {
	tokenObj, err := jwt.ParseWithClaims(ticketString, &SSOTicketClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(setting.AppSetting.JwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tokenObj.Claims.(*SSOTicketClaims); ok && tokenObj.Valid {
		if claims.Issuer != "message-nest-sso-ticket" {
			return nil, fmt.Errorf("非法凭证类型")
		}
		return claims, nil
	}
	return nil, fmt.Errorf("无效或已过期的凭证")
}

