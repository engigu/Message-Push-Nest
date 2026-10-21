package v1

import (
	"bytes"
	"encoding/json"
	"message-nest/middleware"
	"message-nest/models"
	"message-nest/pkg/captcha"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupUserRouter(suffix ...string) (*gin.Engine, *models.Auth, *models.Auth) {
	gin.SetMode(gin.TestMode)
	dbName := "router_user_test_" + util.GenerateUniqueID()
	if len(suffix) > 0 && suffix[0] != "" {
		dbName = "router_user_test_" + suffix[0]
	}
	models.InitTestDB(dbName)
	setting.AppSetting.JwtSecret = "RouterTestSecretKey123"

	// 初始化测试账号
	adminUser, _ := models.AddUser("admin", "adminPass", "admin", "system")
	normalUser, _ := models.AddUser("normal_user", "userPass", "user", "system")

	r := gin.New()

	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.JWT())
	{
		// 管理员专用路由组
		adminGroup := apiV1.Group("")
		adminGroup.Use(middleware.AdminRequired())
		{
			adminGroup.GET("/users", GetUserList)
			adminGroup.POST("/users", AddUser)
			adminGroup.PUT("/users", EditUser)
			adminGroup.DELETE("/users", DeleteUser)
			adminGroup.POST("/users/reset-passwd", ResetUserPassword)
		}
	}

	return r, adminUser, normalUser
}

func TestUserAPIAuthAndPermissions(t *testing.T) {
	router, adminUser, normalUser := setupUserRouter()

	adminToken, _ := util.GenerateToken(adminUser.ID, adminUser.Username, adminUser.Password, adminUser.Role, 1)
	userToken, _ := util.GenerateToken(normalUser.ID, normalUser.Username, normalUser.Password, normalUser.Role, 1)

	// 1. 无 Token 访问 -> 401
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/users", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for request without token, got %d", w.Code)
	}

	// 2. 普通用户访问管理员接口 -> 403 Forbidden
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for normal user accessing admin route, got %d", w.Code)
	}

	// 3. 管理员访问 -> 200 OK
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for admin accessing user list, got %d, body: %s", w.Code, w.Body.String())
	}

	// 4. 管理员添加新用户 -> 200 OK
	addUserBody := map[string]interface{}{
		"username": "newbie",
		"password": "password123",
		"email":    "newbie@example.com",
		"role":     "user",
	}
	bodyBytes, _ := json.Marshal(addUserBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for AddUser, got %d, body: %s", w.Code, w.Body.String())
	}

	// 查出新建的用户
	newbie, err := models.GetUserByUsername("newbie")
	if err != nil || newbie.ID == "" {
		t.Fatalf("newbie was not created: %v", err)
	}

	// 5. 管理员修改用户 -> 200 OK
	editUserBody := map[string]interface{}{
		"id":     newbie.ID,
		"email":  "newbie_updated@example.com",
		"role":   "user",
		"status": 1,
	}
	bodyBytes, _ = json.Marshal(editUserBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/users", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for EditUser, got %d, body: %s", w.Code, w.Body.String())
	}

	// 6. 管理员重置密码 -> 200 OK
	resetPasswdBody := map[string]interface{}{
		"id":       newbie.ID,
		"password": "newPassword789",
	}
	bodyBytes, _ = json.Marshal(resetPasswdBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/users/reset-passwd", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for ResetUserPassword, got %d, body: %s", w.Code, w.Body.String())
	}

	// 7. 管理员删除用户 -> 200 OK
	deleteUserBody := map[string]interface{}{
		"id": newbie.ID,
	}
	bodyBytes, _ = json.Marshal(deleteUserBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/users", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for DeleteUser, got %d, body: %s", w.Code, w.Body.String())
	}

	// 8. 管理员试图删除自己 -> 400 Bad Request
	deleteSelfBody := map[string]interface{}{
		"id": adminUser.ID,
	}
	bodyBytes, _ = json.Marshal(deleteSelfBody)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/users", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var respBody map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &respBody)
	if respCode, ok := respBody["code"].(float64); !ok || int(respCode) != 400 {
		t.Fatalf("expected business code 400 when deleting self, got %v, body: %s", respBody["code"], w.Body.String())
	}
}

