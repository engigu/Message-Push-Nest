package v1

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/logging"
	"message-nest/pkg/setting"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_message_service/unified/channels"
	"message-nest/service/send_way_service"
)

const (
	aliyunSMSRequestLimit       = 64 << 10
	aliyunSMSLimiterTimeout     = 3 * time.Second
	aliyunSMSFingerprintMinSize = 32
)

var testPhonePattern = regexp.MustCompile(`^\+?[0-9]{5,20}$`)

// AliyunSMSChannelSender is the synchronous provider boundary used by the test API.
// The default instance uses the production channel adapter; callers may inject a
// controlled adapter when composing an isolated router.
type AliyunSMSChannelSender interface {
	SendTemplate(*send_way_service.WayDetailAliyunSMS, models.InsAliyunSMSConfig, string, map[string]interface{}, map[string]interface{}) (channels.AliyunSMSSendResult, error)
}

type aliyunProviderResponseError interface {
	ErrorCode() string
	Message() string
	RequestId() string
}

type AliyunSMSInstanceAPI struct {
	limiter       func(context.Context, string, string, string) (models.RateLimitDecision, error)
	channelSender func() AliyunSMSChannelSender
}

func NewAliyunSMSInstanceAPI() *AliyunSMSInstanceAPI {
	return &AliyunSMSInstanceAPI{
		limiter: models.ReserveAliyunSMSTestRate,
		channelSender: func() AliyunSMSChannelSender {
			return channels.NewAliyunSMSChannel()
		},
	}
}

// NewAliyunSMSInstanceAPIWithSender composes an instance API with a supplied
// synchronous sender. It is useful for isolated router integrations that must
// prove request flow without contacting the live provider.
func NewAliyunSMSInstanceAPIWithSender(sender AliyunSMSChannelSender) *AliyunSMSInstanceAPI {
	api := NewAliyunSMSInstanceAPI()
	if sender != nil {
		api.channelSender = func() AliyunSMSChannelSender { return sender }
	}
	return api
}

var defaultAliyunSMSInstanceAPI = NewAliyunSMSInstanceAPI()

func EditAliyunSMSInstanceConfig(c *gin.Context) {
	defaultAliyunSMSInstanceAPI.EditConfig(c)
}

func TestAliyunSMSInstance(c *gin.Context) {
	defaultAliyunSMSInstanceAPI.TestSend(c)
}

type editAliyunSMSConfigRequest struct {
	ID     string          `json:"id"`
	Config json.RawMessage `json:"config"`
}

type testAliyunSMSRequest struct {
	InstanceID   string                 `json:"instance_id,omitempty"`
	TemplateID   string                 `json:"template_id,omitempty"`
	WayID        string                 `json:"way_id,omitempty"`
	Config       json.RawMessage        `json:"config,omitempty"`
	TestPhone    string                 `json:"test_phone"`
	Placeholders map[string]interface{} `json:"placeholders,omitempty"`
}

type testAliyunSMSResponse struct {
	Provider        string                           `json:"provider"`
	Code            string                           `json:"code"`
	Message         string                           `json:"message"`
	Troubleshooting string                           `json:"troubleshooting,omitempty"`
	RequestParams   *channels.AliyunSMSRequestParams `json:"request_params,omitempty"`
	RequestID       string                           `json:"request_id,omitempty"`
	BizID           string                           `json:"biz_id,omitempty"`
	Accepted        bool                             `json:"accepted"`
	DeliveryStatus  string                           `json:"delivery_status"`
	DeliveryNotice  string                           `json:"delivery_notice"`
}

