package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"message-nest/middleware"
	"message-nest/models"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"message-nest/service/send_message_service/unified/channels"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type apiMockAliyunClient struct {
	request  *dysmsapi.SendSmsRequest
	response *dysmsapi.SendSmsResponse
	err      error
	calls    int
}

func (m *apiMockAliyunClient) SendSms(request *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error) {
	m.calls++
	m.request = request
	return m.response, m.err
}

type aliyunAPITestFixture struct {
	router   *gin.Engine
	handler  *AliyunSMSInstanceAPI
	client   *apiMockAliyunClient
	ownerID  string
	otherID  string
	template models.Template
	way      models.SendWays
	instance models.SendTasksIns
	token    string
}

func newAliyunAPITestFixture(t *testing.T) *aliyunAPITestFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	models.InitTestDB("aliyun_instance_api_" + util.GenerateUniqueID())
	setting.AppSetting.JwtSecret = "AliyunAPITestSecretKey123"
	setting.AppSetting.AliyunSMSTestFingerprintSecret = "test-aliyun-sms-fingerprint-secret-32-bytes"

	ownerID := "UR_OWNER_A"
	otherID := "UR_OWNER_B"
	token, err := util.GenerateToken(ownerID, "owner-a", "hash", "user", 1)
	if err != nil {
		t.Fatal(err)
	}
	template := models.Template{
		UUIDModel: models.UUIDModel{ID: "TP_TEST_A", CreatedBy: ownerID, ModifiedBy: ownerID},
		Name:      "SMS template", Status: "enabled",
		TextTemplate: "Hello {{name}} {{code}}",
		Placeholders: `[{"key":"name","default":"DefaultName"},{"key":"code"}]`,
	}
	way := models.SendWays{
		UUIDModel: models.UUIDModel{ID: "WY_TEST_A", CreatedBy: ownerID, ModifiedBy: ownerID},
		Name:      "Aliyun test way", Type: channels.MessageTypeAliyunSMS,
		Auth: `{"access_key_id":"test-access-id","access_key_secret":"DO_NOT_LEAK_TEST_SECRET","sign_name":"TestSign","region_id":"cn-test"}`,
	}
	instance := models.SendTasksIns{
		UUIDModel:  models.UUIDModel{ID: "IN_TEST_A", CreatedBy: ownerID, ModifiedBy: ownerID},
		TemplateID: template.ID, WayID: way.ID, WayType: channels.MessageTypeAliyunSMS,
		ContentType: "text", Config: `{"phone_number":"15500000000","template_code":"SMS_SAVED","future_config":{"keep":true}}`,
	}
	for _, value := range []interface{}{&template, &way, &instance} {
		if err := models.GetDB().Create(value).Error; err != nil {
			t.Fatalf("seed test database: %v", err)
		}
	}

	client := &apiMockAliyunClient{response: &dysmsapi.SendSmsResponse{Code: "OK", Message: "accepted", RequestId: "mock-request-id", BizId: "mock-biz-id"}}
	handler := NewAliyunSMSInstanceAPI()
	handler.channelSender = func() AliyunSMSChannelSender {
		return channels.NewAliyunSMSChannelWithClientFactory(func(accessKeyID, accessKeySecret, regionID string) (channels.AliyunSMSClient, error) {
			if accessKeyID != "test-access-id" || accessKeySecret != "DO_NOT_LEAK_TEST_SECRET" || regionID != "cn-test" {
				t.Fatalf("unexpected stored credentials passed to mock: %q %q %q", accessKeyID, accessKeySecret, regionID)
			}
			return client, nil
		})
	}
	router := gin.New()
	api := router.Group("/api/v1")
	api.Use(middleware.JWT())
	api.POST("/templates/ins/aliyun/edit", handler.EditConfig)
	api.POST("/templates/ins/aliyun/test", handler.TestSend)

	return &aliyunAPITestFixture{
		router: router, handler: handler, client: client, ownerID: ownerID, otherID: otherID,
		template: template, way: way, instance: instance, token: token,
	}
}

