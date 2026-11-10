package channels

import (
	"fmt"
	"message-nest/models"
	"reflect"
)

const AliyunSMSStructuredParamMode = "structured"

// BuildAliyunTemplateParams creates the single TemplateParam map used by the
// production Aliyun channel for both legacy content instances and structured
// V2 template instances.
func BuildAliyunTemplateParams(config models.InsAliyunSMSConfig, content string, placeholders, extra map[string]interface{}) (map[string]interface{}, error) {
	switch config.TemplateParamMode {
	case "", "content":
		params := map[string]interface{}{"content": content}
		for key, value := range extra {
			params[key] = value
		}
		return params, nil
	case AliyunSMSStructuredParamMode:
		return buildStructuredAliyunTemplateParams(config, placeholders)
	default:
		return nil, fmt.Errorf("不支持的阿里云短信参数模式: %q", config.TemplateParamMode)
	}
}

func buildStructuredAliyunTemplateParams(config models.InsAliyunSMSConfig, placeholders map[string]interface{}) (map[string]interface{}, error) {
	if err := config.ValidateTemplateParamConfig(); err != nil {
		return nil, err
	}

	params := make(map[string]interface{}, len(config.SelectedParams))
	for _, sourceName := range config.SelectedParams {
		value, exists := placeholders[sourceName]
		if !exists {
			return nil, fmt.Errorf("模板参数 %q 缺少运行时值", sourceName)
		}
		stringValue, err := stringifyAliyunTemplateValue(value)
		if err != nil {
			return nil, fmt.Errorf("模板参数 %q 无效: %w", sourceName, err)
		}

		targetName := sourceName
		if override, exists := config.ParamNameOverrides[sourceName]; exists {
			targetName = override
		}
		params[targetName] = stringValue
	}
	return params, nil
}

func stringifyAliyunTemplateValue(value interface{}) (string, error) {
	if value == nil {
		return "", fmt.Errorf("值不能为 null")
	}
	valueKind := reflect.ValueOf(value).Kind()
	switch valueKind {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return fmt.Sprint(value), nil
	default:
		return "", fmt.Errorf("只支持字符串、数字或布尔值")
	}
}
