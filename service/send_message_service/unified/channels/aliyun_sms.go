package channels

import (
	"encoding/json"
	"fmt"
	"message-nest/models"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

type AliyunSMSClient interface {
	SendSms(*dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error)
}

type AliyunSMSClientFactory func(accessKeyID, accessKeySecret, regionID string) (AliyunSMSClient, error)

type AliyunSMSSendResult struct {
	Code      string
	Message   string
	RequestID string
	BizID     string
}

type AliyunSMSChannel struct {
	*BaseChannel
	clientFactory AliyunSMSClientFactory
}

func NewAliyunSMSChannel() *AliyunSMSChannel {
	return NewAliyunSMSChannelWithClientFactory(func(accessKeyID, accessKeySecret, regionID string) (AliyunSMSClient, error) {
		return dysmsapi.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret)
	})
}

// NewAliyunSMSChannelWithClientFactory injects an SDK client for tests while
// keeping production and test requests on the same TemplateParam path.
func NewAliyunSMSChannelWithClientFactory(factory AliyunSMSClientFactory) *AliyunSMSChannel {
	return &AliyunSMSChannel{
		BaseChannel:   NewBaseChannel(MessageTypeAliyunSMS, []string{FormatTypeText}),
		clientFactory: factory,
	}
}

func (c *AliyunSMSChannel) SendUnified(msgObj interface{}, ins models.SendTasksIns, content *UnifiedMessageContent) (string, string) {
	auth, ok := msgObj.(*send_way_service.WayDetailAliyunSMS)
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

	_, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}

	result, err := c.SendTemplate(auth, config, formattedContent, content.TemplateParams, content.Extra)
	if err != nil {
		return "", fmt.Sprintf("发送短信失败: %s", err.Error())
	}
	if result.Code != "OK" {
		return "", fmt.Sprintf("发送失败: %s - %s", result.Code, result.Message)
	}
	return fmt.Sprintf("RequestId: %s, BizId: %s", result.RequestID, result.BizID), ""
}

// SendTemplate synchronously sends one SMS and returns the provider response.
// Both the normal channel and the protected test API call this method.
func (c *AliyunSMSChannel) SendTemplate(auth *send_way_service.WayDetailAliyunSMS, config models.InsAliyunSMSConfig, content string, placeholders, extra map[string]interface{}) (AliyunSMSSendResult, error) {
	// Legacy and structured modes share the same parameter builder.
	templateParam, err := BuildAliyunTemplateParams(config, content, placeholders, extra)
	if err != nil {
		return AliyunSMSSendResult{}, err
	}
	templateParamJSON, err := json.Marshal(templateParam)
	if err != nil {
		return AliyunSMSSendResult{}, fmt.Errorf("序列化阿里云短信模板参数失败: %w", err)
	}

	// Invalid configuration is rejected before constructing an SDK client.
	client, err := c.createClient(auth.AccessKeyId, auth.AccessKeySecret, auth.RegionId)
	if err != nil {
		return AliyunSMSSendResult{}, fmt.Errorf("创建阿里云短信客户端失败: %w", err)
	}

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = config.PhoneNumber
	request.SignName = auth.SignName
	request.TemplateCode = config.TemplateCode
	request.TemplateParam = string(templateParamJSON)

	response, err := client.SendSms(request)
	if err != nil {
		return AliyunSMSSendResult{}, err
	}
	if response == nil {
		return AliyunSMSSendResult{}, fmt.Errorf("阿里云短信响应为空")
	}
	return AliyunSMSSendResult{
		Code:      response.Code,
		Message:   response.Message,
		RequestID: response.RequestId,
		BizID:     response.BizId,
	}, nil
}

func (c *AliyunSMSChannel) createClient(accessKeyId, accessKeySecret, regionId string) (AliyunSMSClient, error) {
	if regionId == "" {
		regionId = "cn-hangzhou"
	}
	if c.clientFactory == nil {
		return nil, fmt.Errorf("阿里云短信客户端未配置")
	}
	return c.clientFactory(accessKeyId, accessKeySecret, regionId)
}