func (f *aliyunAPITestFixture) request(path, token, body string) (*httptest.ResponseRecorder, modelsResponse) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	f.router.ServeHTTP(w, req)
	var response modelsResponse
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	return w, response
}

type modelsResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func TestEditAliyunInstancePreservesUnknownConfigAndImmutableAssociations(t *testing.T) {
	f := newAliyunAPITestFixture(t)
	w, response := f.request("/api/v1/templates/ins/aliyun/edit", f.token,
		`{"id":"IN_TEST_A","config":{"template_code":"SMS_UPDATED","phone_number":"15500000001"}}`)
	if w.Code != http.StatusOK || response.Code != http.StatusOK {
		t.Fatalf("edit failed: HTTP %d response=%+v body=%s", w.Code, response, w.Body.String())
	}
	var updated models.SendTasksIns
	if err := models.GetDB().Where("id = ?", f.instance.ID).First(&updated).Error; err != nil {
		t.Fatal(err)
	}
	if updated.TemplateID != f.instance.TemplateID || updated.WayID != f.instance.WayID || updated.TaskID != f.instance.TaskID || updated.ID != f.instance.ID {
		t.Fatalf("immutable association changed: %+v", updated)
	}
	if updated.CreatedBy != f.ownerID || updated.ModifiedBy != f.ownerID {
		t.Fatalf("unexpected audit owner: created_by=%q modified_by=%q", updated.CreatedBy, updated.ModifiedBy)
	}
	var config map[string]json.RawMessage
	if err := json.Unmarshal([]byte(updated.Config), &config); err != nil {
		t.Fatal(err)
	}
	if string(config["template_code"]) != `"SMS_UPDATED"` || len(config["future_config"]) == 0 {
		t.Fatalf("supported edit or unknown field preservation failed: %s", updated.Config)
	}
	if _, exists := config["template_param_mode"]; exists {
		t.Fatalf("legacy instance was silently converted to structured: %s", updated.Config)
	}
}