// EditConfig updates only the supported fields of a template-bound Aliyun SMS
// instance. Immutable associations and pre-existing unknown config fields are
// left untouched.
func (h *AliyunSMSInstanceAPI) EditConfig(c *gin.Context) {
	appG := app.Gin{C: c}
	var req editAliyunSMSConfigRequest
	if err := decodeStrictAliyunRequest(c, &req); err != nil {
		appG.CResponse(http.StatusBadRequest, "请求参数错误", nil)
		return
	}
	if req.ID == "" || len(req.Config) == 0 {
		appG.CResponse(http.StatusBadRequest, "实例ID和Config不能为空", nil)
		return
	}

	user, ownerID, ok := getAliyunSMSRequestScope(c)
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}
	ins, err := models.GetTemplateInstanceByIDForOwner(req.ID, ownerID)
	if err != nil {
		logAliyunSMSAudit("edit", user, ownerID, req.ID, "", "", "denied", "", "")
		appG.CResponse(http.StatusNotFound, "阿里云短信模板实例不存在或无权访问", nil)
		return
	}
	template, way, err := loadAliyunSMSResources(ownerID, ins.TemplateID, ins.WayID)
	if err != nil || ins.WayType != channels.MessageTypeAliyunSMS || way.Type != channels.MessageTypeAliyunSMS {
		logAliyunSMSAudit("edit", user, ownerID, ins.ID, ins.TemplateID, ins.WayID, "denied", "", "")
		appG.CResponse(http.StatusNotFound, "阿里云短信模板实例资源不存在或无权访问", nil)
		return
	}

	mergedConfig, configJSON, err := mergeSupportedAliyunConfig([]byte(ins.Config), req.Config)
	if err != nil {
		logAliyunSMSAudit("edit", user, ownerID, ins.ID, ins.TemplateID, ins.WayID, "invalid_config", "", "")
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := validateAliyunConfigForTemplate(mergedConfig, []byte(configJSON), template.Placeholders); err != nil {
		logAliyunSMSAudit("edit", user, ownerID, ins.ID, ins.TemplateID, ins.WayID, "invalid_config", "", "")
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := models.UpdateTemplateInstanceConfig(ins.ID, ownerID, user.UserID, configJSON); err != nil {
		logAliyunSMSAudit("edit", user, ownerID, ins.ID, ins.TemplateID, ins.WayID, "update_failed", "", "")
		appG.CResponse(http.StatusNotFound, "实例不存在或无权访问", nil)
		return
	}
	logAliyunSMSAudit("edit", user, ownerID, ins.ID, ins.TemplateID, ins.WayID, "updated", "", "")
	appG.CResponse(http.StatusOK, "阿里云短信实例配置已更新", editConfigResponseID(ins.ID))
}

// TestSend synchronously calls Aliyun using a mockable channel sender. Draft
// configurations are never persisted and this path does not enter task logs
// or send statistics.
func (h *AliyunSMSInstanceAPI) TestSend(c *gin.Context) {
	appG := app.Gin{C: c}
	var req testAliyunSMSRequest
	if err := decodeStrictAliyunRequest(c, &req); err != nil {
		appG.CResponse(http.StatusBadRequest, "请求参数错误", nil)
		return
	}
	if !testPhonePattern.MatchString(req.TestPhone) {
		appG.CResponse(http.StatusBadRequest, "测试手机号格式无效", nil)
		return
	}
	if containsCredentialField(req.Config) {
		appG.CResponse(http.StatusBadRequest, "测试请求Config不得包含渠道凭据", nil)
		return
	}

	user, ownerID, ok := getAliyunSMSRequestScope(c)
	if !ok {
		appG.CResponse(http.StatusUnauthorized, "用户身份无效", nil)
		return
	}

	instance, template, way, configJSON, err := h.resolveTestResources(ownerID, req)
	if err != nil {
		logAliyunSMSAudit("test", user, ownerID, req.InstanceID, req.TemplateID, req.WayID, "denied_or_invalid", maskAliyunPhone(req.TestPhone), "")
		appG.CResponse(http.StatusNotFound, "模板、实例或阿里云渠道不存在或无权访问", nil)
		return
	}
	if template.Status != "enabled" {
		appG.CResponse(http.StatusBadRequest, "模板已禁用，不能执行发送测试", nil)
		return
	}

	config, err := decodeAliyunSMSConfig(configJSON)
	if err != nil {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "invalid_config", maskAliyunPhone(req.TestPhone), "")
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}
	validationConfigJSON, err := configWithTestPhone(configJSON, req.TestPhone)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "Config格式无效", nil)
		return
	}
	if err := validateAliyunConfigForTemplate(config, validationConfigJSON, template.Placeholders); err != nil {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "invalid_config", maskAliyunPhone(req.TestPhone), "")
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := rejectCredentialConfig(configJSON); err != nil {
		appG.CResponse(http.StatusBadRequest, "Config中不得包含渠道凭据", nil)
		return
	}

	definitions, values, err := resolveAliyunTemplateValues(template.Placeholders, req.Placeholders, config.TemplateParamMode == channels.AliyunSMSStructuredParamMode)
	if err != nil {
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}
	structuredValues := selectDeclaredAliyunValues(definitions, values)
	content := renderAliyunSMSContent(template.TextTemplate, values)
	if _, err := channels.BuildAliyunTemplateParams(config, content, structuredValues, nil); err != nil {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "invalid_params", maskAliyunPhone(req.TestPhone), "")
		appG.CResponse(http.StatusBadRequest, err.Error(), nil)
		return
	}
	config.PhoneNumber = req.TestPhone // temporary effective recipient for validation and the SDK request
	auth, err := validateAliyunWay(way)
	if err != nil {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "invalid_way_auth", maskAliyunPhone(req.TestPhone), "")
		appG.CResponse(http.StatusBadRequest, "阿里云渠道配置无效", nil)
		return
	}

	fingerprintSecret := setting.AppSetting.AliyunSMSTestFingerprintSecret
	if len([]byte(fingerprintSecret)) < aliyunSMSFingerprintMinSize {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "fingerprint_secret_missing", maskAliyunPhone(req.TestPhone), "")
		msg := fmt.Sprintf("阿里云短信测试未启用：当前服务实例未加载有效的 ALIYUN_SMS_TEST_FINGERPRINT_SECRET（当前 %d 字节，至少需要 %d 字节）。请核对实际处理请求的实例配置并重启服务。", len([]byte(fingerprintSecret)), aliyunSMSFingerprintMinSize)
		c.JSON(http.StatusServiceUnavailable, app.Response{Code: http.StatusServiceUnavailable, Msg: msg})
		return
	}

	fingerprintInput, err := json.Marshal(struct {
		UserID     string                    `json:"user_id"`
		InstanceID string                    `json:"instance_id"`
		TemplateID string                    `json:"template_id"`
		WayID      string                    `json:"way_id"`
		TestPhone  string                    `json:"test_phone"`
		Config     models.InsAliyunSMSConfig `json:"config"`
		Values     map[string]interface{}    `json:"placeholders"`
		Content    string                    `json:"content"`
	}{user.GetEffectiveUserID(), instance.ID, template.ID, way.ID, req.TestPhone, config, structuredValues, content})
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "测试请求无效", nil)
		return
	}
	if h.limiter == nil {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "limiter_unavailable", maskAliyunPhone(req.TestPhone), "")
		logging.Security.Errorf("aliyun_sms_test_limiter_unavailable instance_id=%s", instance.ID)
		c.JSON(http.StatusServiceUnavailable, app.Response{Code: http.StatusServiceUnavailable, Msg: "阿里云短信测试不可用：当前服务实例未初始化限流器，请检查部署版本和服务启动日志。"})
		return
	}
	userHash := aliyunSMSFingerprint(fingerprintSecret, "scope:user", user.GetEffectiveUserID())
	wayHash := aliyunSMSFingerprint(fingerprintSecret, "scope:way", way.ID)
	requestHash := aliyunSMSFingerprint(fingerprintSecret, "request", string(fingerprintInput))
	limitCtx, cancel := context.WithTimeout(c.Request.Context(), aliyunSMSLimiterTimeout)
	decision, limitErr := h.limiter(limitCtx, userHash, wayHash, requestHash)
	cancel()
	if limitErr != nil {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "limiter_database_error", maskAliyunPhone(req.TestPhone), "")
		logging.Security.Errorf("aliyun_sms_test_limiter_database_error instance_id=%s way_id=%s err=%v", instance.ID, way.ID, limitErr)
		c.JSON(http.StatusServiceUnavailable, app.Response{Code: http.StatusServiceUnavailable, Msg: "阿里云短信测试限流存储不可用：数据库限流表操作失败。请检查数据库连通性、迁移权限，以及 rate_limit_scope_lock 和 rate_limit_event 表是否已完成迁移；详细错误已写入服务端 Security 日志。"})
		return
	}
	if !decision.Allowed {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, decision.Reason, maskAliyunPhone(req.TestPhone), "")
		c.Header("Retry-After", fmt.Sprintf("%d", decision.RetryAfterSeconds))
		c.JSON(http.StatusTooManyRequests, app.Response{Code: http.StatusTooManyRequests, Msg: "测试发送过于频繁或重复，请稍后重试"})
		return
	}

	result, err := h.channelSender().SendTemplate(auth, config, content, structuredValues, nil)
	recordAliyunSMSTestLog(c, user, ownerID, instance, template, way, req.TestPhone, result, err, auth, config, content, structuredValues)
	if err != nil {
		var providerErr aliyunProviderResponseError
		if errors.As(err, &providerErr) {
			providerCode := providerErr.ErrorCode()
			providerMessage := redactAliyunTestError(providerErr.Message(), auth, config, content, structuredValues)
			requestID := providerErr.RequestId()
			if providerMessage == "" {
				providerMessage = "阿里云已返回错误响应"
			}
			logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, providerCode, maskAliyunPhone(req.TestPhone), requestID)
			logging.Security.Errorf("aliyun_sms_test_provider_error instance_id=%s way_id=%s code=%s request_id=%s message=%s", instance.ID, way.ID, providerCode, requestID, providerMessage)
			appG.CResponse(http.StatusOK, "阿里云已返回业务响应", testAliyunSMSResponse{
				Provider: "aliyun", Code: providerCode, Message: providerMessage,
				Troubleshooting: aliyunProviderTroubleshooting(providerCode), RequestParams: &result.RequestParams,
				RequestID: requestID, Accepted: false, DeliveryStatus: "not_accepted",
				DeliveryNotice: "阿里云已返回错误响应，接口未受理；该结果不代表短信送达状态",
			})
			return
		}
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "transport_error", maskAliyunPhone(req.TestPhone), "")
		logging.Security.Errorf("aliyun_sms_test_transport_error instance_id=%s way_id=%s cause=%s", instance.ID, way.ID, redactAliyunTestError(err.Error(), auth, config, content, structuredValues))
		appG.CResponse(http.StatusBadGateway, "阿里云接口调用失败", testAliyunSMSResponse{
			Provider: "aliyun", Code: "TRANSPORT_ERROR", Message: "服务端未收到阿里云的有效响应",
			Troubleshooting: "请检查服务端 Security 日志中的 aliyun_sms_test_transport_error；重点核对服务进程的出网、DNS、HTTPS/TLS、代理/防火墙及 RegionId。该错误发生在收到阿里云业务响应之前，因此没有阿里云 RequestId。",
			RequestParams:   &result.RequestParams,
			Accepted:        false,
			DeliveryStatus:  "not_accepted", DeliveryNotice: "接口未受理；该结果不代表短信送达状态",
		})
		return
	}
	accepted := result.Code == "OK"
	status := "not_accepted"
	message := result.Message
	if accepted {
		status = "accepted_by_provider"
		if message == "" {
			message = "阿里云接口已受理；受理不代表短信最终送达"
		}
	}
	logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, result.Code, maskAliyunPhone(req.TestPhone), result.RequestID)
	appG.CResponse(http.StatusOK, "同步测试调用完成", testAliyunSMSResponse{
		Provider: "aliyun", Code: result.Code, Message: message, RequestParams: &result.RequestParams, RequestID: result.RequestID,
		BizID: result.BizID, Accepted: accepted, DeliveryStatus: status,
		DeliveryNotice: "阿里云接口受理结果不代表短信最终送达",
	})
}

