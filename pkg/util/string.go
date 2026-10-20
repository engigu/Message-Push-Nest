package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

func GenerateRandomString(length int) string {
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charSetLen := big.NewInt(int64(len(charSet)))
	result := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charSetLen)
		if err != nil {
			result[i] = charSet[0]
		} else {
			result[i] = charSet[num.Int64()]
		}
	}

	return string(result)
}

func GenerateUniqueID() string {
	return GenerateRandomString(10)
}

// GetDefaultCronForUser 为指定用户生成错开的默认定时清理 Cron 表达式
// 基于用户 ID 哈希打散在凌晨 1:00 ~ 4:59 的低峰期（共 240 个不同分钟槽位），避免多用户并发执行造成数据库洪峰
func GetDefaultCronForUser(userID string, seedOffset int) string {
	if userID == "" {
		if seedOffset == 0 {
			return "1 0 * * *"
		}
		return "30 0 * * *"
	}
	var hash uint32 = 2166136261
	for _, c := range userID {
		hash ^= uint32(c)
		hash *= 16777619
	}
	hash += uint32(seedOffset)
	minute := hash % 60
	hour := 1 + (hash/60)%4 // 凌晨 1 点 ~ 4 点
	return fmt.Sprintf("%d %d * * *", minute, hour)
}
