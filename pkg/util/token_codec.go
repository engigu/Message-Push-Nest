package util

import (
	"encoding/hex"
	"errors"
)

//// getDeterministicSalt 根据字符串内容生成确定性 salt（范围 0~255）
//func getDeterministicSalt(text string) byte {
//	var sum int
//	for i := 0; i < len(text); i++ {
//		sum = (sum + int(text[i])*(i+1)) & 0xFF
//	}
//	return byte(sum)
//}

//// EncryptTokenHex 按照确定性 salt + 异或规则生成十六进制 token
//// 规则：首字节为 salt；其余字节为 char ^ key ^ ((salt + idx) & 0xFF)
//func EncryptTokenHex(text string, key byte) string {
//    salt := getDeterministicSalt(text)
//    // 预分配：2个hex字符的salt + 每个字符2个hex
//    out := make([]byte, 0, 2+len(text)*2)
//    // 写入salt
//    out = append(out, []byte(hex.EncodeToString([]byte{salt}))...)
//    // 写入加密数据
//    for i := 0; i < len(text); i++ {
//        b := text[i] ^ key ^ byte((int(salt)+i)&0xFF)
//        enc := make([]byte, 2)
//        hex.Encode(enc, []byte{b})
//        out = append(out, enc...)
//    }
//    return string(out)
//}

// DecryptTokenHex 解析十六进制 token 为原始字符串
// 规则：首字节为salt；后续字节与 key ^ ((salt + idx) & 0xFF) 异或
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
