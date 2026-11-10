package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"message-nest/middleware"
	"message-nest/models"
	"message-nest/pkg/util"
)

func TestGetMessageTemplateEnforcesOwnerScopeAndAdminImpersonation(t *testing.T) {
	f := newAliyunAPITestFixture(t)
	foreign := models.Template{
		UUIDModel:    models.UUIDModel{ID: "TP_TEST_B", CreatedBy: f.otherID, ModifiedBy: f.otherID},
		Name:         "foreign template",
		Status:       "enabled",
		TextTemplate: "private body {{secret_value}}",
		Placeholders: `[{"key":"secret_value"}]`,
	}
	if err := models.GetDB().Create(&foreign).Error; err != nil {
		t.Fatal(err)
	}
	adminToken, err := util.GenerateToken("UR_ADMIN", "admin", "hash", "admin", 1)
	if err != nil {
		t.Fatal(err)
	}
	f.router.GET("/api/v1/templates/get", middleware.JWT(), GetMessageTemplate)

	request := func(id, token, impersonate string) modelsResponse {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/templates/get?id="+id, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		if impersonate != "" {
			req.Header.Set("X-Tenant-User-Id", impersonate)
		}
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, req)
		var response modelsResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		return response
	}

	if response := request(f.template.ID, f.token, ""); response.Code != http.StatusOK {
		t.Fatalf("ordinary user should read their own template: %+v", response)
	}
	if response := request(foreign.ID, f.token, ""); response.Code == http.StatusOK {
		t.Fatal("ordinary user must not read another owner's template by ID")
	}
	if response := request(foreign.ID, adminToken, ""); response.Code != http.StatusOK {
		t.Fatalf("global admin should retain global template access: %+v", response)
	}
	if response := request(foreign.ID, adminToken, f.ownerID); response.Code == http.StatusOK {
		t.Fatal("impersonating admin must not read a template outside the impersonated owner")
	}
}

