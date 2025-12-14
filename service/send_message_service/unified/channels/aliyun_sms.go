package channels

import (
	"encoding/json"
	"fmt"
	"message-nest/models"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type AliyunSMSChannel struct{ *BaseChannel }

func NewAliyunSMSChannel() *AliyunSMSChannel {
	return &AliyunSMSChannel{BaseChannel: NewBaseChannel(MessageTypeAliyunSMS, []string{FormatTypeText})}
}

func (c *AliyunSMSChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(send_way_service.WayDetailAliyunSMS)
	if !ok {
		return "", "类型转换失败"
	}

	insService := send_ins_service.SendTaskInsService{}
	errStr, configInterface := insService.ValidateDiffIns(ins)
	if errStr != "" {
		return "", errStr
	}

	config, ok := configInterface.(models.InsAliyunSMSConfig)
	if !ok {
		return "", "阿里云短信config校验失败"
	}

	// 格式化内容
	_, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}

	// 创建阿里云短信客户端
	client, err := c.createClient(auth.AccessKeyId, auth.AccessKeySecret)
	if err != nil {
		return "", fmt.Sprintf("创建阿里云短信客户端失败: %s", err.Error())
	}

	// 准备模板参数
	templateParam := map[string]interface{}{
		"content": formattedContent,
	}
	if content.Extra != nil {
		for k, v := range content.Extra {
			templateParam[k] = v
		}
	}

	templateParamJSON, _ := json.Marshal(templateParam)

	// 发送短信
	sendSmsRequest := &dysmsapi.SendSmsRequest{
		PhoneNumbers:  tea.String(config.PhoneNumber),
		SignName:      tea.String(auth.SignName),
		TemplateCode:  tea.String(config.TemplateCode),
		TemplateParam: tea.String(string(templateParamJSON)),
	}

	runtime := &util.RuntimeOptions{}
	tryErr := func() error {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				// 处理panic
			}
		}()

		response, _err := client.SendSmsWithOptions(sendSmsRequest, runtime)
		if _err != nil {
			return _err
		}

		if response.Body.Code != nil && *response.Body.Code != "OK" {
			return fmt.Errorf("发送失败: %s - %s", tea.StringValue(response.Body.Code), tea.StringValue(response.Body.Message))
		}

		return nil
	}()

	if tryErr != nil {
		return "", tryErr.Error()
	}

	return "", ""
}

// createClient 创建阿里云短信客户端
func (c *AliyunSMSChannel) createClient(accessKeyId, accessKeySecret string) (*dysmsapi.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(accessKeyId),
		AccessKeySecret: tea.String(accessKeySecret),
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}
	return dysmsapi.NewClient(config)
}
