package send_ins_service

import (
	"encoding/json"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/constant"
)

// InsValidator 实例验证接口
type InsValidator interface {
	Validate(configJson string) (string, interface{})
}

// 渠道实例验证器注册表
var insValidatorRegistry = map[string]func() InsValidator{
	constant.MessageTypeEmail:           func() InsValidator { return &InsEmailValidator{} },
	constant.MessageTypeDtalk:           func() InsValidator { return &InsDtalkValidator{} },
	constant.MessageTypeQyWeiXin:        func() InsValidator { return &InsQyWeiXinValidator{} },
	constant.MessageTypeQyWeiXinApp:     func() InsValidator { return &InsQyWeiXinAppValidator{} },
	constant.MessageTypeFeishu:          func() InsValidator { return &InsFeishuValidator{} },
	constant.MessageTypeMessageNest:     func() InsValidator { return &InsMessageNestValidator{} },
	constant.MessageTypeCustom:          func() InsValidator { return &InsCustomValidator{} },
	constant.MessageTypeWeChatOFAccount: func() InsValidator { return &InsWeChatOFAccountValidator{} },
	constant.MessageTypeAliyunSMS:       func() InsValidator { return &InsAliyunSMSValidator{} },
	constant.MessageTypeTelegram:        func() InsValidator { return &InsTelegramValidator{} },
	constant.MessageTypeBark:            func() InsValidator { return &InsBarkValidator{} },
	constant.MessageTypePushMe:          func() InsValidator { return &InsPushMeValidator{} },
	constant.MessageTypeNtfy:            func() InsValidator { return &InsNtfyValidator{} },
	constant.MessageTypeGotify:          func() InsValidator { return &InsGotifyValidator{} },
}

// checkAllowMultiRecip 检查是否启用了动态接收者模式
func checkAllowMultiRecip(configJson string) bool {
	var configMap map[string]interface{}
	err := json.Unmarshal([]byte(configJson), &configMap)
	if err != nil {
		return false
	}
	allowMultiRecip, exists := configMap["allowMultiRecip"].(bool)
	return exists && allowMultiRecip
}

// InsEmailValidator 邮箱配置校验
type InsEmailValidator struct{}

func (v *InsEmailValidator) Validate(configJson string) (string, interface{}) {
	var empty interface{}
	var emailConfig models.InsEmailConfig
	err := json.Unmarshal([]byte(configJson), &emailConfig)
	if err != nil {
		return "邮箱auth反序列化失败！", empty
	}

	// checkAllowMultiRecip 为真代表动态模式，to_account 可以为空（但如果有值也要验证格式）
	if checkAllowMultiRecip(configJson) {
		if emailConfig.ToAccount == "" {
			return "", emailConfig
		}
		_, Msg := app.CommonPlaygroundValid(emailConfig)
		return Msg, emailConfig
	}

	// 固定模式：必须验证to_account
	_, Msg := app.CommonPlaygroundValid(emailConfig)
	return Msg, emailConfig
}

// InsDtalkValidator 钉钉配置校验
type InsDtalkValidator struct{}

func (v *InsDtalkValidator) Validate(configJson string) (string, interface{}) {
	var Config models.InsDtalkConfig
	return "", Config
}

// InsQyWeiXinValidator 企业微信配置校验
type InsQyWeiXinValidator struct{}

func (v *InsQyWeiXinValidator) Validate(configJson string) (string, interface{}) {
	var Config models.InsQyWeiXinConfig
	return "", Config
}

// InsQyWeiXinAppValidator 企业微信自建应用配置校验
type InsQyWeiXinAppValidator struct{}

func (v *InsQyWeiXinAppValidator) Validate(configJson string) (string, interface{}) {
	var empty interface{}
	var Config models.InsQyWeiXinAppConfig
	err := json.Unmarshal([]byte(configJson), &Config)
	if err != nil {
		return "企业微信自建应用配置反序列化失败！", empty
	}

	// checkAllowMultiRecip 为真代表动态模式，to_user 可以为空
	if checkAllowMultiRecip(configJson) && Config.ToUser == "" {
		return "", Config
	}

	_, Msg := app.CommonPlaygroundValid(Config)
	return Msg, Config
}