func recordAliyunSMSTestLog(c *gin.Context, user app.UserContextInfo, ownerID string, instance models.SendTasksIns, template *models.TemplateResult, way models.SendWays, testPhone string, result channels.AliyunSMSSendResult, sendErr error, auth *send_way_service.WayDetailAliyunSMS, config models.InsAliyunSMSConfig, content string, values map[string]interface{}) {
	createdBy := ownerID
	if createdBy == "" {
		createdBy = user.GetEffectiveUserID()
	}

	code := result.Code
	message := result.Message
	requestID := result.RequestID
	bizID := result.BizID
	accepted := sendErr == nil && result.Code == "OK"
	if sendErr != nil {
		var providerErr aliyunProviderResponseError
		if errors.As(sendErr, &providerErr) {
			code = providerErr.ErrorCode()
			message = providerErr.Message()
			requestID = providerErr.RequestId()
		} else {
			code = "TRANSPORT_ERROR"
			message = sendErr.Error()
		}
	}
	message = redactAliyunTestError(message, auth, config, content, values)

	status := 0
	resultText := "未受理"
	if accepted {
		status = 1
		resultText = "阿里云已受理"
	}
	logText := fmt.Sprintf("阿里云短信发送测试\n结果：%s\nCode：%s\nMessage：%s\n测试手机号：%s\nTemplateCode：%s\nRegionId：%s\nRequestId：%s\nBizId：%s",
		resultText, code, message, maskAliyunPhone(testPhone), config.TemplateCode, result.RequestParams.RegionID, requestID, bizID)
	name := "阿里云短信测试"
	if template != nil && template.Name != "" {
		name += " - " + template.Name
	}
	logEntry := models.SendTasksLogs{
		Type: models.SendTasksLogTypeAliyunSMSTest, Name: name, Log: logText, Status: &status,
		CallerIp: c.ClientIP(), CreatedBy: createdBy, ModifiedBy: user.GetEffectiveUserID(),
	}
	if err := logEntry.Add(); err != nil {
		logging.Security.Errorf("aliyun_sms_test_log_write_failed instance_id=%s way_id=%s err=%v", instance.ID, way.ID, err)
	}
}

