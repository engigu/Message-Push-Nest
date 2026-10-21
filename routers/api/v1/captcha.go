package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"message-nest/pkg/app"
	"message-nest/pkg/captcha"
)

// GetCaptcha 获取人机验证码 (支持 math 算术与 digit 数字，默认 math)
func GetCaptcha(c *gin.Context) {
	appG := app.Gin{C: c}
	captchaType := c.DefaultQuery("type", "math")

	var (
		res *captcha.CaptchaResult
		err error
	)

	if captchaType == "digit" {
		res, err = captcha.GenerateDigitCaptcha()
	} else {
		res, err = captcha.GenerateMathCaptcha()
	}

	if err != nil {
		appG.CResponse(http.StatusInternalServerError, "生成验证码失败: "+err.Error(), nil)
		return
	}

	appG.CResponse(http.StatusOK, "获取成功", gin.H{
		"captcha_id":    res.CaptchaID,
		"captcha_image": res.ImageBase64,
	})
}

type VerifyCaptchaReq struct {
	CaptchaID   string `json:"captcha_id" validate:"required" label:"验证码ID"`
	CaptchaCode string `json:"captcha_code" validate:"required" label:"验证码答案"`
}

// VerifyCaptchaAndGetToken 校验验证码并签发人机验证 Token (有效期 5 分钟，防刷接口复用)
func VerifyCaptchaAndGetToken(c *gin.Context) {
	appG := app.Gin{C: c}
	var req VerifyCaptchaReq
	if err := c.ShouldBindJSON(&req); err != nil {
		appG.CResponse(http.StatusBadRequest, "参数格式错误", nil)
		return
	}

	if !captcha.VerifyDirect(req.CaptchaID, req.CaptchaCode) {
		appG.CResponse(http.StatusBadRequest, "验证码错误或已过期，请点击图片重试", nil)
		return
	}

	// 验证成功，签发人机验证 Token (绑定当前客户端 IP)
	token := captcha.GenerateVerificationToken(c.ClientIP())
	appG.CResponse(http.StatusOK, "人机验证通过", gin.H{
		"captcha_token": token,
	})
}
