package routers

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"message-nest/models"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"message-nest/routers/api/v1"
	"message-nest/service/send_message_service/unified/channels"
)

type mysqlRouteMockSMSClient struct {
	mu       sync.Mutex
	requests []*dysmsapi.SendSmsRequest
}

func (m *mysqlRouteMockSMSClient) SendSms(request *dysmsapi.SendSmsRequest) (*dysmsapi.SendSmsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	copyOfRequest := *request
	m.requests = append(m.requests, &copyOfRequest)
	return &dysmsapi.SendSmsResponse{Code: "OK", Message: "mock accepted", RequestId: "mock-route-request", BizId: "mock-route-biz"}, nil
}

func (m *mysqlRouteMockSMSClient) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.requests)
}

func (m *mysqlRouteMockSMSClient) last() *dysmsapi.SendSmsRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.requests) == 0 {
		return nil
	}
	copyOfRequest := *m.requests[len(m.requests)-1]
	return &copyOfRequest
}

func TestAliyunSMSFullRouterSharedLimiterWithMySQLCompatibleMock(t *testing.T) {
	dsn := os.Getenv("MN_ALIYUN_ROUTE_DSN")
	if dsn == "" {
		t.Skip("set MN_ALIYUN_ROUTE_DSN for full-router MySQL/MariaDB Mock integration")
	}
	prefix := "r" + fmt.Sprintf("%x", time.Now().UnixNano()%0xffffff) + "_"
	database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: prefix, SingularTable: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(32)
	t.Cleanup(func() { _ = sqlDB.Close() })
	models.SetDB(database)
	if err := database.AutoMigrate(&models.Template{}, &models.SendWays{}, &models.SendTasksIns{}, &models.SendTasksLogs{}, &models.SendStats{}); err != nil {
		t.Fatalf("migrate isolated route fixture tables: %v", err)
	}
	if err := models.MigrateAliyunSMSTestRateLimitTables(database); err != nil {
		t.Fatalf("migrate limiter tables: %v", err)
	}

	gin.SetMode(gin.TestMode)
	previousPrefix := setting.ServerSetting.UrlPrefix
	setting.ServerSetting.UrlPrefix = ""
	t.Cleanup(func() { setting.ServerSetting.UrlPrefix = previousPrefix })
	setting.AppSetting.JwtSecret = "isolated-mysql-router-jwt-secret"
	setting.AppSetting.AliyunSMSTestFingerprintSecret = "isolated-shared-rate-fingerprint-key-32-bytes"
	ownerID := "UR_MYSQL_ROUTE_OWNER"
	otherID := "UR_MYSQL_ROUTE_OTHER"
	ownerToken, err := util.GenerateToken(ownerID, "rate-owner", "test", "user", 1)
	if err != nil {
		t.Fatal(err)
	}
	otherToken, err := util.GenerateToken(otherID, "rate-other", "test", "user", 1)
	if err != nil {
		t.Fatal(err)
	}

	template := models.Template{
		UUIDModel:    models.UUIDModel{ID: "TPMYSQL001", CreatedBy: ownerID, ModifiedBy: ownerID},
		Name:         "isolated MySQL rate test",
		Status:       "enabled",
		TextTemplate: "Hello {{name}}",
		Placeholders: `[{"key":"name"}]`,
	}
	way := models.SendWays{
		UUIDModel: models.UUIDModel{ID: "WYMysql001", CreatedBy: ownerID, ModifiedBy: ownerID},
		Name:      "isolated provider mock",
		Type:      channels.MessageTypeAliyunSMS,
		Auth:      `{"access_key_id":"fake-access-id","access_key_secret":"fake-access-secret","sign_name":"MockSign","region_id":"cn-test"}`,
	}
	instance := models.SendTasksIns{
		UUIDModel:   models.UUIDModel{ID: "INMYSQL001", CreatedBy: ownerID, ModifiedBy: ownerID},
		TemplateID:  template.ID,
		WayID:       way.ID,
		WayType:     channels.MessageTypeAliyunSMS,
		ContentType: "text",
		Config:      `{"phone_number":"15500000000","template_code":"SMS_BEFORE_EDIT","template_param_mode":"structured","selected_params":["name"]}`,
	}
	for _, row := range []any{&template, &way, &instance} {
		if err := database.Create(row).Error; err != nil {
			t.Fatalf("seed isolated route resource: %v", err)
		}
	}

	mockClient := &mysqlRouteMockSMSClient{}
	sender := channels.NewAliyunSMSChannelWithClientFactory(func(accessKeyID, accessKeySecret, regionID string) (channels.AliyunSMSClient, error) {
		if accessKeyID != "fake-access-id" || accessKeySecret != "fake-access-secret" || regionID != "cn-test" {
			return nil, fmt.Errorf("unexpected isolated fixture credentials")
		}
		return mockClient, nil
	})
	router := initRouter(embed.FS{}, v1.NewAliyunSMSInstanceAPIWithSender(sender))

	var audit bytes.Buffer
	previousLoggerOutput := logrus.StandardLogger().Out
	logrus.SetOutput(&audit)
	t.Cleanup(func() { logrus.SetOutput(previousLoggerOutput) })

	request := func(path, token, body string) (*httptest.ResponseRecorder, map[string]any) {
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		router.ServeHTTP(recorder, req)
		var envelope map[string]any
		_ = json.Unmarshal(recorder.Body.Bytes(), &envelope)
		return recorder, envelope
	}

	pathEdit := "/api/v1/templates/ins/aliyun/edit"
	pathTest := "/api/v1/templates/ins/aliyun/test"
	unauthorized, _ := request(pathTest, "", `{"instance_id":"INMYSQL001","test_phone":"15555550123","placeholders":{"name":"private-vcode-991"}}`)
	if unauthorized.Code != http.StatusUnauthorized || mockClient.count() != 0 {
		t.Fatalf("full router must enforce JWT before provider call: status=%d calls=%d body=%s", unauthorized.Code, mockClient.count(), unauthorized.Body.String())
	}
	foreignTemplate := models.Template{
		UUIDModel: models.UUIDModel{ID: "TPMYSQL002", CreatedBy: otherID, ModifiedBy: otherID},
		Name:      "foreign route template", Status: "enabled", TextTemplate: "Hello {{name}}", Placeholders: `[{"key":"name"}]`,
	}
	foreignWay := models.SendWays{
		UUIDModel: models.UUIDModel{ID: "WYMysql002", CreatedBy: otherID, ModifiedBy: otherID},
		Name:      "foreign provider mock", Type: channels.MessageTypeAliyunSMS,
		Auth: `{"access_key_id":"fake-access-id","access_key_secret":"fake-access-secret","sign_name":"MockSign","region_id":"cn-test"}`,
	}
	if err := database.Create(&foreignTemplate).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&foreignWay).Error; err != nil {
		t.Fatal(err)
	}
	for _, body := range []string{
		`{"template_id":"TPMYSQL002","way_id":"WYMysql001","config":{"template_code":"SMS_DENIED","template_param_mode":"structured","selected_params":["name"]},"test_phone":"15555550126","placeholders":{"name":"private-vcode-991"}}`,
		`{"template_id":"TPMYSQL001","way_id":"WYMysql002","config":{"template_code":"SMS_DENIED","template_param_mode":"structured","selected_params":["name"]},"test_phone":"15555550127","placeholders":{"name":"private-vcode-991"}}`,
	} {
		denied, envelope := request(pathTest, ownerToken, body)
		if envelope["code"] != float64(http.StatusNotFound) || mockClient.count() != 0 {
			t.Fatalf("full router must reject foreign template/Way before provider call: status=%d envelope=%v calls=%d", denied.Code, envelope, mockClient.count())
		}
	}

	editRecorder, editEnvelope := request(pathEdit, ownerToken, `{"id":"INMYSQL001","config":{"template_code":"SMS_REAL_DB_MOCK","param_name_overrides":{"name":"customerName"}}}`)
	if editRecorder.Code != http.StatusOK || editEnvelope["code"] != float64(http.StatusOK) {
		t.Fatalf("full router Config edit failed: status=%d envelope=%v body=%s", editRecorder.Code, editEnvelope, editRecorder.Body.String())
	}
	var updated models.SendTasksIns
	if err := database.Where("id = ?", instance.ID).First(&updated).Error; err != nil {
		t.Fatal(err)
	}
	if updated.ID != instance.ID || updated.TemplateID != instance.TemplateID || updated.WayID != instance.WayID {
		t.Fatalf("edit changed immutable associations: %+v", updated)
	}
	var updatedConfig models.InsAliyunSMSConfig
	if err := json.Unmarshal([]byte(updated.Config), &updatedConfig); err != nil {
		t.Fatal(err)
	}
	if updatedConfig.TemplateCode != "SMS_REAL_DB_MOCK" || updatedConfig.ParamNameOverrides["name"] != "customerName" {
		t.Fatalf("edit config not persisted: %+v", updatedConfig)
	}

	testBody := func(phone string) string {
		body, _ := json.Marshal(map[string]any{
			"instance_id":  instance.ID,
			"test_phone":   phone,
			"placeholders": map[string]any{"name": "private-vcode-991"},
		})
		return string(body)
	}
	first, firstEnvelope := request(pathTest, ownerToken, testBody("15555550123"))
	data, _ := json.Marshal(firstEnvelope["data"])
	if first.Code != http.StatusOK || firstEnvelope["code"] != float64(http.StatusOK) || !strings.Contains(string(data), "mock-route-request") {
		t.Fatalf("full router Mock test failed: status=%d envelope=%v body=%s", first.Code, firstEnvelope, first.Body.String())
	}
	last := mockClient.last()
	if last == nil || last.TemplateCode != "SMS_REAL_DB_MOCK" || last.PhoneNumbers != "15555550123" || last.TemplateParam != `{"customerName":"private-vcode-991"}` {
		t.Fatalf("Mock SDK received unexpected request: %+v", last)
	}

	denied, deniedEnvelope := request(pathTest, otherToken, testBody("15555550124"))
	if deniedEnvelope["code"] != float64(http.StatusNotFound) || mockClient.count() != 1 {
		t.Fatalf("foreign instance access was not denied before provider: status=%d envelope=%v calls=%d", denied.Code, deniedEnvelope, mockClient.count())
	}

	for i := 1; i < models.AliyunSMSTestUserLimit; i++ {
		phone := fmt.Sprintf("15555550%04d", 1300+i)
		recorder, envelope := request(pathTest, ownerToken, testBody(phone))
		if recorder.Code != http.StatusOK || envelope["code"] != float64(http.StatusOK) {
			t.Fatalf("authorized route request %d should pass: status=%d envelope=%v", i+1, recorder.Code, envelope)
		}
	}
	rateLimited, rateEnvelope := request(pathTest, ownerToken, testBody("15555550999"))
	if rateLimited.Code != http.StatusTooManyRequests || rateEnvelope["code"] != float64(http.StatusTooManyRequests) || mockClient.count() != models.AliyunSMSTestUserLimit {
		t.Fatalf("full router rate-limit response/call count incorrect: status=%d envelope=%v calls=%d", rateLimited.Code, rateEnvelope, mockClient.count())
	}

	var events, logs, stats int64
	if err := database.Model(&models.AliyunSMSTestEvent{}).Count(&events).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&models.SendTasksLogs{}).Count(&logs).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&models.SendStats{}).Count(&stats).Error; err != nil {
		t.Fatal(err)
	}
	if events != models.AliyunSMSTestUserLimit || logs != 0 || stats != 0 {
		t.Fatalf("reservation/API and business stats counts mismatch: events=%d logs=%d stats=%d", events, logs, stats)
	}
	var storedConfig models.SendTasksIns
	if err := database.Where("id = ?", instance.ID).First(&storedConfig).Error; err != nil {
		t.Fatal(err)
	}
	if storedConfig.Config != updated.Config {
		t.Fatal("saved-instance test request modified the persisted Config")
	}
	var storedEvents []models.AliyunSMSTestEvent
	if err := database.Find(&storedEvents).Error; err != nil {
		t.Fatal(err)
	}
	eventJSON, _ := json.Marshal(storedEvents)
	for _, sensitive := range []string{"15555550123", "private-vcode-991", "fake-access-secret", "SMS_REAL_DB_MOCK"} {
		if strings.Contains(string(eventJSON), sensitive) {
			t.Fatalf("rate event contains sensitive/config value %q", sensitive)
		}
	}
	if strings.Contains(audit.String(), "15555550123") || strings.Contains(audit.String(), "private-vcode-991") || strings.Contains(audit.String(), "fake-access-secret") {
		t.Fatalf("API audit log leaked sensitive test data: %s", audit.String())
	}

	if err := database.Migrator().DropTable(&models.AliyunSMSTestEvent{}); err != nil {
		t.Fatal(err)
	}
	unavailable, unavailableEnvelope := request(pathTest, ownerToken, testBody("15555550888"))
	if unavailable.Code != http.StatusServiceUnavailable || unavailableEnvelope["code"] != float64(http.StatusServiceUnavailable) || mockClient.count() != models.AliyunSMSTestUserLimit {
		t.Fatalf("missing limiter table must return 503 without calling Mock: status=%d envelope=%v calls=%d", unavailable.Code, unavailableEnvelope, mockClient.count())
	}
}