func (h *AliyunSMSInstanceAPI) resolveTestResources(ownerID string, req testAliyunSMSRequest) (models.SendTasksIns, *models.TemplateResult, models.SendWays, json.RawMessage, error) {
	instance := models.SendTasksIns{}
	configJSON := req.Config
	usingSavedConfig := len(req.Config) == 0

	if req.InstanceID != "" {
		loaded, err := models.GetTemplateInstanceByIDForOwner(req.InstanceID, ownerID)
		if err != nil {
			return instance, nil, models.SendWays{}, nil, err
		}
		instance = loaded
		if usingSavedConfig {
			if req.TemplateID != "" || req.WayID != "" {
				return instance, nil, models.SendWays{}, nil, fmt.Errorf("saved test cannot override associated resources")
			}
			configJSON = json.RawMessage(instance.Config)
		} else if req.TemplateID == "" || req.WayID == "" || req.TemplateID != instance.TemplateID || req.WayID != instance.WayID {
			return instance, nil, models.SendWays{}, nil, fmt.Errorf("draft resources do not match the saved instance")
		}
	} else if usingSavedConfig || req.TemplateID == "" || req.WayID == "" {
		return instance, nil, models.SendWays{}, nil, fmt.Errorf("new draft requires template_id, way_id and config")
	}

	templateID := req.TemplateID
	wayID := req.WayID
	if instance.ID != "" {
		templateID = instance.TemplateID
		wayID = instance.WayID
		if instance.WayType != channels.MessageTypeAliyunSMS {
			return instance, nil, models.SendWays{}, nil, fmt.Errorf("instance is not Aliyun SMS")
		}
	}
	template, way, err := loadAliyunSMSResources(ownerID, templateID, wayID)
	if err != nil {
		return instance, nil, models.SendWays{}, nil, err
	}
	if way.Type != channels.MessageTypeAliyunSMS {
		return instance, nil, models.SendWays{}, nil, fmt.Errorf("way is not Aliyun SMS")
	}
	if instance.ID == "" {
		instance = models.SendTasksIns{TemplateID: templateID, WayID: wayID, WayType: channels.MessageTypeAliyunSMS}
	}
	return instance, template, way, configJSON, nil
}

