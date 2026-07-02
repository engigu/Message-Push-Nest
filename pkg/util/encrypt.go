package util

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"encoding/base64"

	"github.com/klauspost/compress/zstd"
)

// TripleDesEncrypt 3DES 加密，使用 PKCS5Padding
func TripleDesEncrypt(origData, key []byte) (string, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return "", err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:8]) // 使用 Key 的前 8 字节作为 IV
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return base64.StdEncoding.EncodeToString(crypted), nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// ZstdCompress 使用 zstd 压缩字节
func ZstdCompress(src []byte) []byte {
	var encoder, _ = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedDefault))
	return encoder.EncodeAll(src, make([]byte, 0, len(src)))
}
