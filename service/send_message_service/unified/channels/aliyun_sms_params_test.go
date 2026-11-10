package channels

import (
	"encoding/json"
	"errors"
	"testing"

	"message-nest/models"
	"message-nest/service/send_way_service"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
)

type mockAliyunSMSClient struct {
	request *dysmsapi.SendSmsRequest
	err     error
}

func (m *mockAliyunSMSClient) SendSms(request *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error) {
	m.request = request
	if m.err != nil {
		return nil, m.err
	}
	return &dysmsapi.SendSmsResponse{Code: "OK", RequestId: "mock-request", BizId: "mock-biz"}, nil
}

func TestAliyunSMSChannelStructuredParamsUseSelectedMappings(t *testing.T) {
	client := &mockAliyunSMSClient{}
	channel := NewAliyunSMSChannel()
	channel.clientFactory = func(accessKeyID, accessKeySecret, regionID string) (aliyunSMSClient, error) {
		if accessKeyID != "test-key" || accessKeySecret != "test-secret" || regionID != "cn-test" {
			t.Fatalf("unexpected credentials/region passed to factory: %q %q %q", accessKeyID, accessKeySecret, regionID)
		}
		return client, nil
	}

	ins := models.SendTasksIns{
		WayType: "AliyunSMS",
		Config:  `{"phone_number":"15500000000","template_code":"SMS_TEST","template_param_mode":"structured","selected_params":["name","count"],"param_name_overrides":{"name":"customer_name"}}`,
	}
	content := &UnifiedMessageContent{
		Text: "Hello Ada",
		TemplateParams: map[string]interface{}{
			"name":  "Ada",
			"count": 3,
			"extra": "must not be sent",
		},
		Extra: map[string]interface{}{"unrelated": "must not be sent"},
	}

	result, errText := channel.SendUnified(&send_way_service.WayDetailAliyunSMS{
		AccessKeyId: "test-key", AccessKeySecret: "test-secret", SignName: "Test", RegionId: "cn-test",
	}, ins, content)
	if errText != "" {
		t.Fatalf("SendUnified returned error: %s", errText)
	}
	if result == "" || client.request == nil {
		t.Fatalf("expected mock request and result, got result=%q request=%v", result, client.request)
	}
	if client.request.PhoneNumbers != "15500000000" || client.request.TemplateCode != "SMS_TEST" || client.request.SignName != "Test" {
		t.Fatalf("unexpected request target fields: %+v", client.request)
	}
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(client.request.TemplateParam), &params); err != nil {
		t.Fatalf("invalid TemplateParam JSON: %v", err)
	}
	want := map[string]interface{}{"customer_name": "Ada", "count": "3"}
	if !equalJSONMap(params, want) {
		t.Fatalf("TemplateParam = %#v, want %#v", params, want)
	}
}

func TestAliyunSMSChannelLegacyModePreservesContentAndExtra(t *testing.T) {
	client := &mockAliyunSMSClient{}
	channel := NewAliyunSMSChannel()
	channel.clientFactory = func(string, string, string) (aliyunSMSClient, error) { return client, nil }
	ins := models.SendTasksIns{
		WayType: "AliyunSMS",
		Config:  `{"phone_number":"15500000000","template_code":"SMS_LEGACY"}`,
	}
	content := &UnifiedMessageContent{Text: "legacy body", Extra: map[string]interface{}{"content": "legacy override", "code": "123"}}

	_, errText := channel.SendUnified(&send_way_service.WayDetailAliyunSMS{SignName: "Test"}, ins, content)
	if errText != "" {
		t.Fatalf("SendUnified returned error: %s", errText)
	}
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(client.request.TemplateParam), &params); err != nil {
		t.Fatalf("invalid TemplateParam JSON: %v", err)
	}
	want := map[string]interface{}{"content": "legacy override", "code": "123"}
	if !equalJSONMap(params, want) {
		t.Fatalf("legacy TemplateParam = %#v, want %#v", params, want)
	}
}

func TestBuildAliyunTemplateParamsRejectsInvalidStructuredInputs(t *testing.T) {
	tests := []struct {
		name   string
		config models.InsAliyunSMSConfig
		values map[string]interface{}
	}{
		{name: "missing explicit selection", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured"}},
		{name: "duplicate selection", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured", SelectedParams: []string{"code", "code"}}, values: map[string]interface{}{"code": "123"}},
		{name: "missing runtime value", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured", SelectedParams: []string{"code"}}},
		{name: "unknown override source", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured", SelectedParams: []string{"code"}, ParamNameOverrides: map[string]string{"other": "target"}}, values: map[string]interface{}{"code": "123"}},
		{name: "empty override", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured", SelectedParams: []string{"code"}, ParamNameOverrides: map[string]string{"code": " "}}, values: map[string]interface{}{"code": "123"}},
		{name: "redundant same-name override", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured", SelectedParams: []string{"code"}, ParamNameOverrides: map[string]string{"code": "code"}}, values: map[string]interface{}{"code": "123"}},
		{name: "duplicate target", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured", SelectedParams: []string{"a", "b"}, ParamNameOverrides: map[string]string{"b": "a"}}, values: map[string]interface{}{"a": "1", "b": "2"}},
		{name: "non scalar value", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured", SelectedParams: []string{"code"}}, values: map[string]interface{}{"code": map[string]interface{}{"nested": true}}},
		{name: "null value", config: models.InsAliyunSMSConfig{TemplateParamMode: "structured", SelectedParams: []string{"code"}}, values: map[string]interface{}{"code": nil}},
		{name: "unsupported mode", config: models.InsAliyunSMSConfig{TemplateParamMode: "future", SelectedParams: []string{}}, values: map[string]interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := BuildAliyunTemplateParams(tt.config, "ignored", tt.values, nil); err == nil {
				t.Fatal("expected invalid input to fail")
			}
		})
	}
}

func TestAliyunSMSChannelDoesNotCallProviderForInvalidConfig(t *testing.T) {
	called := false
	channel := NewAliyunSMSChannel()
	channel.clientFactory = func(string, string, string) (aliyunSMSClient, error) {
		called = true
		return nil, errors.New("should not be called")
	}
	ins := models.SendTasksIns{
		WayType: "AliyunSMS",
		Config:  `{"phone_number":"15500000000","template_code":"SMS_TEST","template_param_mode":"structured"}`,
	}
	_, errText := channel.SendUnified(&send_way_service.WayDetailAliyunSMS{SignName: "Test"}, ins, &UnifiedMessageContent{Text: "body"})
	if errText == "" {
		t.Fatal("expected invalid structured configuration to fail")
	}
	if called {
		t.Fatal("provider client factory was called for invalid configuration")
	}
}

func equalJSONMap(actual, expected map[string]interface{}) bool {
	if len(actual) != len(expected) {
		return false
	}
	for key, expectedValue := range expected {
		if actual[key] != expectedValue {
			return false
		}
	}
	return true
}
