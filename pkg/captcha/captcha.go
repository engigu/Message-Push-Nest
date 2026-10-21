package captcha

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mojocn/base64Captcha"
	"message-nest/pkg/setting"
)

type captchaItem struct {
	value     string
	createdAt time.Time
}

// MemoryStoreWithCron 实现 base64Captcha.Store 接口，
// 统一交由系统全局 Cron 周期性调度清理过期键，杜绝后台不可控孤儿协程。
type MemoryStoreWithCron struct {
	sync.RWMutex
	data       map[string]captchaItem
	expiration time.Duration
	maxEntries int
}

// NewMemoryStoreWithCron 创建接入 Cron 调度的验证码存储器
func NewMemoryStoreWithCron(expiration time.Duration, maxEntries int) *MemoryStoreWithCron {
	return &MemoryStoreWithCron{
		data:       make(map[string]captchaItem),
		expiration: expiration,
		maxEntries: maxEntries,
	}
}

func (s *MemoryStoreWithCron) Set(id string, value string) error {
	s.Lock()
	defer s.Unlock()

	// 达到最大容量时进行保护性淘汰
	if len(s.data) >= s.maxEntries {
		now := time.Now()
		for k, v := range s.data {
			if now.Sub(v.createdAt) > s.expiration {
				delete(s.data, k)
			}
		}
	}

	s.data[id] = captchaItem{
		value:     value,
		createdAt: time.Now(),
	}
	return nil
}

func (s *MemoryStoreWithCron) Get(id string, clear bool) string {
	s.Lock()
	defer s.Unlock()

	item, ok := s.data[id]
	if !ok {
		return ""
	}
	if time.Since(item.createdAt) > s.expiration {
		delete(s.data, id)
		return ""
	}
	if clear {
		delete(s.data, id)
	}
	return item.value
}

func (s *MemoryStoreWithCron) Verify(id, answer string, clear bool) bool {
	v := s.Get(id, clear)
	return v != "" && v == strings.TrimSpace(answer)
}

// CleanExpired 由系统统一定时任务 (cron) 调度执行，集中释放超时未消费的验证码记录
func (s *MemoryStoreWithCron) CleanExpired() int {
	s.Lock()
	defer s.Unlock()

	now := time.Now()
	cleaned := 0
	for id, item := range s.data {
		if now.Sub(item.createdAt) > s.expiration {
			delete(s.data, id)
			cleaned++
		}
	}
	return cleaned
}

// CaptchaStore 采用接入 cron 调度的内存存储，默认最多保留 10240 个，过期时间 5 分钟
var CaptchaStore = NewMemoryStoreWithCron(5*time.Minute, 10240)

// CleanExpiredCaptchas 由系统统一 cron 调度周期性调用，清理所有超时未核验的验证码
func CleanExpiredCaptchas() int {
	return CaptchaStore.CleanExpired()
}

// CaptchaResult 生成验证码的结果
type CaptchaResult struct {
	CaptchaID   string `json:"captcha_id"`
	ImageBase64 string `json:"captcha_image"`
}

// GenerateDigitCaptcha 生成纯数字验证码 (高 40，宽 120，4 位数字，抗机器识别干扰点)
func GenerateDigitCaptcha() (*CaptchaResult, error) {
	driver := base64Captcha.NewDriverDigit(40, 120, 4, 0.7, 80)
	cp := base64Captcha.NewCaptcha(driver, CaptchaStore)
	id, b64s, _, err := cp.Generate()
	if err != nil {
		return nil, err
	}
	return &CaptchaResult{
		CaptchaID:   id,
		ImageBase64: b64s,
	}, nil
}

// GenerateMathCaptcha 生成算术验证码 (如 3 + 5 = ?)
func GenerateMathCaptcha() (*CaptchaResult, error) {
	// 背景使用浅灰白色 #F5F7FA
	bgColor := &color.RGBA{R: 245, G: 247, B: 250, A: 255}
	driver := base64Captcha.NewDriverMath(40, 120, 0, 0, bgColor, nil, nil)
	cp := base64Captcha.NewCaptcha(driver, CaptchaStore)
	id, b64s, _, err := cp.Generate()
	if err != nil {
		return nil, err
	}
	return &CaptchaResult{
		CaptchaID:   id,
		ImageBase64: b64s,
	}, nil
}

// VerifyDirect 直接核验验证码 (一次性消费，验证后立即清除，防止重放攻击)
func VerifyDirect(captchaID, answer string) bool {
	if captchaID == "" || answer == "" {
		return false
	}
	return CaptchaStore.Verify(captchaID, strings.TrimSpace(answer), true)
}

// =========================================================================
// 人机验证 Token 机制 (Captcha Token)
// 适用于：前端通过一次验证码考核后，签发有效期 5 分钟的 Token，
// 供“输入框失焦/实时排重防抖校验”等高频交互免重复弹窗使用。
// 格式: base64(timestamp.ip.random).signature
// =========================================================================

const TokenValidDuration = 5 * time.Minute

func getSecret() []byte {
	s := setting.AppSetting.JwtSecret
	if s == "" {
		s = "message-nest-captcha-secret"
	}
	return []byte(s)
}

// GenerateVerificationToken 验证成功后，签发一个人机验证 Token
func GenerateVerificationToken(clientIP string) string {
	now := time.Now().Unix()
	payload := fmt.Sprintf("%d|%s|%d", now, clientIP, time.Now().UnixNano())
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte(payload))

	h := hmac.New(sha256.New, getSecret())
	h.Write([]byte(payloadB64))
	sig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return fmt.Sprintf("%s.%s", payloadB64, sig)
}

// VerifyToken 检验人机验证 Token 是否合法且在有效期内 (严格校验客户端 IP 与时间窗口)
func VerifyToken(token string, clientIP string) (bool, string) {
	if token == "" {
		return false, "人机验证 Token 不能为空"
	}
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false, "人机验证 Token 格式错误"
	}
	payloadB64, sig := parts[0], parts[1]

	// 1. 签名防篡改校验
	h := hmac.New(sha256.New, getSecret())
	h.Write([]byte(payloadB64))
	expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return false, "人机验证 Token 签名无效"
	}

	// 2. 解码 Payload
	rawPayload, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return false, "人机验证 Token 解码失败"
	}
	fields := strings.Split(string(rawPayload), "|")
	if len(fields) < 2 {
		return false, "人机验证 Token 载荷损坏"
	}

	issueTime, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return false, "人机验证 Token 时间戳解析失败"
	}
	tokenIP := fields[1]

	// 3. 有效期校验 (默认 5 分钟)
	if time.Now().Unix()-issueTime > int64(TokenValidDuration.Seconds()) {
		return false, "人机验证 Token 已过期，请重新验证"
	}

	// 4. IP 绑定校验 (防止 Token 被盗用给其他 IP 刷接口，除非是无 IP 代理模式)
	if clientIP != "" && tokenIP != "" && clientIP != tokenIP {
		return false, "人机验证 Token 客户端环境异常"
	}

	return true, ""
}