func loadAliyunSMSResources(ownerID, templateID, wayID string) (*models.TemplateResult, models.SendWays, error) {
	template, err := models.GetTemplateByIDForOwner(templateID, ownerID)
	if err != nil {
		return nil, models.SendWays{}, err
	}
	way, err := models.GetWayByIDForOwner(wayID, ownerID)
	if err != nil {
		return nil, models.SendWays{}, err
	}
	if way.Type != channels.MessageTypeAliyunSMS {
		return nil, models.SendWays{}, fmt.Errorf("way is not Aliyun SMS")
	}
	return template, way, nil
}

func getAliyunSMSRequestScope(c *gin.Context) (app.UserContextInfo, string, bool) {
	user, ok := app.GetUserFromContext(c.Request.Context())
	if !ok || user.UserID == "" {
		return app.UserContextInfo{}, "", false
	}
	ownerID := user.GetEffectiveUserID()
	if user.IsAdmin() && !user.IsImpersonating() {
		ownerID = "" // the existing model permission system grants global admins access to all owners
	}
	return user, ownerID, true
}

func mergeSupportedAliyunConfig(existingJSON, submittedJSON []byte) (models.InsAliyunSMSConfig, string, error) {
	var existing map[string]json.RawMessage
	if len(existingJSON) == 0 {
		existing = make(map[string]json.RawMessage)
	} else if err := json.Unmarshal(existingJSON, &existing); err != nil || existing == nil {
		return models.InsAliyunSMSConfig{}, "", fmt.Errorf("已保存的Config不是有效JSON对象")
	}
	var submitted map[string]json.RawMessage
	if err := json.Unmarshal(submittedJSON, &submitted); err != nil || submitted == nil {
		return models.InsAliyunSMSConfig{}, "", fmt.Errorf("Config必须是JSON对象")
	}
	allowed := map[string]bool{
		"phone_number": true, "template_code": true, "template_param_mode": true,
		"selected_params": true, "param_name_overrides": true, "allowMultiRecip": true,
	}
	if len(submitted) == 0 {
		return models.InsAliyunSMSConfig{}, "", fmt.Errorf("Config中没有可更新字段")
	}
	for key, value := range submitted {
		if !allowed[key] {
			return models.InsAliyunSMSConfig{}, "", fmt.Errorf("Config字段 %q 不允许通过此接口更新", key)
		}
		if credentialKey(key) {
			return models.InsAliyunSMSConfig{}, "", fmt.Errorf("Config不得包含渠道凭据")
		}
		if key == "allowMultiRecip" {
			var allow bool
			if err := json.Unmarshal(value, &allow); err != nil {
				return models.InsAliyunSMSConfig{}, "", fmt.Errorf("allowMultiRecip 必须是布尔值")
			}
		}
		existing[key] = value
	}
	mergedBytes, err := json.Marshal(existing)
	if err != nil {
		return models.InsAliyunSMSConfig{}, "", fmt.Errorf("Config序列化失败")
	}
	config, err := decodeAliyunSMSConfig(mergedBytes)
	if err != nil {
		return models.InsAliyunSMSConfig{}, "", err
	}
	return config, string(mergedBytes), nil
}

