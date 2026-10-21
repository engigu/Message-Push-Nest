package captcha

import (
	"bytes"
	"crypto/hmac"
	cryptorand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	mathrand "math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"message-nest/pkg/util"
)

type captchaItem struct {
	value     string
	createdAt time.Time
}

// MemoryStoreWithCron 实现受控的验证码内存存储，
// 统一交由系统全局 Cron 周期性调度清理过期键，杜绝后台不可控孤儿协程。
type MemoryStoreWithCron struct {
	sync.RWMutex
	data       map[string]captchaItem
	expiration time.Duration
	maxEntries int
}

// NewMemoryStoreWithCron 创建接入 Cron 调度的轻量内存验证码存储器
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

// CaptchaStore 采用接入 cron 调度的轻量内存存储，默认最多保留 10240 个，过期时间 5 分钟
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

// GenerateDigitCaptcha 生成纯数字验证码 (4 位数字，高强度复合波浪扭曲，零堆常驻)
func GenerateDigitCaptcha() (*CaptchaResult, error) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	digits := fmt.Sprintf("%04d", r.Intn(10000))
	id := util.GenerateUniqueID()

	_ = CaptchaStore.Set(id, digits)
	b64, err := renderCaptchaImage(digits)
	if err != nil {
		return nil, err
	}

	return &CaptchaResult{
		CaptchaID:   id,
		ImageBase64: b64,
	}, nil
}

// GenerateMathCaptcha 生成高安全算术验证码 (如 18 - 4 = ?，零第三方字库内存占用)
func GenerateMathCaptcha() (*CaptchaResult, error) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	opType := r.Intn(3) // 0: 加法, 1: 减法, 2: 乘法

	var a, b, result int
	var opStr string

	switch opType {
	case 0:
		a = r.Intn(15) + 1
		b = r.Intn(15) + 1
		opStr = "+"
		result = a + b
	case 1:
		a = r.Intn(20) + 5
		b = r.Intn(a-1) + 1 // 保证结果为正数
		opStr = "-"
		result = a - b
	case 2:
		a = r.Intn(9) + 1
		b = r.Intn(9) + 1
		opStr = "x"
		result = a * b
	}

	expr := fmt.Sprintf("%d %s %d = ?", a, opStr, b)
	id := util.GenerateUniqueID()

	_ = CaptchaStore.Set(id, strconv.Itoa(result))
	b64, err := renderCaptchaImage(expr)
	if err != nil {
		return nil, err
	}

	return &CaptchaResult{
		CaptchaID:   id,
		ImageBase64: b64,
	}, nil
}