func TestTemplateEditDeleteAndPreviewRespectImpersonatedOwner(t *testing.T) {
	f := newAliyunAPITestFixture(t)
	foreign := models.Template{
		UUIDModel:    models.UUIDModel{ID: "TP_TEST_B", CreatedBy: f.otherID, ModifiedBy: f.otherID},
		Name:         "foreign template",
		Status:       "enabled",
		TextTemplate: "private body {{secret_value}}",
		Placeholders: `[{"key":"secret_value"}]`,
	}
	if err := models.GetDB().Create(&foreign).Error; err != nil {
		t.Fatal(err)
	}
	adminID := "UR_ADMIN"
	adminToken, err := util.GenerateToken(adminID, "admin", "hash", "admin", 1)
	if err != nil {
		t.Fatal(err)
	}
	f.router.POST("/api/v1/templates/edit", middleware.JWT(), EditMessageTemplate)
	f.router.POST("/api/v1/templates/delete", middleware.JWT(), DeleteMessageTemplate)
	f.router.POST("/api/v1/templates/preview", middleware.JWT(), PreviewMessageTemplate)
	f.router.GET("/api/v1/templates/list", middleware.JWT(), GetMessageTemplateList)
	f.router.POST("/api/v1/templates/add", middleware.JWT(), AddMessageTemplate)

	request := func(path, token, tenantID, body string) (*httptest.ResponseRecorder, modelsResponse) {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		if tenantID != "" {
			req.Header.Set("X-Tenant-User-Id", tenantID)
		}
		w := httptest.NewRecorder()
		f.router.ServeHTTP(w, req)
		var response modelsResponse
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		return w, response
	}

	foreignEdit := `{"id":"TP_TEST_B","name":"changed","text_template":"changed"}`
	if recorder, response := request("/api/v1/templates/edit", f.token, "", foreignEdit); response.Code != http.StatusNotFound {
		t.Fatalf("ordinary user must not edit a foreign template: envelope=%+v body=%s", response, recorder.Body.String())
	}
	if recorder, response := request("/api/v1/templates/edit", adminToken, f.ownerID, foreignEdit); response.Code != http.StatusNotFound {
		t.Fatalf("impersonating admin must not edit outside the selected owner: envelope=%+v body=%s", response, recorder.Body.String())
	}
	if recorder, response := request("/api/v1/templates/delete", adminToken, f.ownerID, `{"id":"TP_TEST_B"}`); response.Code != http.StatusNotFound {
		t.Fatalf("impersonating admin must not delete outside the selected owner: envelope=%+v body=%s", response, recorder.Body.String())
	}
	previewBody := `{"id":"TP_TEST_B","params":{"secret_value":"do-not-disclose"}}`
	if recorder, response := request("/api/v1/templates/preview", f.token, "", previewBody); response.Code != http.StatusNotFound {
		t.Fatalf("ordinary user must not preview a foreign template: envelope=%+v body=%s", response, recorder.Body.String())
	}
	if recorder, response := request("/api/v1/templates/preview", adminToken, f.ownerID, previewBody); response.Code != http.StatusNotFound {
		t.Fatalf("impersonating admin must not preview outside the selected owner: envelope=%+v body=%s", response, recorder.Body.String())
	}
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/templates/list", nil)
	listReq.Header.Set("Authorization", "Bearer "+adminToken)
	listReq.Header.Set("X-Tenant-User-Id", f.ownerID)
	listRecorder := httptest.NewRecorder()
	f.router.ServeHTTP(listRecorder, listReq)
	if strings.Contains(listRecorder.Body.String(), "foreign template") || !strings.Contains(listRecorder.Body.String(), f.template.Name) {
		t.Fatalf("impersonating admin template list must remain in tenant scope: %s", listRecorder.Body.String())
	}
	addRecorder, addResponse := request("/api/v1/templates/add", adminToken, f.ownerID,
		`{"name":"created in tenant scope","text_template":"Hello"}`)
	if addResponse.Code != http.StatusOK {
		t.Fatalf("admin should add a template inside the impersonated owner scope: envelope=%+v body=%s", addResponse, addRecorder.Body.String())
	}
	var created models.Template
	if err := models.GetDB().Where("name = ?", "created in tenant scope").First(&created).Error; err != nil {
		t.Fatal(err)
	}
	if created.CreatedBy != f.ownerID || created.ModifiedBy != adminID {
		t.Fatalf("impersonated create should assign tenant owner and preserve real audit actor: created_by=%q modified_by=%q", created.CreatedBy, created.ModifiedBy)
	}

	var unchanged models.Template
	if err := models.GetDB().Where("id = ?", foreign.ID).First(&unchanged).Error; err != nil {
		t.Fatal(err)
	}
	if unchanged.Name != foreign.Name || unchanged.TextTemplate != foreign.TextTemplate {
		t.Fatalf("denied owner operations changed foreign template: %+v", unchanged)
	}
	if recorder, response := request("/api/v1/templates/edit", adminToken, f.ownerID,
		`{"id":"TP_TEST_A","name":"edited as impersonated owner","text_template":"Hello {{name}}"}`); response.Code != http.StatusOK {
		t.Fatalf("admin should be able to edit within the impersonated owner scope: envelope=%+v body=%s", response, recorder.Body.String())
	}
	var edited models.Template
	if err := models.GetDB().Where("id = ?", f.template.ID).First(&edited).Error; err != nil {
		t.Fatal(err)
	}
	if edited.CreatedBy != f.ownerID || edited.ModifiedBy != adminID {
		t.Fatalf("impersonated edit should preserve owner and audit real operator: created_by=%q modified_by=%q", edited.CreatedBy, edited.ModifiedBy)
	}
}