func TestEditAliyunInstanceRejectsUnsupportedAndStaleConfig(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "immutable top-level fields", body: `{"id":"IN_TEST_A","template_id":"OTHER","config":{"template_code":"SMS"}}`},
		{name: "unsupported config field", body: `{"id":"IN_TEST_A","config":{"access_key_secret":"secret"}}`},
		{name: "removed selected placeholder", body: `{"id":"IN_TEST_A","config":{"template_param_mode":"structured","selected_params":["renamed"],"template_code":"SMS"}}`},
		{name: "null mode cannot downgrade silently", body: `{"id":"IN_TEST_A","config":{"template_param_mode":null}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newAliyunAPITestFixture(t)
			w, response := f.request("/api/v1/templates/ins/aliyun/edit", f.token, tt.body)
			if w.Code != http.StatusOK || response.Code == http.StatusOK {
				t.Fatalf("expected rejection, got HTTP %d response=%+v body=%s", w.Code, response, w.Body.String())
			}
			if f.client.calls != 0 {
				t.Fatal("edit must not invoke provider")
			}
		})
	}
}

func TestEditAliyunInstanceEnforcesOwnershipAndChannelType(t *testing.T) {
	f := newAliyunAPITestFixture(t)
	foreign := f.instance
	foreign.ID = "IN_FOREIGN_B"
	foreign.CreatedBy = f.otherID
	foreign.ModifiedBy = f.otherID
	if err := models.GetDB().Create(&foreign).Error; err != nil {
		t.Fatal(err)
	}
	_, response := f.request("/api/v1/templates/ins/aliyun/edit", f.token,
		`{"id":"IN_FOREIGN_B","config":{"template_code":"SMS_CHANGED"}}`)
	if response.Code == http.StatusOK {
		t.Fatal("cross-owner instance edit should be denied")
	}

	nonSMS := f.instance
	nonSMS.ID = "IN_EMAIL_A"
	nonSMS.WayType = "Email"
	if err := models.GetDB().Create(&nonSMS).Error; err != nil {
		t.Fatal(err)
	}
	_, response = f.request("/api/v1/templates/ins/aliyun/edit", f.token,
		`{"id":"IN_EMAIL_A","config":{"template_code":"SMS_CHANGED"}}`)
	if response.Code == http.StatusOK {
		t.Fatal("non-Aliyun instance edit should be denied")
	}
}

func TestAliyunTestSendNewDraftUsesMockAndDoesNotPersistOrCreateBusinessLogs(t *testing.T) {
	f := newAliyunAPITestFixture(t)
	var output bytes.Buffer
	originalOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&output)
	t.Cleanup(func() { logrus.SetOutput(originalOutput) })

	beforeConfig := f.instance.Config
	w, response := f.request("/api/v1/templates/ins/aliyun/test", f.token,
		`{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"phone_number":"15500000000","template_code":"SMS_DRAFT","template_param_mode":"structured","selected_params":["name"],"param_name_overrides":{}},"test_phone":"15555550123","placeholders":{"name":"DRAFT_VALUE","code":"UNSELECTED_VALUE"}}`)
	if w.Code != http.StatusOK || response.Code != http.StatusOK {
		t.Fatalf("draft test failed: HTTP %d response=%+v body=%s", w.Code, response, w.Body.String())
	}
	if f.client.calls != 1 || f.client.request == nil {
		t.Fatalf("expected one mock provider call, got %d", f.client.calls)
	}
	if f.client.request.PhoneNumbers != "15555550123" || f.client.request.TemplateCode != "SMS_DRAFT" {
		t.Fatalf("draft request did not use test phone/config: %+v", f.client.request)
	}
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(f.client.request.TemplateParam), &params); err != nil {
		t.Fatal(err)
	}
	if params["name"] != "DRAFT_VALUE" || len(params) != 1 {
		t.Fatalf("only selected declared params should be sent: %#v", params)
	}
	var stored models.SendTasksIns
	if err := models.GetDB().Where("id = ?", f.instance.ID).First(&stored).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Config != beforeConfig {
		t.Fatal("draft test changed persisted instance config")
	}
	var logs, stats int64
	models.GetDB().Model(&models.SendTasksLogs{}).Count(&logs)
	models.GetDB().Model(&models.SendStats{}).Count(&stats)
	if logs != 0 || stats != 0 {
		t.Fatalf("test send entered business logs/statistics: logs=%d stats=%d", logs, stats)
	}

	var payload testAliyunSMSResponse
	if err := json.Unmarshal(response.Data, &payload); err != nil {
		t.Fatal(err)
	}
	if !payload.Accepted || payload.DeliveryStatus != "accepted_by_provider" || payload.BizID != "mock-biz-id" {
		t.Fatalf("provider response contract incorrect: %+v", payload)
	}
	if strings.Contains(w.Body.String(), "DO_NOT_LEAK_TEST_SECRET") || strings.Contains(output.String(), "DO_NOT_LEAK_TEST_SECRET") || strings.Contains(output.String(), "DRAFT_VALUE") || strings.Contains(output.String(), "15555550123") {
		t.Fatalf("sensitive test data leaked: response=%s logs=%s", w.Body.String(), output.String())
	}
	if !strings.Contains(output.String(), "phone=***0123") {
		t.Fatalf("masked phone should appear in audit: %s", output.String())
	}
}

func TestAliyunTestSendEditDraftAndSavedInstance(t *testing.T) {
	t.Run("edit draft", func(t *testing.T) {
		f := newAliyunAPITestFixture(t)
		body := `{"instance_id":"IN_TEST_A","template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"phone_number":"15500000000","template_code":"SMS_EDIT_DRAFT","template_param_mode":"structured","selected_params":["name"],"param_name_overrides":{}},"test_phone":"15555550124","placeholders":{"name":"EDIT_DRAFT"}}`
		_, response := f.request("/api/v1/templates/ins/aliyun/test", f.token, body)
		if response.Code != http.StatusOK || f.client.calls != 1 || f.client.request.TemplateCode != "SMS_EDIT_DRAFT" {
			t.Fatalf("edit draft test failed: response=%+v calls=%d", response, f.client.calls)
		}
		var stored models.SendTasksIns
		_ = models.GetDB().Where("id = ?", f.instance.ID).First(&stored).Error
		if stored.Config != f.instance.Config {
			t.Fatal("edit draft test persisted its unsaved config")
		}
	})

	t.Run("saved instance", func(t *testing.T) {
		f := newAliyunAPITestFixture(t)
		f.instance.Config = `{"phone_number":"15500000000","template_code":"SMS_SAVED","template_param_mode":"structured","selected_params":["name"],"param_name_overrides":{},"future_config":{"keep":true}}`
		if err := models.GetDB().Model(&models.SendTasksIns{}).Where("id = ?", f.instance.ID).Update("config", f.instance.Config).Error; err != nil {
			t.Fatal(err)
		}
		body := `{"instance_id":"IN_TEST_A","test_phone":"15555550125"}`
		_, response := f.request("/api/v1/templates/ins/aliyun/test", f.token, body)
		if response.Code != http.StatusOK || f.client.calls != 1 || f.client.request.TemplateCode != "SMS_SAVED" {
			t.Fatalf("saved instance test failed: response=%+v calls=%d", response, f.client.calls)
		}
		var params map[string]interface{}
		_ = json.Unmarshal([]byte(f.client.request.TemplateParam), &params)
		if params["name"] != "DefaultName" || len(params) != 1 {
			t.Fatalf("unexpected saved test params: %#v", params)
		}
	})
}