// renderCaptchaImage 纯标准库渲染轻量复杂验证码图片：
// 采用微分同胚平滑水波形变（保证字形拓扑完整、不破碎撕裂）、字符中心剪切倾斜、
// 双线性抗锯齿采样、穿透式半透明波浪干扰线、细密噪点与防抠图底纹。
func renderCaptchaImage(text string) (string, error) {
	r := mathrand.New(mathrand.NewSource(time.Now().UnixNano()))
	width, height := 140, 42
	rawImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// 1. 底板浅色 #F8FAFC
	bgColor := color.RGBA{R: 248, G: 250, B: 252, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			rawImg.Set(x, y, bgColor)
		}
	}

	// 2. 底纹浅色随机波浪线 (2 条，半透明淡灰蓝，防简单二值化分割)
	for i := 0; i < 2; i++ {
		bAmp := 2.5 + r.Float64()*2.5
		bPeriod := 35.0 + r.Float64()*20.0
		bPhase := r.Float64() * math.Pi * 2
		bBaseY := 10 + r.Intn(height-20)
		bCol := color.RGBA{
			R: uint8(215 + r.Intn(20)),
			G: uint8(225 + r.Intn(15)),
			B: uint8(235 + r.Intn(15)),
			A: 160,
		}
		for x := 0; x < width; x++ {
			wy := bBaseY + int(bAmp*math.Sin(2*math.Pi*float64(x)/bPeriod+bPhase))
			if wy >= 0 && wy < height {
				rawImg.Set(x, wy, bCol)
			}
		}
	}

	// 3. 文字深色高对比色彩池 (保证文字主体醒目易识)
	charColors := []color.RGBA{
		{R: 20, G: 65, B: 185, A: 255},  // 经典深蓝
		{R: 195, G: 28, B: 38, A: 255},  // 鲜红
		{R: 10, G: 125, B: 60, A: 255},  // 深林绿
		{R: 125, G: 25, B: 185, A: 255}, // 蓝紫
		{R: 195, G: 80, B: 12, A: 255},  // 暖橙褐
		{R: 12, G: 110, B: 125, A: 255}, // 黛绿
	}

	scale := 2
	totalWidth := 0
	for _, ch := range text {
		if ch == ' ' {
			totalWidth += 6
		} else {
			totalWidth += 13
		}
	}
	startX := (width - totalWidth) / 2
	if startX < 6 {
		startX = 6
	}
	startY := (height - 13*scale) / 2

	// 4. 逐字符绘制：字符中心倾斜 (Shear) + 垂直微扰动 (Jitter) + 2x2 笔划加粗
	currX := startX
	for i, ch := range text {
		if ch == ' ' {
			currX += 6
			continue
		}

		charImg := image.NewRGBA(image.Rect(0, 0, 8, 14))
		col := charColors[i%len(charColors)]
		d := &font.Drawer{
			Dst:  charImg,
			Src:  image.NewUniform(col),
			Face: basicfont.Face7x13,
			Dot:  fixed.Point26_6{X: fixed.I(0), Y: fixed.I(11)},
		}
		d.DrawString(string(ch))

		shearX := (r.Float64() - 0.5) * 0.35 // 字符自身倾斜度 (-0.17 ~ +0.17)
		yJitter := r.Intn(5) - 2             // -2 ~ +2 px 错落抖动

		for cy := 0; cy < 13; cy++ {
			for cx := 0; cx < 7; cx++ {
				_, _, _, a := charImg.At(cx, cy).RGBA()
				if a > 0 {
					shearOffset := int(float64(cy-6) * shearX * float64(scale))
					for dy := 0; dy < scale; dy++ {
						for dx := 0; dx < scale; dx++ {
							px := currX + cx*scale + dx + shearOffset
							py := startY + cy*scale + dy + yJitter
							if px >= 0 && px < width && py >= 0 && py < height {
								rawImg.Set(px, py, col)
								// 形态学加粗，确保形变后字迹厚实饱满
								if px+1 < width {
									rawImg.Set(px+1, py, col)
								}
								if py+1 < height {
									rawImg.Set(px, py+1, col)
								}
							}
						}
					}
				}
			}
		}
		currX += 13
	}

	// 5. 微分同胚复合正弦平滑水波形变 (X 依赖 Y，Y 依赖 X，无撕裂破折，平滑自然流线)
	distorted := image.NewRGBA(image.Rect(0, 0, width, height))

	// Y 轴波动 (波长 45~65px，波幅 3.0~4.5px)
	ampY1 := 2.8 + r.Float64()*1.4
	periodX1 := 46.0 + r.Float64()*20.0
	phaseX1 := r.Float64() * math.Pi * 2

	// X 轴波动 (波长 28~40px，波幅 1.5~2.5px)
	ampX1 := 1.6 + r.Float64()*1.0
	periodY1 := 30.0 + r.Float64()*12.0
	phaseY1 := r.Float64() * math.Pi * 2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := float64(x) + ampX1*math.Sin(2*math.Pi*float64(y)/periodY1+phaseY1)
			srcY := float64(y) + ampY1*math.Sin(2*math.Pi*float64(x)/periodX1+phaseX1)
			distorted.Set(x, y, sampleBilinear(rawImg, srcX, srcY, bgColor))
		}
	}

	// 6. 叠加 2 条半透明贯穿波浪线 (半透明度 A: 120~160，既割裂 OCR 连通域，又不遮挡人眼识读)
	for i := 0; i < 2; i++ {
		waveAmp := 3.5 + r.Float64()*2.5
		wavePeriod := 40.0 + r.Float64()*20.0
		wavePhase := r.Float64() * math.Pi * 2
		baseY := 10 + r.Intn(height-20)
		lineCol := charColors[r.Intn(len(charColors))]
		lineCol.A = uint8(120 + r.Intn(45))

		for x := 0; x < width; x++ {
			wy := baseY + int(waveAmp*math.Sin(2*math.Pi*float64(x)/wavePeriod+wavePhase))
			if wy >= 0 && wy < height {
				distorted.Set(x, wy, blendPixel(distorted.RGBAAt(x, wy), lineCol))
			}
			if wy+1 >= 0 && wy+1 < height {
				distorted.Set(x, wy+1, blendPixel(distorted.RGBAAt(x, wy+1), lineCol))
			}
		}
	}

	// 7. 叠加 3 条随机短斜线干扰 (半透明，长度 15~25px)
	for i := 0; i < 3; i++ {
		x0 := r.Intn(width)
		y0 := r.Intn(height)
		x1 := x0 + r.Intn(30) - 15
		y1 := y0 + r.Intn(20) - 10
		cutCol := charColors[r.Intn(len(charColors))]
		cutCol.A = uint8(110 + r.Intn(50))
		drawBlendedLine(distorted, x0, y0, x1, y1, cutCol)
	}

	// 8. 散布 75 个随机彩色微噪点
	for i := 0; i < 75; i++ {
		dotCol := color.RGBA{
			R: uint8(70 + r.Intn(160)),
			G: uint8(70 + r.Intn(160)),
			B: uint8(70 + r.Intn(160)),
			A: uint8(130 + r.Intn(80)),
		}
		rx, ry := r.Intn(width), r.Intn(height)
		distorted.Set(rx, ry, blendPixel(distorted.RGBAAt(rx, ry), dotCol))
	}

	// 9. 精致微边框包裹 #E2E8F0
	borderColor := color.RGBA{R: 226, G: 232, B: 240, A: 255}
	for x := 0; x < width; x++ {
		distorted.Set(x, 0, borderColor)
		distorted.Set(x, height-1, borderColor)
	}
	for y := 0; y < height; y++ {
		distorted.Set(0, y, borderColor)
		distorted.Set(width-1, y, borderColor)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, distorted); err != nil {
		return "", err
	}

	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// blendPixel 将带有透明度的源像素与底层像素进行 Alpha 混合