func TestTemplateInstanceAPIsEnforceOwnerScopeAndPersistCreator(t *testing.T) {
	f := newAliyunAPITestFixture(t)
	for _, way := range []models.SendWays{
		{UUIDModel: models.UUIDModel{ID: "WYPHASE4A001", CreatedBy: f.ownerID}, Name: "owner test way", Type: f.instance.WayType, Auth: `{"access_key_id":"fake","access_key_secret":"fake"}`},
		{UUIDModel: models.UUIDModel{ID: "WYPHASE4B001", CreatedBy: f.otherID}, Name: "foreign test way", Type: f.instance.WayType, Auth: `{"access_key_id":"fake","access_key_secret":"fake"}`},
	} {
		if err := models.GetDB().Create(&way).Error; err != nil {
			t.Fatal(err)
		}
	}
	legacyInstance := f.instance
	legacyInstance.ID = "INLEGACY0001"
	legacyInstance.CreatedBy = ""
	legacyInstance.ModifiedBy = ""
	legacyInstance.Config = `{"phone_number":"15500000000","template_code":"SMS_LEGACY","content":"legacy content"}`
	if err := models.GetDB().Create(&legacyInstance).Error; err != nil {
		t.Fatal(err)
	}
	f.router.GET("/api/v1/templates/ins/get", middleware.JWT(), GetTemplateWithIns)
	f.router.GET("/api/v1/templates/get", middleware.JWT(), GetMessageTemplate)
	f.router.POST("/api/v1/templates/ins/addone", middleware.JWT(), AddTemplateIns)

	_, own := requestTemplateGet(f, "/api/v1/templates/ins/get?id="+f.template.ID)
	if own.Code != http.StatusOK {
		t.Fatalf("owner should read their template instances: %+v", own)
	}
	var ownData struct {
		InsList []map[string]interface{} `json:"ins_list"`
	}
	if err := json.Unmarshal(own.Data, &ownData); err != nil {
		t.Fatal(err)
	}
	foundLegacy := false
	for _, ins := range ownData.InsList {
		if ins["id"] == legacyInstance.ID {
			foundLegacy = true
		}
	}
	if !foundLegacy {
		t.Fatal("legacy instance with empty created_by should remain visible when its template and way belong to the user")
	}
	_, legacyTest := f.request("/api/v1/templates/ins/aliyun/test", f.token,
		`{"instance_id":"INLEGACY0001","test_phone":"15555550126","placeholders":{"name":"legacy","code":"654321"}}`)
	if legacyTest.Code != http.StatusOK {
		t.Fatalf("owner should test a legacy instance whose template and way they own: %+v", legacyTest)
	}
	_, legacyEdit := f.request("/api/v1/templates/ins/aliyun/edit", f.token,
		`{"id":"INLEGACY0001","config":{"template_code":"SMS_LEGACY_EDIT"}}`)
	if legacyEdit.Code != http.StatusOK {
		t.Fatalf("owner should edit supported fields of an owned legacy instance: %+v", legacyEdit)
	}
	var editedLegacy models.SendTasksIns
	if err := models.GetDB().Where("id = ?", "INLEGACY0001").First(&editedLegacy).Error; err != nil {
		t.Fatal(err)
	}
	var legacyConfig map[string]interface{}
	if err := json.Unmarshal([]byte(editedLegacy.Config), &legacyConfig); err != nil {
		t.Fatal(err)
	}
	if legacyConfig["template_code"] != "SMS_LEGACY_EDIT" || legacyConfig["template_param_mode"] != nil || legacyConfig["content"] != "legacy content" {
		t.Fatalf("legacy edit changed compatibility mode/content: %s", editedLegacy.Config)
	}
	_, foreign := requestTemplateGet(f, "/api/v1/templates/ins/get?id=TP_TEST_B")
	if foreign.Code == http.StatusOK {
		t.Fatal("user must not read another owner's template or instances")
	}
	wNoToken := httptest.NewRecorder()
	f.router.ServeHTTP(wNoToken, httptest.NewRequest(http.MethodGet, "/api/v1/templates/ins/get?id=TP_TEST_A", nil))
	if wNoToken.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated template instance read should be rejected, got HTTP %d", wNoToken.Code)
	}
	_, foreignTemplate := f.request("/api/v1/templates/ins/addone", f.token,
		`{"id":"INFRGN000002","template_id":"TP_TEST_B","way_id":"WYPHASE4A001","way_type":"AliyunSMS","content_type":"text","config":"{}"}`)
	if foreignTemplate.Code == http.StatusOK {
		t.Fatal("user must not attach their way to another owner's template")
	}
	_, foreignWay := f.request("/api/v1/templates/ins/addone", f.token,
		`{"id":"INFRGN000001","template_id":"TP_TEST_A","way_id":"WYPHASE4B001","way_type":"AliyunSMS","content_type":"text","config":"{}"}`)
	if foreignWay.Code == http.StatusOK {
		t.Fatal("user must not attach another owner's way")
	}

	_, ownAdd := f.request("/api/v1/templates/ins/addone", f.token,
		`{"id":"INOWNER00001","template_id":"TP_TEST_A","way_id":"WYPHASE4A001","way_type":"AliyunSMS","content_type":"text","config":"{\"phone_number\":\"15500000000\",\"template_code\":\"SMS_ADD\"}"}`)
	if ownAdd.Code != http.StatusOK {
		t.Fatalf("owner should add an instance to their template and way: %+v", ownAdd)
	}
	var created models.SendTasksIns
	if err := models.GetDB().Where("id = ?", "INOWNER00001").First(&created).Error; err != nil {
		t.Fatal(err)
	}
	if created.CreatedBy != f.ownerID || created.ModifiedBy != f.ownerID {
		t.Fatalf("new instance audit ownership not persisted: created_by=%q modified_by=%q", created.CreatedBy, created.ModifiedBy)
	}

}

func requestTemplateGet(f *aliyunAPITestFixture, path string) (*httptest.ResponseRecorder, modelsResponse) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer "+f.token)
	f.router.ServeHTTP(w, req)
	var response modelsResponse
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	return w, response
}
