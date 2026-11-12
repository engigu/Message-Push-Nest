package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"message-nest/pkg/setting"
)

func getSecretKey() []byte {
	secret := setting.AppSetting.JwtSecret
	if secret == "" {
		secret = "DefaultMessageNestSecretKey2026"
	}
	h := sha256.Sum256([]byte(secret))
	return h[:]
}

// EncryptTokenSecure 使用 AES-256-GCM 生成高强度防伪造发信 Token（带消息认证防篡改）
func EncryptTokenSecure(text string) (string, error) {
	key := getSecretKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return "v2_" + hex.EncodeToString(ciphertext), nil
}

// DecryptToken 解密发信 Token（优先按安全 AES-GCM 解密，若非安全格式则平滑兼容旧版异或 Token）
func DecryptToken(enc string) (string, error) {
	if enc == "" {
		return "", errors.New("empty token")
	}

	if strings.HasPrefix(enc, "v2_") {
		rawHex := enc[3:]
		data, err := hex.DecodeString(rawHex)
		if err != nil {
			return "", err
		}
		key := getSecretKey()
		block, err := aes.NewCipher(key)
		if err != nil {
			return "", err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", err
		}
		nonceSize := gcm.NonceSize()
		if len(data) < nonceSize {
			return "", errors.New("token too short")
		}
		nonce, ciphertext := data[:nonceSize], data[nonceSize:]
		plain, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return "", fmt.Errorf("token 认证解密失败: %w", err)
		}
		return string(plain), nil
	}

	// 存量兼容：旧版 71 确定性 salt + 单字节异或
	return DecryptTokenHex(enc, 71)
}

// getDeterministicSalt 根据字符串内容生成确定性 salt（范围 0~255）
func getDeterministicSalt(text string) byte {
	var sum int
	for i := 0; i < len(text); i++ {
		sum = (sum + int(text[i])*(i+1)) & 0xFF
	}
	return byte(sum)
}

// EncryptTokenHex 按照确定性 salt + 异或规则生成十六进制 token（兼容保留）
func EncryptTokenHex(text string, key byte) string {
	salt := getDeterministicSalt(text)
	out := make([]byte, 0, 2+len(text)*2)
	out = append(out, []byte(hex.EncodeToString([]byte{salt}))...)
	for i := 0; i < len(text); i++ {
		b := text[i] ^ key ^ byte((int(salt)+i)&0xFF)
		enc := make([]byte, 2)
		hex.Encode(enc, []byte{b})
		out = append(out, enc...)
	}
	return string(out)
}

// DecryptTokenHex 解析十六进制 token 为原始字符串（兼容保留）
func DecryptTokenHex(enc string, key byte) (string, error) {
	if enc == "" {
		return "", errors.New("empty token")
	}
	bs, err := hex.DecodeString(enc)
	if err != nil {
		return "", err
	}
	if len(bs) < 1 {
		return "", errors.New("invalid token length")
	}
	salt := bs[0]
	out := make([]byte, len(bs)-1)
	for i := 1; i < len(bs); i++ {
		out[i-1] = bs[i] ^ key ^ byte((int(salt)+i-1)&0xFF)
	}
	return string(out), nil
}
