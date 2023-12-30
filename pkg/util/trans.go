package util

import (
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_trans "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
	"strings"
)

//var trans ut.Translator

func TransInit() (trans ut.Translator, validate *validator.Validate) {
	// 创建翻译器
	zhTrans := zh.New() // 中文转换器
	enTrans := en.New() // 因为转换器

	uni := ut.New(zhTrans, zhTrans, enTrans) // 创建一个通用转换器

	curLocales := "zh"                       // 设置当前语言类型
	trans, _ = uni.GetTranslator(curLocales) // 获取对应语言的转换器

	validate = validator.New() // 创建验证器
	_ = zh_trans.RegisterDefaultTranslations(validate, trans)

	//switch curLocales {
	//case "zh":
	//	// 内置tag注册 中文翻译器
	//	_ = zh_trans.RegisterDefaultTranslations(validate, trans)
	//case "en":
	//	// 内置tag注册 英文翻译器
	//	_ = en_trans.RegisterDefaultTranslations(validate, trans)
	//}

	// 注册 RegisterTagNameFunc
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("label"), ",", 2)[0]
		if name == "-" {
			return ""
		}

		return name
	})
	return trans, validate
}

var Trans, CustomerValidate = TransInit()