func TestAliyunTestSendRejectsUnauthorizedResourcesAndClientCredentials(t *testing.T) {
	f := newAliyunAPITestFixture(t)
	foreignTemplate := f.template
	foreignTemplate.ID = "TP_FOREIGN_B"
	foreignTemplate.CreatedBy = f.otherID
	if err := models.GetDB().Create(&foreignTemplate).Error; err != nil {
		t.Fatal(err)
	}
	foreignWay := f.way
	foreignWay.ID = "WY_FOREIGN_B"
	foreignWay.CreatedBy = f.otherID
	if err := models.GetDB().Create(&foreignWay).Error; err != nil {
		t.Fatal(err)
	}
	nonSMS := f.way
	nonSMS.ID = "WY_EMAIL_A"
	nonSMS.Type = "Email"
	if err := models.GetDB().Create(&nonSMS).Error; err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		body  string
		token string
	}{
		{name: "foreign template", body: `{"template_id":"TP_FOREIGN_B","way_id":"WY_TEST_A","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"15555550120"}`, token: f.token},
		{name: "foreign way", body: `{"template_id":"TP_TEST_A","way_id":"WY_FOREIGN_B","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"15555550120"}`, token: f.token},
		{name: "non SMS way", body: `{"template_id":"TP_TEST_A","way_id":"WY_EMAIL_A","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"15555550120"}`, token: f.token},
		{name: "cross owner saved instance", body: `{"instance_id":"IN_FOREIGN_B","test_phone":"15555550120"}`, token: f.token},
		{name: "client secret top-level", body: `{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"template_code":"SMS"},"test_phone":"15555550120","access_key_secret":"client-secret"}`, token: f.token},
		{name: "client secret in config", body: `{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"template_code":"SMS","access_key_secret":"client-secret"},"test_phone":"15555550120"}`, token: f.token},
		{name: "missing JWT", body: `{"instance_id":"IN_TEST_A","test_phone":"15555550120"}`, token: ""},
	}
	foreignIns := f.instance
	foreignIns.ID = "IN_FOREIGN_B"
	foreignIns.CreatedBy = f.otherID
	if err := models.GetDB().Create(&foreignIns).Error; err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, response := f.request("/api/v1/templates/ins/aliyun/test", tt.token, tt.body)
			if response.Code == http.StatusOK {
				t.Fatal("expected request rejection")
			}
		})
	}
	if f.client.calls != 0 {
		t.Fatalf("unauthorized/invalid requests reached mock provider %d times", f.client.calls)
	}
}

