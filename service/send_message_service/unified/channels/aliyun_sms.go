package channels

import (
	"encoding/json"
	"fmt"
	"message-nest/models"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_way_service"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

type aliyunSMSClient interface {
	SendSms(*dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error)
}

type aliyunSMSClientFactory func(accessKeyID, accessKeySecret, regionID string) (aliyunSMSClient, error)

type AliyunSMSChannel struct {
	*BaseChannel
	clientFactory aliyunSMSClientFactory
}

func NewAliyunSMSChannel() *AliyunSMSChannel {
	return &AliyunSMSChannel{
		BaseChannel: NewBaseChannel(MessageTypeAliyunSMS, []string{FormatTypeText}),
		clientFactory: func(accessKeyID, accessKeySecret, regionID string) (aliyunSMSClient, error) {
			return dysmsapi.NewClientWithAccessKey(regionID, accessKeyID, accessKeySecret)
		},
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

	// 格式化内容
	_, formattedContent, err := c.FormatContent(content)
	if err != nil {
		return "", err.Error()
	}

	// 准备模板参数。legacy content 和结构化模板共用这一构造入口。
	templateParam, err := BuildAliyunTemplateParams(config, formattedContent, content.TemplateParams, content.Extra)
	if err != nil {
		return "", err.Error()
	}

	templateParamJSON, err := json.Marshal(templateParam)
	if err != nil {
		return "", fmt.Sprintf("序列化阿里云短信模板参数失败: %s", err.Error())
	}

	// 客户端在参数构造和校验成功后才创建，异常配置不会触达 SDK。
	client, err := c.createClient(auth.AccessKeyId, auth.AccessKeySecret, auth.RegionId)
	if err != nil {
		return "", fmt.Sprintf("创建阿里云短信客户端失败: %s", err.Error())
	}

	// 创建请求对象
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https" // 使用HTTPS协议

	// 设置请求参数
	request.PhoneNumbers = config.PhoneNumber         // 接收短信的手机号码
	request.SignName = auth.SignName                  // 短信签名名称
	request.TemplateCode = config.TemplateCode        // 短信模板ID
	request.TemplateParam = string(templateParamJSON) // 短信模板变量对应的实际值，JSON格式

	// 发送短信
	response, err := client.SendSms(request)
	if err != nil {
		return "", fmt.Sprintf("发送短信失败: %s", err.Error())
	}

	// 处理响应
	if response.Code != "OK" {
		return "", fmt.Sprintf("发送失败: %s - %s", response.Code, response.Message)
	}

	return fmt.Sprintf("RequestId: %s, BizId: %s", response.RequestId, response.BizId), ""
}

// createClient 创建阿里云短信客户端
func (c *AliyunSMSChannel) createClient(accessKeyId, accessKeySecret, regionId string) (aliyunSMSClient, error) {
	if regionId == "" {
		regionId = "cn-hangzhou"
	}
	if c.clientFactory == nil {
		return nil, fmt.Errorf("阿里云短信客户端未配置")
	}
	return c.clientFactory(accessKeyId, accessKeySecret, regionId)
}