func TestSSOBindAndRegisterAPIs(t *testing.T) {
	router, _, normalUser := setupUserRouter()
	router.POST("/api/v1/auth/sso/bind-login", SSOBindLogin)
	router.POST("/api/v1/auth/sso/register-bind", SSORegisterBind)

	// 1. 测试未绑定第三方身份通过 Ticket 关联已有内部账号
	ticket1, _ := util.GenerateSSOTicket("oidc", "corp_sub_alice_123", "alice@corp.com", "alice_corp")

	// 1.1 密码错误测试 -> 业务 code 400
	wrongPassBody, _ := json.Marshal(map[string]string{
		"ticket":   ticket1,
		"username": normalUser.Username,
		"password": "wrongPassword",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/sso/bind-login", bytes.NewBuffer(wrongPassBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var resp1 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp1)
	if code, ok := resp1["code"].(float64); !ok || int(code) != 400 {
		t.Fatalf("expected business code 400 for wrong password, got %v, body: %s", resp1["code"], w.Body.String())
	}

	// 1.2 正确账号密码关联 -> 业务 code 200
	correctPassBody, _ := json.Marshal(map[string]string{
		"ticket":   ticket1,
		"username": normalUser.Username,
		"password": "userPass",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/sso/bind-login", bytes.NewBuffer(correctPassBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var resp2 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp2)
	if code, ok := resp2["code"].(float64); !ok || int(code) != 200 {
		t.Fatalf("expected business code 200 for bind-login, got %v, body: %s", resp2["code"], w.Body.String())
	}

	// 验证绑定关系生效
	boundUser, err := models.GetUserBySSO("oidc", "corp_sub_alice_123")
	if err != nil || boundUser.ID != normalUser.ID {
		t.Fatalf("GetUserBySSO failed after bind-login: %+v", boundUser)
	}

	// 2. 测试未绑定第三方身份通过 Ticket 创建内部新账号并关联
	ticket2, _ := util.GenerateSSOTicket("oauth2", "gh_sub_bob_999", "bob@github.com", "github_bob")

	// 2.1 未开启注册时 -> 业务 code 403
	_ = models.AddOneSetting(models.Settings{
		Section: "site_config",
		Key:     "allow_registration",
		Value:   "false",
	})
	regBody, _ := json.Marshal(map[string]string{
		"ticket":   ticket2,
		"username": "internal_bob",
		"password": "password456",
		"email":    "bob@github.com",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/sso/register-bind", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var resp3 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp3)
	if code, ok := resp3["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected business code 403 when registration is closed, got %v, body: %s", resp3["code"], w.Body.String())
	}

	// 2.2 开启注册后正常注册并绑定 -> 业务 code 200
	_ = models.DeleteSettingByKey("site_config", "allow_registration")
	_ = models.AddOneSetting(models.Settings{
		Section: "site_config",
		Key:     "allow_registration",
		Value:   "true",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/sso/register-bind", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var resp4 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp4)
	if code, ok := resp4["code"].(float64); !ok || int(code) != 200 {
		t.Fatalf("expected business code 200 for register-bind when enabled, got %v, body: %s", resp4["code"], w.Body.String())
	}

	// 验证新建内部账号且绑定成功
	createdBob, err := models.GetUserBySSO("oauth2", "gh_sub_bob_999")
	if err != nil || createdBob.Username != "internal_bob" || createdBob.Role != "user" {
		t.Fatalf("GetUserBySSO failed after register-bind: %+v", createdBob)
	}

	// 3. 测试公开免鉴权用户名可用性检查接口
	router.GET("/api/v1/auth/check-username", CheckUsernameAvailability)

	// 3.1 查询已被使用的用户名 -> available = false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/check-username?username=internal_bob", nil)
	router.ServeHTTP(w, req)
	var resp5 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp5)
	data5 := resp5["data"].(map[string]interface{})
	if data5["available"].(bool) {
		t.Fatalf("expected 'internal_bob' to be unavailable, got available=true")
	}

	// 3.2 大小写不敏感排重检查 -> available = false
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/check-username?username=INTERNAL_BOB", nil)
	router.ServeHTTP(w, req)
	var resp6 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp6)
	data6 := resp6["data"].(map[string]interface{})
	if data6["available"].(bool) {
		t.Fatalf("expected 'INTERNAL_BOB' to be unavailable due to case-insensitivity")
	}

	// 3.3 未被使用的合法用户名 -> available = true
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/check-username?username=fresh_brand_new_user", nil)
	router.ServeHTTP(w, req)
	var resp7 map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp7)
	data7 := resp7["data"].(map[string]interface{})
	if !data7["available"].(bool) {
		t.Fatalf("expected 'fresh_brand_new_user' to be available")
	}
}

func TestDemoModeRestrictions(t *testing.T) {
	// 设置环境变量激活演示模式
	_ = os.Setenv("DEMO_MODE", "true")
	defer func() {
		_ = os.Unsetenv("DEMO_MODE")
	}()

	gin.SetMode(gin.TestMode)
	models.InitTestDB("demo_mode_test_" + util.GenerateUniqueID())

	r := gin.New()
	r.POST("/api/v1/auth/register", Register)
	r.GET("/api/v1/public/site/config", GetPublicSiteConfig)
	r.GET("/api/v1/auth/sso/status", GetSSOStatus)
	r.POST("/api/v1/auth/sso/bind-login", SSOBindLogin)
	r.POST("/api/v1/auth/sso/register-bind", SSORegisterBind)

	// 1. 尝试注册 -> 应该直接返回 403 Forbidden
	regBody, _ := json.Marshal(map[string]string{
		"username": "demo_hacker",
		"password": "password123",
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(regBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var regResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &regResp)
	if code, ok := regResp["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected business code 403 when registering in demo mode, got %v, body: %s", regResp["code"], w.Body.String())
	}

	// 2. 获取公开站点配置 -> allow_registration 必须为 false，is_demo 必须为 true
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/public/site/config", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for site config, got %d", w.Code)
	}
	var siteResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &siteResp)
	siteData := siteResp["data"].(map[string]interface{})
	if siteData["allow_registration"].(bool) {
		t.Fatalf("expected allow_registration to be false in demo mode")
	}
	if !siteData["is_demo"].(bool) {
		t.Fatalf("expected is_demo to be true in demo mode")
	}

	// 3. 获取 SSO 状态 -> enabled 必须为 false，is_demo 必须为 true，providers 必须为空
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/sso/status", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for sso status, got %d", w.Code)
	}
	var ssoResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &ssoResp)
	ssoData := ssoResp["data"].(map[string]interface{})
	if ssoData["enabled"].(bool) {
		t.Fatalf("expected SSO enabled to be false in demo mode")
	}
	if !ssoData["is_demo"].(bool) {
		t.Fatalf("expected SSO is_demo to be true in demo mode")
	}
	providers := ssoData["providers"].([]interface{})
	if len(providers) != 0 {
		t.Fatalf("expected SSO providers to be empty in demo mode, got %d", len(providers))
	}

	// 4. 尝试 SSO 绑定登录 / 创建账号 -> 应该直接返回 403 Forbidden
	bindBody, _ := json.Marshal(map[string]string{
		"ticket":   "dummy_ticket",
		"username": "demo_user",
		"password": "pwd",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/sso/bind-login", bytes.NewBuffer(bindBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var bindResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &bindResp)
	if code, ok := bindResp["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected business code 403 for SSOBindLogin in demo mode, got %v, body: %s", bindResp["code"], w.Body.String())
	}

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/sso/register-bind", bytes.NewBuffer(bindBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var regBindResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &regBindResp)
	if code, ok := regBindResp["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected business code 403 for SSORegisterBind in demo mode, got %v, body: %s", regBindResp["code"], w.Body.String())
	}
}