func decodeAliyunSMSConfig(configJSON []byte) (models.InsAliyunSMSConfig, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(configJSON, &object); err != nil || object == nil {
		return models.InsAliyunSMSConfig{}, fmt.Errorf("阿里云实例Config必须是JSON对象")
	}
	if err := validateSupportedAliyunConfigTypes(object); err != nil {
		return models.InsAliyunSMSConfig{}, err
	}
	var config models.InsAliyunSMSConfig
	if err := json.Unmarshal(configJSON, &config); err != nil {
		return models.InsAliyunSMSConfig{}, fmt.Errorf("阿里云实例Config格式无效")
	}
	if strings.TrimSpace(config.TemplateCode) == "" {
		return models.InsAliyunSMSConfig{}, fmt.Errorf("短信模板CODE不能为空")
	}
	return config, nil
}

func configWithTestPhone(configJSON []byte, phone string) ([]byte, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(configJSON, &object); err != nil || object == nil {
		return nil, fmt.Errorf("Config必须是JSON对象")
	}
	phoneJSON, err := json.Marshal(phone)
	if err != nil {
		return nil, err
	}
	object["phone_number"] = phoneJSON
	return json.Marshal(object)
}

func validateSupportedAliyunConfigTypes(object map[string]json.RawMessage) error {
	stringFields := []string{"phone_number", "template_code", "template_param_mode"}
	for _, key := range stringFields {
		if raw, exists := object[key]; exists {
			var value string
			if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
				return fmt.Errorf("Config字段 %q 必须是字符串", key)
			}
		}
	}
	if raw, exists := object["selected_params"]; exists {
		var values []string
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &values) != nil || values == nil {
			return fmt.Errorf("Config字段 selected_params 必须是字符串数组")
		}
	}
	if raw, exists := object["param_name_overrides"]; exists {
		var values map[string]string
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &values) != nil || values == nil {
			return fmt.Errorf("Config字段 param_name_overrides 必须是对象")
		}
	}
	if raw, exists := object["allowMultiRecip"]; exists {
		var value bool
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &value) != nil {
			return fmt.Errorf("Config字段 allowMultiRecip 必须是布尔值")
		}
	}
	return nil
}

