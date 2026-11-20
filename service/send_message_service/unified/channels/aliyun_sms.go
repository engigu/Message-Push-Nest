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
	Code          string
	Message       string
	RequestID     string
	BizID         string
	RequestParams AliyunSMSRequestParams
}

// AliyunSMSRequestParams is the exact non-secret parameter set passed to the
// Alibaba Cloud SDK. The SDK serializes and signs the RPC request itself.
type AliyunSMSRequestParams struct {
	RegionID      string `json:"RegionId"`
	Scheme        string `json:"Scheme"`
	PhoneNumbers  string `json:"PhoneNumbers"`
	SignName      string `json:"SignName"`
	TemplateCode  string `json:"TemplateCode"`
	TemplateParam string `json:"TemplateParam"`
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

	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = config.PhoneNumber
	request.SignName = auth.SignName
	request.TemplateCode = config.TemplateCode
	request.TemplateParam = string(templateParamJSON)
	regionID := auth.RegionId
	if regionID == "" {
		regionID = "cn-hangzhou"
	}
	result := AliyunSMSSendResult{RequestParams: AliyunSMSRequestParams{
		RegionID: regionID, Scheme: request.Scheme, PhoneNumbers: request.PhoneNumbers,
		SignName: request.SignName, TemplateCode: request.TemplateCode, TemplateParam: request.TemplateParam,
	}}

	// Invalid configuration is rejected before constructing an SDK client.
	client, err := c.createClient(auth.AccessKeyId, auth.AccessKeySecret, auth.RegionId)
	if err != nil {
		return result, fmt.Errorf("创建阿里云短信客户端失败: %w", err)
	}

	response, err := client.SendSms(request)
	if err != nil {
		return result, err
	}
	if response == nil {
		return result, fmt.Errorf("阿里云短信响应为空")
	}
	result.Code = response.Code
	result.Message = response.Message
	result.RequestID = response.RequestId
	result.BizID = response.BizId
	return result, nil
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