func blendPixel(dst color.RGBA, src color.RGBA) color.RGBA {
	sa := float64(src.A) / 255.0
	da := float64(dst.A) / 255.0 * (1.0 - sa)
	outA := sa + da
	if outA == 0 {
		return dst
	}
	r := uint8((float64(src.R)*sa + float64(dst.R)*da) / outA)
	g := uint8((float64(src.G)*sa + float64(dst.G)*da) / outA)
	b := uint8((float64(src.B)*sa + float64(dst.B)*da) / outA)
	return color.RGBA{R: r, G: g, B: b, A: uint8(outA * 255.0)}
}

// drawBlendedLine 带有 Alpha 混合的 Bresenham 划线
func drawBlendedLine(img *image.RGBA, x0, y0, x1, y1 int, col color.RGBA) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		if x0 >= 0 && x0 < img.Bounds().Dx() && y0 >= 0 && y0 < img.Bounds().Dy() {
			img.Set(x0, y0, blendPixel(img.RGBAAt(x0, y0), col))
		}
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// sampleBilinear 双线性插值平滑采样，消除网格波浪扭曲后的像素块与锯齿感
func sampleBilinear(img *image.RGBA, x, y float64, bg color.RGBA) color.RGBA {
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := x0 + 1
	y1 := y0 + 1

	fx := x - float64(x0)
	fy := y - float64(y0)

	getPixel := func(px, py int) color.RGBA {
		if px < 0 || px >= img.Bounds().Dx() || py < 0 || py >= img.Bounds().Dy() {
			return bg
		}
		return img.RGBAAt(px, py)
	}

	c00 := getPixel(x0, y0)
	c10 := getPixel(x1, y0)
	c01 := getPixel(x0, y1)
	c11 := getPixel(x1, y1)

	w00 := (1.0 - fx) * (1.0 - fy)
	w10 := fx * (1.0 - fy)
	w01 := (1.0 - fx) * fy
	w11 := fx * fy

	r := uint8(float64(c00.R)*w00 + float64(c10.R)*w10 + float64(c01.R)*w01 + float64(c11.R)*w11)
	g := uint8(float64(c00.G)*w00 + float64(c10.G)*w10 + float64(c01.G)*w01 + float64(c11.G)*w11)
	b := uint8(float64(c00.B)*w00 + float64(c10.B)*w10 + float64(c01.B)*w01 + float64(c11.B)*w11)
	a := uint8(float64(c00.A)*w00 + float64(c10.A)*w10 + float64(c01.A)*w01 + float64(c11.A)*w11)

	return color.RGBA{R: r, G: g, B: b, A: a}
}

// drawLine 快速 Bresenham 划线算法
func drawLine(img *image.RGBA, x0, y0, x1, y1 int, col color.Color) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx := -1
	if x0 < x1 {
		sx = 1
	}
	sy := -1
	if y0 < y1 {
		sy = 1
	}
	err := dx - dy

	for {
		img.Set(x0, y0, col)
		if x0 == x1 && y0 == y1 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
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

var (
	captchaSecretOnce sync.Once
	dynamicSecret     []byte
)

func getSecret() []byte {
	// 在进程生命周期内通过 CSPRNG 生成唯一的随机 32 字节高强度密钥
	// 与登录态 JWT 密钥完全隔离，杜绝硬编码与签名伪造隐患，在单次运行周期内保证签名与校验一致
	captchaSecretOnce.Do(func() {
		buf := make([]byte, 32)
		if _, err := cryptorand.Read(buf); err != nil {
			dynamicSecret = []byte(util.GenerateUniqueID() + "-captcha-secret")
		} else {
			dynamicSecret = buf
		}
	})
	return dynamicSecret
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