func validateAliyunConfigForTemplate(config models.InsAliyunSMSConfig, configJSON []byte, placeholderJSON string) error {
	validator := &send_ins_service.InsAliyunSMSValidator{}
	message, _ := validator.Validate(string(configJSON))
	if message != "" {
		return fmt.Errorf("阿里云短信Config无效: %s", message)
	}
	if config.TemplateParamMode != channels.AliyunSMSStructuredParamMode {
		return nil
	}
	definitions, err := parseAliyunPlaceholderDefinitions(placeholderJSON)
	if err != nil {
		return fmt.Errorf("模板Placeholders无效: %w", err)
	}
	declared := make(map[string]struct{}, len(definitions))
	for _, definition := range definitions {
		if definition.Key != "" {
			declared[definition.Key] = struct{}{}
		}
	}
	for _, name := range config.SelectedParams {
		if _, exists := declared[name]; !exists {
			return fmt.Errorf("所选变量 %q 已从模板删除或改名，请重新选择模板变量", name)
		}
	}
	return nil
}

type aliyunPlaceholderDefinition struct {
	Key     string `json:"key"`
	Default string `json:"default"`
}

func parseAliyunPlaceholderDefinitions(raw string) ([]aliyunPlaceholderDefinition, error) {
	if raw == "" {
		return []aliyunPlaceholderDefinition{}, nil
	}
	var definitions []aliyunPlaceholderDefinition
	if err := json.Unmarshal([]byte(raw), &definitions); err != nil || definitions == nil {
		return nil, fmt.Errorf("必须是变量定义数组")
	}
	return definitions, nil
}

func resolveAliyunTemplateValues(raw string, provided map[string]interface{}, strictDefinitions bool) ([]aliyunPlaceholderDefinition, map[string]interface{}, error) {
	definitions, err := parseAliyunPlaceholderDefinitions(raw)
	if err != nil {
		if strictDefinitions {
			return nil, nil, err
		}
		definitions = nil // match V2 behavior for legacy content mode: invalid definitions do not inject defaults
	}
	values := make(map[string]interface{}, len(provided)+len(definitions))
	for key, value := range provided {
		values[key] = value
	}
	for _, definition := range definitions {
		if _, exists := values[definition.Key]; !exists && definition.Default != "" {
			values[definition.Key] = definition.Default
		}
	}
	return definitions, values, nil
}

func selectDeclaredAliyunValues(definitions []aliyunPlaceholderDefinition, values map[string]interface{}) map[string]interface{} {
	selected := make(map[string]interface{}, len(definitions))
	for _, definition := range definitions {
		if value, exists := values[definition.Key]; exists && definition.Key != "" {
			selected[definition.Key] = value
		}
	}
	return selected
}

func renderAliyunSMSContent(template string, values map[string]interface{}) string {
	result := template
	for key, value := range values {
		result = strings.ReplaceAll(result, "{{"+key+"}}", fmt.Sprintf("%v", value))
	}
	return result
}

