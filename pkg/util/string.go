package util

import (
	"math/rand"
	"time"
)

func GenerateRandomString(length int) string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result string
	var charSetLen = len(charSet)

	for i := 0; i < length; i++ {
		randomIndex := random.Intn(charSetLen)
		result += string(charSet[randomIndex])
	}

	return result
}

func GenerateUniqueID() string {
	randomString := GenerateRandomString(10)
	return randomString
}