// InsFeishuValidator 飞书配置校验
type InsFeishuValidator struct{}

func (v *InsFeishuValidator) Validate(configJson string) (string, interface{}) {
	var Config models.InsFeishuConfig
	return "", Config
}

// InsMessageNestValidator 托管消息配置校验
type InsMessageNestValidator struct{}

func (v *InsMessageNestValidator) Validate(configJson string) (string, interface{}) {
	var empty interface{}
	var Config models.InsMessageNestConfig
	err := json.Unmarshal([]byte(configJson), &Config)
	if err != nil {
		return "托管消息配置反序列化失败！", empty
	}
	return "", Config
}

// InsCustomValidator 自定义webhook配置校验
type InsCustomValidator struct{}

func (v *InsCustomValidator) Validate(configJson string) (string, interface{}) {
	var empty interface{}
	var Config models.InsCustomConfig
	err := json.Unmarshal([]byte(configJson), &Config)
	if err != nil {
		return "自定义webhook反序列化失败！", empty
	}
	_, Msg := app.CommonPlaygroundValid(Config)
	return Msg, Config
}

// InsWeChatOFAccountValidator 微信公众号配置校验
type InsWeChatOFAccountValidator struct{}

func (v *InsWeChatOFAccountValidator) Validate(configJson string) (string, interface{}) {
	var empty interface{}
	var Config models.InsWeChatAccountConfig
	err := json.Unmarshal([]byte(configJson), &Config)
	if err != nil {
		return "微信公众号发送配置反序列化失败！", empty
	}

	// checkAllowMultiRecip 为真代表动态模式，to_account 可以为空
	if checkAllowMultiRecip(configJson) && Config.ToAccount == "" {
		return "", Config
	}

	// 固定模式或有to_account时，进行常规验证
	_, Msg := app.CommonPlaygroundValid(Config)
	return Msg, Config
}

// InsAliyunSMSValidator 阿里云短信配置校验
type InsAliyunSMSValidator struct{}

func (v *InsAliyunSMSValidator) Validate(configJson string) (string, interface{}) {
	var empty interface{}
	var Config models.InsAliyunSMSConfig
	err := json.Unmarshal([]byte(configJson), &Config)
	if err != nil {
		return "阿里云短信配置反序列化失败！", empty
	}

	// checkAllowMultiRecip 为真代表动态模式，phone_number 可以为空，但 template_code 仍需验证
	if checkAllowMultiRecip(configJson) && Config.PhoneNumber == "" {
		if Config.TemplateCode == "" {
			return "短信模板CODE不能为空", empty
		}
		return "", Config
	}

	// 固定模式或有phone_number时，进行常规验证
	_, Msg := app.CommonPlaygroundValid(Config)
	return Msg, Config
}

// InsTelegramValidator Telegram配置校验
type InsTelegramValidator struct{}

func (v *InsTelegramValidator) Validate(configJson string) (string, interface{}) {
	var Config models.InsTelegramConfig
	return "", Config
}

// InsBarkValidator Bark配置校验
type InsBarkValidator struct{}

func (v *InsBarkValidator) Validate(configJson string) (string, interface{}) {
	var Config models.InsBarkConfig
	return "", Config
}

// InsPushMeValidator PushMe配置校验
type InsPushMeValidator struct{}

func (v *InsPushMeValidator) Validate(configJson string) (string, interface{}) {
	var Config models.InsPushMeConfig
	return "", Config
}

// InsNtfyValidator Ntfy配置校验
type InsNtfyValidator struct{}

func (v *InsNtfyValidator) Validate(configJson string) (string, interface{}) {
	var Config models.InsNtfyConfig
	return "", Config
}

// InsGotifyValidator Gotify配置校验
type InsGotifyValidator struct{}

func (v *InsGotifyValidator) Validate(configJson string) (string, interface{}) {
	var Config models.InsGotifyConfig
	return "", Config
}
