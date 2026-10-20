package util

import (
	"crypto/rand"
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
