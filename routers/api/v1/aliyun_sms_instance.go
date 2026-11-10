package v1

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/logging"
	"message-nest/service/send_ins_service"
	"message-nest/service/send_message_service/unified/channels"
	"message-nest/service/send_way_service"
)

const (
	aliyunSMSUserTestLimit = 10
	aliyunSMSWayTestLimit  = 30
	aliyunSMSRateWindow    = time.Minute
	aliyunSMSRepeatWindow  = 20 * time.Second
	aliyunSMSRequestLimit  = 64 << 10
)

var testPhonePattern = regexp.MustCompile(`^\+?[0-9]{5,20}$`)

type aliyunSMSChannelSender interface {
	SendTemplate(*send_way_service.WayDetailAliyunSMS, models.InsAliyunSMSConfig, string, map[string]interface{}, map[string]interface{}) (channels.AliyunSMSSendResult, error)
}

type AliyunSMSInstanceAPI struct {
	guard         *aliyunSMSTestGuard
	channelSender func() aliyunSMSChannelSender
}

func NewAliyunSMSInstanceAPI() *AliyunSMSInstanceAPI {
	return &AliyunSMSInstanceAPI{
		guard: newAliyunSMSTestGuard(aliyunSMSUserTestLimit, aliyunSMSWayTestLimit, aliyunSMSRateWindow, aliyunSMSRepeatWindow),
		channelSender: func() aliyunSMSChannelSender {
			return channels.NewAliyunSMSChannel()
		},
	}
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
	Provider       string `json:"provider"`
	Code           string `json:"code"`
	Message        string `json:"message"`
	RequestID      string `json:"request_id,omitempty"`
	BizID          string `json:"biz_id,omitempty"`
	Accepted       bool   `json:"accepted"`
	DeliveryStatus string `json:"delivery_status"`
	DeliveryNotice string `json:"delivery_notice"`
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

	fingerprintInput, err := json.Marshal(struct {
		InstanceID string                    `json:"instance_id"`
		TemplateID string                    `json:"template_id"`
		WayID      string                    `json:"way_id"`
		TestPhone  string                    `json:"test_phone"`
		Config     models.InsAliyunSMSConfig `json:"config"`
		Values     map[string]interface{}    `json:"placeholders"`
		Content    string                    `json:"content"`
	}{instance.ID, template.ID, way.ID, req.TestPhone, config, structuredValues, content})
	if err != nil {
		appG.CResponse(http.StatusBadRequest, "测试请求无效", nil)
		return
	}
	if retryAfter, reason := h.guard.Allow(user.GetEffectiveUserID(), way.ID, fingerprintInput); reason != "" {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, reason, maskAliyunPhone(req.TestPhone), "")
		c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
		c.JSON(http.StatusTooManyRequests, app.Response{Code: http.StatusTooManyRequests, Msg: "测试发送过于频繁或重复，请稍后重试"})
		return
	}

	auth, err := validateAliyunWay(way)
	if err != nil {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "invalid_way_auth", maskAliyunPhone(req.TestPhone), "")
		appG.CResponse(http.StatusBadRequest, "阿里云渠道配置无效", nil)
		return
	}
	result, err := h.channelSender().SendTemplate(auth, config, content, structuredValues, nil)
	if err != nil {
		logAliyunSMSAudit("test", user, ownerID, instance.ID, template.ID, way.ID, "transport_error", maskAliyunPhone(req.TestPhone), "")
		appG.CResponse(http.StatusBadGateway, "阿里云接口调用失败", testAliyunSMSResponse{
			Provider: "aliyun", Code: "TRANSPORT_ERROR", Message: "阿里云接口调用失败", Accepted: false,
			DeliveryStatus: "not_accepted", DeliveryNotice: "接口未受理；该结果不代表短信送达状态",
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
		Provider: "aliyun", Code: result.Code, Message: message, RequestID: result.RequestID,
		BizID: result.BizID, Accepted: accepted, DeliveryStatus: status,
		DeliveryNotice: "阿里云接口受理结果不代表短信最终送达",
	})
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

type aliyunRateWindow struct {
	count   int
	resetAt time.Time
}

type aliyunSMSRepeatEntry struct {
	expires time.Time
}

type aliyunSMSTestGuard struct {
	mu             sync.Mutex
	userLimit      int
	wayLimit       int
	window         time.Duration
	repeatWindow   time.Duration
	key            [32]byte
	userWindows    map[string]aliyunRateWindow
	wayWindows     map[string]aliyunRateWindow
	recentRequests map[string]aliyunSMSRepeatEntry
}

func newAliyunSMSTestGuard(userLimit, wayLimit int, window, repeatWindow time.Duration) *aliyunSMSTestGuard {
	guard := &aliyunSMSTestGuard{
		userLimit: userLimit, wayLimit: wayLimit, window: window, repeatWindow: repeatWindow,
		userWindows: make(map[string]aliyunRateWindow), wayWindows: make(map[string]aliyunRateWindow),
		recentRequests: make(map[string]aliyunSMSRepeatEntry),
	}
	if _, err := rand.Read(guard.key[:]); err != nil {
		panic("cannot initialize Aliyun SMS repeat guard")
	}
	return guard
}

func (g *aliyunSMSTestGuard) Allow(userID, wayID string, payload []byte) (int, string) {
	now := time.Now()
	mac := hmac.New(sha256.New, g.key[:])
	_, _ = mac.Write(payload)
	repeatKey := userID + ":" + wayID + ":" + hex.EncodeToString(mac.Sum(nil))

	g.mu.Lock()
	defer g.mu.Unlock()
	g.cleanup(now)
	if entry, exists := g.recentRequests[repeatKey]; exists && now.Before(entry.expires) {
		return int(g.repeatWindow.Seconds()), "duplicate"
	}
	if len(g.recentRequests) >= 4096 {
		return int(g.window.Seconds()), "guard_capacity"
	}
	if retry := g.windowRetry(g.userWindows, userID, g.userLimit, now); retry > 0 {
		return retry, "user_rate_limited"
	}
	if retry := g.windowRetry(g.wayWindows, wayID, g.wayLimit, now); retry > 0 {
		return retry, "way_rate_limited"
	}
	g.incrementWindow(g.userWindows, userID, now)
	g.incrementWindow(g.wayWindows, wayID, now)
	g.recentRequests[repeatKey] = aliyunSMSRepeatEntry{expires: now.Add(g.repeatWindow)}
	return 0, ""
}

func (g *aliyunSMSTestGuard) windowRetry(windows map[string]aliyunRateWindow, key string, limit int, now time.Time) int {
	window, exists := windows[key]
	if !exists || !now.Before(window.resetAt) {
		if len(windows) >= 4096 {
			return int(g.window.Seconds())
		}
		return 0
	}
	if window.count >= limit {
		seconds := int(window.resetAt.Sub(now).Seconds())
		if seconds < 1 {
			seconds = 1
		}
		return seconds
	}
	return 0
}

func (g *aliyunSMSTestGuard) incrementWindow(windows map[string]aliyunRateWindow, key string, now time.Time) {
	window, exists := windows[key]
	if !exists || !now.Before(window.resetAt) {
		windows[key] = aliyunRateWindow{count: 1, resetAt: now.Add(g.window)}
		return
	}
	window.count++
	windows[key] = window
}

func (g *aliyunSMSTestGuard) cleanup(now time.Time) {
	for key, window := range g.userWindows {
		if !now.Before(window.resetAt) {
			delete(g.userWindows, key)
		}
	}
	for key, window := range g.wayWindows {
		if !now.Before(window.resetAt) {
			delete(g.wayWindows, key)
		}
	}
	for key, entry := range g.recentRequests {
		if !now.Before(entry.expires) {
			delete(g.recentRequests, key)
		}
	}
}

func editConfigResponseID(instanceID string) gin.H {
	return gin.H{"id": instanceID}
}