func TestUserCleanSettingsAndIsolation(t *testing.T) {
	router, adminUser, normalUser := setupUserRouter("clean_isolation")

	apiV1 := router.Group("/api/v1")
	apiV1.POST("/settings/set", EditSettings)
	apiV1.GET("/settings/getsetting", GetUserSetting)

	adminToken, _ := util.GenerateToken(adminUser.ID, adminUser.Username, adminUser.Password, adminUser.Role, 1)
	userToken, _ := util.GenerateToken(normalUser.ID, normalUser.Username, normalUser.Password, normalUser.Role, 1)

	// 1. 普通用户允许修改用户级设置：log_config
	logConfigBody, _ := json.Marshal(map[string]interface{}{
		"section": "log_config",
		"data": map[string]string{
			"cron":     "10 2 * * *",
			"keep_num": "500",
			"enabled":  "true",
		},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/settings/set", bytes.NewBuffer(logConfigBody))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for user modifying log_config, got %d, body: %s", w.Code, w.Body.String())
	}

	// 2. 普通用户尝试修改系统级设置：site_config -> 403 Forbidden
	siteConfigBody, _ := json.Marshal(map[string]interface{}{
		"section": "site_config",
		"data": map[string]string{
			"title": "Hacked Title",
		},
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/settings/set", bytes.NewBuffer(siteConfigBody))
	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	var respBody map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &respBody)
	if code, ok := respBody["code"].(float64); !ok || int(code) != 403 {
		t.Fatalf("expected business code 403 for normal user modifying site_config, got %v, body: %s", respBody["code"], w.Body.String())
	}

	// 2.1 验证普通用户获取自身数据清理配置时已分配专属随机打散时间
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/settings/getsetting?section=hosted_msg_config", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GetUserSetting, got %d", w.Code)
	}
	var getResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &getResp)
	getData := getResp["data"].(map[string]interface{})
	if getData["cron"] == "" || getData["cron"] == "30 0 * * *" {
		t.Fatalf("expected randomized cron for user, but got: %v", getData["cron"])
	}

	// 3. 管理员修改 site_config -> 200 OK
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/settings/set", bytes.NewBuffer(siteConfigBody))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for admin modifying site_config, got %d, body: %s", w.Code, w.Body.String())
	}

	// 4. 数据层按用户清理隔离测试
	// 4.1 日志隔离：为 admin 和 normalUser 分别插入 3 条日志
	statusVal := 0
	for i := 1; i <= 3; i++ {
		logA := models.SendTasksLogs{
			TaskID:    "T-USER-A",
			Log:       "User A log",
			Status:    &statusVal,
			CreatedBy: normalUser.ID,
		}
		_ = logA.Add()

		logB := models.SendTasksLogs{
			TaskID:    "T-USER-B",
			Log:       "User B log",
			Status:    &statusVal,
			CreatedBy: adminUser.ID,
		}
		_ = logB.Add()
	}

	// 执行仅清理 normalUser 的日志，保留 1 条
	affectedA, err := models.DeleteOutDateLogs(1, normalUser.ID)
	if err != nil {
		t.Fatalf("failed to delete outdate logs for normalUser: %v", err)
	}
	if affectedA != 2 {
		t.Fatalf("expected 2 logs deleted for normalUser, got %d", affectedA)
	}

	// 检查 admin 的 3 条日志是否完好无损
	adminLogsCount, _ := models.GetSendLogsTotal("", "", map[string]interface{}{"created_by": adminUser.ID})
	if adminLogsCount != 3 {
		t.Fatalf("admin logs should remain 3, but got %d", adminLogsCount)
	}

	// 4.2 托管消息隔离：为 admin 和 normalUser 分别插入 3 条托管消息
	for i := 1; i <= 3; i++ {
		msgA := models.HostedMessage{
			Title:     "Title A",
			Content:   "Content A",
			CreatedBy: normalUser.ID,
		}
		_ = msgA.Add()

		msgB := models.HostedMessage{
			Title:     "Title B",
			Content:   "Content B",
			CreatedBy: adminUser.ID,
		}
		_ = msgB.Add()
	}

	// 执行仅清理 normalUser 的托管消息，保留 1 条
	affectedMsgA, err := models.DeleteOutDateHostedMessages(1, normalUser.ID)
	if err != nil {
		t.Fatalf("failed to delete outdate hosted messages for normalUser: %v", err)
	}
	if affectedMsgA != 2 {
		t.Fatalf("expected 2 hosted messages deleted for normalUser, got %d", affectedMsgA)
	}

	// 检查 admin 的 3 条托管消息是否完好无损
	adminMsgCount, _ := models.GetHostMessagesTotal("", map[string]interface{}{"created_by": adminUser.ID})
	if adminMsgCount != 3 {
		t.Fatalf("admin hosted messages should remain 3, but got %d", adminMsgCount)
	}
}

func TestCaptchaAndRegistrationFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	models.InitTestDB("captcha_flow_test_" + util.GenerateUniqueID())

	// 开放自主注册
	_ = models.AddOneSetting(models.Settings{
		Section: "site_config",
		Key:     "allow_registration",
		Value:   "true",
	})

	r := gin.New()
	r.GET("/api/v1/auth/captcha", GetCaptcha)
	r.POST("/api/v1/auth/captcha/verify", VerifyCaptchaAndGetToken)
	r.POST("/api/v1/auth/register", Register)
	r.GET("/api/v1/auth/check-username", CheckUsernameAvailability)

	// 1. 获取验证码
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/auth/captcha?type=math", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for GetCaptcha, got %d", w.Code)
	}
	var capResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &capResp)
	capData := capResp["data"].(map[string]interface{})
	captchaID := capData["captcha_id"].(string)
	if captchaID == "" || capData["captcha_image"].(string) == "" {
		t.Fatalf("captcha_id or captcha_image is empty")
	}

	// 2. 无验证码尝试注册 -> 应该被拦截 (400)
	noCapBody, _ := json.Marshal(map[string]string{
		"username": "bot_attacker",
		"password": "password123",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(noCapBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var noCapResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &noCapResp)
	if code, ok := noCapResp["code"].(float64); !ok || int(code) != 400 {
		t.Fatalf("expected 400 for register without captcha, got %v", noCapResp["code"])
	}

	// 3. 错误验证码尝试注册 -> 应该被拦截 (400)
	wrongCapBody, _ := json.Marshal(map[string]string{
		"username":     "bot_attacker",
		"password":     "password123",
		"captcha_id":   captchaID,
		"captcha_code": "999999",
	})
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(wrongCapBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var wrongCapResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &wrongCapResp)
	if code, ok := wrongCapResp["code"].(float64); !ok || int(code) != 400 {
		t.Fatalf("expected 400 for register with wrong captcha, got %v", wrongCapResp["code"])
	}

	// 4. 使用合法人机验证 Token 注册 -> 200 OK
	validToken := util.GenerateUniqueID() // mock or token
	// 采用真实签发的 Token
	token := captcha.GenerateVerificationToken("")
	validRegBody, _ := json.Marshal(map[string]string{
		"username":      "legit_user",
		"password":      "password123",
		"email":         "legit@example.com",
		"captcha_token": token,
	})
	_ = validToken
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(validRegBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var validResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &validResp)
	if code, ok := validResp["code"].(float64); !ok || int(code) != 200 {
		t.Fatalf("expected 200 for register with valid captcha token, got %v, body: %s", validResp["code"], w.Body.String())
	}

	// 5. 校验 CheckUsernameAvailability 的 exclude_id 访客保护
	// 5.1 未登录时传入 exclude_id 不应该生效探测内部 ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/auth/check-username?username=legit_user&exclude_id=ANY_ID", nil)
	r.ServeHTTP(w, req)
	var chkResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &chkResp)
	chkData := chkResp["data"].(map[string]interface{})
	if chkData["available"].(bool) {
		t.Fatalf("expected legit_user to be unavailable even with exclude_id when not logged in")
	}
}