func rejectCredentialConfig(configJSON []byte) error {
	var object interface{}
	if err := json.Unmarshal(configJSON, &object); err != nil {
		return fmt.Errorf("Config格式无效")
	}
	if valueContainsCredentialKey(object) {
		return fmt.Errorf("Config不得包含渠道凭据")
	}
	return nil
}

func containsCredentialField(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	var object interface{}
	if json.Unmarshal(raw, &object) != nil {
		return true
	}
	return valueContainsCredentialKey(object)
}

func valueContainsCredentialKey(value interface{}) bool {
	switch typed := value.(type) {
	case map[string]interface{}:
		for key, child := range typed {
			if credentialKey(key) || valueContainsCredentialKey(child) {
				return true
			}
		}
	case []interface{}:
		for _, child := range typed {
			if valueContainsCredentialKey(child) {
				return true
			}
		}
	}
	return false
}

func credentialKey(key string) bool {
	key = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(key, "_", ""), "-", ""))
	return strings.Contains(key, "accesskey") || strings.Contains(key, "secret") || strings.Contains(key, "credential") || strings.Contains(key, "password")
}

func decodeStrictAliyunRequest(c *gin.Context, target interface{}) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, aliyunSMSRequestLimit)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing interface{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("请求只能包含一个JSON对象")
	}
	return nil
}

func validateAliyunWay(way models.SendWays) (*send_way_service.WayDetailAliyunSMS, error) {
	service := send_way_service.SendWay{Type: way.Type, Auth: way.Auth}
	message, value := service.ValidateDiffWay()
	if message != "" {
		return nil, fmt.Errorf("invalid Aliyun way config")
	}
	auth, ok := value.(*send_way_service.WayDetailAliyunSMS)
	if !ok {
		return nil, fmt.Errorf("invalid Aliyun way config")
	}
	return auth, nil
}

func logAliyunSMSAudit(action string, user app.UserContextInfo, ownerID, instanceID, templateID, wayID, outcome, maskedPhone, requestID string) {
	logging.Security.Infof("aliyun_sms_operation action=%s actor_id=%s effective_owner=%s instance_id=%s template_id=%s way_id=%s outcome=%s phone=%s request_id=%s",
		action, user.UserID, ownerID, instanceID, templateID, wayID, outcome, maskedPhone, requestID)
}

func maskAliyunPhone(phone string) string {
	digits := strings.TrimPrefix(phone, "+")
	if len(digits) <= 4 {
		return strings.Repeat("*", len(digits))
	}
	return "***" + digits[len(digits)-4:]
}

func redactAliyunTestError(detail string, auth *send_way_service.WayDetailAliyunSMS, config models.InsAliyunSMSConfig, content string, values map[string]interface{}) string {
	redact := func(value string) {
		if len([]byte(value)) >= 4 {
			detail = strings.ReplaceAll(detail, value, "[REDACTED]")
		}
	}
	if auth != nil {
		redact(auth.AccessKeySecret)
		redact(auth.AccessKeyId)
	}
	redact(config.PhoneNumber)
	redact(content)
	for _, value := range values {
		if text, ok := value.(string); ok {
			redact(text)
		}
	}
	return detail
}

func aliyunProviderTroubleshooting(code string) string {
	switch code {
	case "NoPermission", "isp.RAM_PERMISSION_DENY", "Forbidden.RAM":
		return "阿里云拒绝了 dysms:SendSms 权限：请在 RAM 中为 AccessKey 所属子用户或角色授予短信发送权限，并检查显式拒绝、权限边界和账号级授权策略。可用 RequestId 打开阿里云诊断。"
	default:
		return "阿里云已返回业务错误。请按 Code 和 Message 检查短信服务配置，并使用 RequestId 在阿里云诊断页面查询；该错误不是服务端网络连接失败。"
	}
}

func aliyunSMSFingerprint(secret, domain, value string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("message-nest:aliyun-sms-test:v1:" + domain + ":"))
	_, _ = mac.Write([]byte(value))
	return hex.EncodeToString(mac.Sum(nil))
}

func editConfigResponseID(instanceID string) gin.H {
	return gin.H{"id": instanceID}
}
