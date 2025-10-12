package util

import (
	"fmt"
	"message-nest/pkg/setting"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	Username string `json:"username"`
	Password string `json:"password"`

	jwt.RegisteredClaims
}

func GenerateToken(username, password string, expDays int) (string, error) {
	// 如果传入的天数小于等于0，使用默认值1天
	if expDays <= 0 {
		expDays = 1
	}
	expHours := time.Duration(expDays) * 24 * time.Hour
	SetClaims := UserClaims{
		Username: username,
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