func TestAliyunTestSendRateLimitRepeatAndProviderFailures(t *testing.T) {
	t.Run("fingerprint key is required", func(t *testing.T) {
		f := newAliyunAPITestFixture(t)
		setting.AppSetting.AliyunSMSTestFingerprintSecret = ""
		body := `{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"15555550110"}`
		w, response := f.request("/api/v1/templates/ins/aliyun/test", f.token, body)
		if w.Code != http.StatusServiceUnavailable || response.Code != http.StatusServiceUnavailable || f.client.calls != 0 {
			t.Fatalf("missing stable key must fail closed: http=%d response=%+v calls=%d", w.Code, response, f.client.calls)
		}
	})

	t.Run("missing rate-limit table fails closed", func(t *testing.T) {
		f := newAliyunAPITestFixture(t)
		if err := models.GetDB().Migrator().DropTable(&models.AliyunSMSTestEvent{}); err != nil {
			t.Fatal(err)
		}
		body := `{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"15555550111"}`
		w, response := f.request("/api/v1/templates/ins/aliyun/test", f.token, body)
		if w.Code != http.StatusServiceUnavailable || response.Code != http.StatusServiceUnavailable || f.client.calls != 0 {
			t.Fatalf("missing event table must fail closed: http=%d response=%+v calls=%d", w.Code, response, f.client.calls)
		}
	})

	t.Run("duplicate request protection", func(t *testing.T) {
		f := newAliyunAPITestFixture(t)
		body := `{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"15555550121"}`
		repeated := `{ "test_phone":"15555550121", "config": {"selected_params": [], "template_param_mode":"structured", "template_code":"SMS"}, "way_id":"WY_TEST_A", "template_id":"TP_TEST_A" }`
		_, first := f.request("/api/v1/templates/ins/aliyun/test", f.token, body)
		w, second := f.request("/api/v1/templates/ins/aliyun/test", f.token, repeated)
		if first.Code != http.StatusOK || second.Code != http.StatusTooManyRequests || w.Header().Get("Retry-After") == "" {
			t.Fatalf("duplicate guard incorrect: first=%d second=%d retry=%q", first.Code, second.Code, w.Header().Get("Retry-After"))
		}
		if f.client.calls != 1 {
			t.Fatalf("duplicate request reached provider: calls=%d", f.client.calls)
		}
	})

	t.Run("user and way rate limits", func(t *testing.T) {
		f := newAliyunAPITestFixture(t)
		for i := 0; i < models.AliyunSMSTestUserLimit+1; i++ {
			body := fmt.Sprintf(`{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"1555555%04d"}`, 1200+i)
			_, response := f.request("/api/v1/templates/ins/aliyun/test", f.token, body)
			if i < models.AliyunSMSTestUserLimit && response.Code != http.StatusOK {
				t.Fatalf("user request %d should be allowed, response=%+v", i+1, response)
			}
			if i == models.AliyunSMSTestUserLimit && response.Code != http.StatusTooManyRequests {
				t.Fatalf("11th user request should be rate limited, response=%+v", response)
			}
		}
		if f.client.calls != models.AliyunSMSTestUserLimit {
			t.Fatalf("user limiter allowed %d SDK calls, want %d", f.client.calls, models.AliyunSMSTestUserLimit)
		}
		var events int64
		if err := models.GetDB().Model(&models.AliyunSMSTestEvent{}).Where("user_hash <> ''").Count(&events).Error; err != nil || events != models.AliyunSMSTestUserLimit {
			t.Fatalf("provider calls and granted reservations should both equal the limit: calls=%d events=%d err=%v", f.client.calls, events, err)
		}
	})

	t.Run("business rejection response", func(t *testing.T) {
		f := newAliyunAPITestFixture(t)
		f.client.response = &dysmsapi.SendSmsResponse{Code: "isv.MOBILE_NUMBER_ILLEGAL", Message: "invalid number", RequestId: "rejected-request", BizId: "rejected-biz"}
		body := `{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"15555550123"}`
		_, response := f.request("/api/v1/templates/ins/aliyun/test", f.token, body)
		var result testAliyunSMSResponse
		_ = json.Unmarshal(response.Data, &result)
		if response.Code != http.StatusOK || result.Accepted || result.Code != "isv.MOBILE_NUMBER_ILLEGAL" || result.RequestID != "rejected-request" || result.BizID != "rejected-biz" {
			t.Fatalf("provider rejection response incorrect: %+v response=%+v", result, response)
		}
	})

	t.Run("transport error is sanitized", func(t *testing.T) {
		f := newAliyunAPITestFixture(t)
		f.client.err = errors.New("transport failed with DO_NOT_LEAK_TEST_SECRET")
		body := `{"template_id":"TP_TEST_A","way_id":"WY_TEST_A","config":{"template_code":"SMS","template_param_mode":"structured","selected_params":[]},"test_phone":"15555550124"}`
		w, response := f.request("/api/v1/templates/ins/aliyun/test", f.token, body)
		if w.Code != http.StatusOK || response.Code != http.StatusBadGateway {
			t.Fatalf("transport failure should be reported safely: http=%d response=%+v", w.Code, response)
		}
		if strings.Contains(w.Body.String(), "DO_NOT_LEAK_TEST_SECRET") {
			t.Fatalf("transport error leaked credentials: %s", w.Body.String())
		}
		var eventCount int64
		if err := models.GetDB().Model(&models.AliyunSMSTestEvent{}).Count(&eventCount).Error; err != nil || eventCount != 1 {
			t.Fatalf("provider transport failure should still consume one reservation: events=%d err=%v", eventCount, err)
		}
		var events []models.AliyunSMSTestEvent
		if err := models.GetDB().Find(&events).Error; err != nil {
			t.Fatal(err)
		}
		serialized, _ := json.Marshal(events)
		for _, secret := range []string{"15555550124", "DO_NOT_LEAK_TEST_SECRET", "SMS", "template_param_mode"} {
			if strings.Contains(string(serialized), secret) {
				t.Fatalf("rate-limit event leaked sensitive/config value %q", secret)
			}
		}
	})
}

func TestAliyunConfigMergeValidatesTypesAndPreservesUnknown(t *testing.T) {
	merged, raw, err := mergeSupportedAliyunConfig(
		[]byte(`{"phone_number":"15500000000","template_code":"SMS_OLD","unknown":{"keep":true}}`),
		[]byte(`{"template_code":"SMS_NEW"}`),
	)
	if err != nil {
		t.Fatal(err)
	}
	if merged.TemplateCode != "SMS_NEW" || strings.Contains(raw, "template_param_mode") || !strings.Contains(raw, `"unknown"`) {
		t.Fatalf("merge should preserve old mode and unknown keys: %s", raw)
	}
	if _, _, err := mergeSupportedAliyunConfig([]byte(`{}`), []byte(`{"template_param_mode":null}`)); err == nil {
		t.Fatal("null parameter mode must not silently switch a legacy instance")
	}
	if _, _, err := mergeSupportedAliyunConfig([]byte(`{}`), []byte(`{"selected_params":null}`)); err == nil {
		t.Fatal("null selected_params should be rejected")
	}
}

func TestMaskAliyunPhone(t *testing.T) {
	if got := maskAliyunPhone("+8615555550123"); got != "***0123" {
		t.Fatalf("unexpected masked phone %q", got)
	}
}
