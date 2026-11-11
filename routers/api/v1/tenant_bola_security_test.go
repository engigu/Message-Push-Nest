package v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"message-nest/middleware"
	"message-nest/models"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	v2 "message-nest/routers/api/v2"
)

type securityTestFixture struct {
	router      *gin.Engine
	userAID     string
	userAToken  string
	userBID     string
	userBToken  string
	adminID     string
	adminToken  string
	wayAID      string
	taskAID     string
	taskBID     string
	cronAID     string
	insAID      string
	templateAID string
}

func newSecurityTestFixture(t *testing.T) *securityTestFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	models.InitTestDB("tenant_security_" + util.GenerateUniqueID())
	setting.AppSetting.JwtSecret = "SecurityTestSecretKey123"

	userAID := "UR_USER_A001"
	userBID := "UR_USER_B001"
	adminID := "UR_ADMIN_001"

	tokenA, err := util.GenerateToken(userAID, "user_a", "hash", "user", 1)
	if err != nil {
		t.Fatal(err)
	}
	tokenB, err := util.GenerateToken(userBID, "user_b", "hash", "user", 1)
	if err != nil {
		t.Fatal(err)
	}
	tokenAdmin, err := util.GenerateToken(adminID, "admin", "hash", "admin", 1)
	if err != nil {
		t.Fatal(err)
	}

	// 1. 初始化 User A 的资产
	wayA := models.SendWays{
		UUIDModel: models.UUIDModel{ID: "WY_TEST_A001", CreatedBy: userAID, ModifiedBy: userAID},
		Name:      "User A Private SMTP",
		Type:      "email",
		Auth:      `{"server":"smtp.privateserver.com","port":465,"account":"userA@private.com","passwd":"SUPER_SECRET_A_PASSWORD"}`,
	}
	taskA := models.SendTasks{
		UUIDModel: models.UUIDModel{ID: "TK_TEST_A001", CreatedBy: userAID, ModifiedBy: userAID},
		Name:      "User A Private Task",
	}
	insA := models.SendTasksIns{
		UUIDModel:   models.UUIDModel{ID: "IN_TEST_A001", CreatedBy: userAID, ModifiedBy: userAID},
		TaskID:      taskA.ID,
		WayID:       wayA.ID,
		WayType:     wayA.Type,
		ContentType: "text",
		Config:      `{"to_account":"boss@company.com"}`,
		Enable:      1,
	}
	cronA := models.CronMessages{
		UUIDModel: models.UUIDModel{ID: "CM_TEST_A001", CreatedBy: userAID, ModifiedBy: userAID},
		Name:      "User A Secret Cron",
		TaskID:    taskA.ID,
		Cron:      "0 9 * * *",
		Title:     "Daily Secret",
		Content:   "Secret Report for A",
		Enable:    1,
	}

	// 2. 初始化 User B 的空任务（用于测试跨租户偷渠道）
	taskB := models.SendTasks{
		UUIDModel: models.UUIDModel{ID: "TK_TEST_B001", CreatedBy: userBID, ModifiedBy: userBID},
		Name:      "User B Public Task",
	}

	// 3. 初始化 User A 的私有模板及其实例
	templateA := models.Template{
		UUIDModel:    models.UUIDModel{ID: "TP_TEST_A001", CreatedBy: userAID, ModifiedBy: userAID},
		Name:         "User A Private Template",
		TextTemplate: "Hello {{name}} from A",
		Status:       "enabled",
	}
	insTmplA := models.SendTasksIns{
		UUIDModel:   models.UUIDModel{ID: "IN_TMPL_A001", CreatedBy: userAID, ModifiedBy: userAID},
		TemplateID:  templateA.ID,
		WayID:       wayA.ID,
		WayType:     wayA.Type,
		ContentType: "text",
		Config:      `{"to_account":"userA_secret_recipient@test.com"}`,
		Enable:      1,
	}

	// 4. 初始化系统管理员配置的敏感安全配置
	adminSecSetting := models.Settings{
		Section:    "security",
		Key:        "oidc_client_secret",
		Value:      "SUPER_ADMIN_OIDC_SECRET_999",
		CreatedBy:  adminID,
		ModifiedBy: adminID,
	}

	for _, item := range []interface{}{&wayA, &taskA, &insA, &cronA, &taskB, &templateA, &insTmplA, &adminSecSetting} {
		if err := models.GetDB().Create(item).Error; err != nil {
			t.Fatalf("failed to seed test fixture: %v", err)
		}
	}

	// 5. 构建带 JWT 的测试路由
	router := gin.New()
	api := router.Group("/api/v1")
	api.Use(middleware.JWT())
	{
		// 发信渠道
		api.GET("/sendways/get", GetMsgSendWay)
		api.GET("/sendways/list", GetMsgSendWayList)
		api.POST("/sendways/edit", EditSendWay)
		api.POST("/sendways/delete", DeleteMsgSendWay)

		// 发信任务
		api.GET("/sendtasks/get", GetMsgSendTask)
		api.GET("/sendtasks/list", GetMsgSendTaskList)
		api.POST("/sendtasks/edit", EditMsgSendTask)
		api.POST("/sendtasks/delete", DeleteMsgSendTask)

		// 发信任务实例
		api.GET("/sendtasks/ins/gettask", GetMsgSendWayIns)
		api.POST("/sendtasks/ins/addone", AddTasksIns)
		api.POST("/sendtasks/ins/addmany", AddManyTasksIns)
		api.POST("/sendtasks/ins/delete", DeleteMsgTaskIns)

		// 定时消息
		api.GET("/cronmessages/list", GetCronMsgList)
		api.POST("/cronmessages/addone", AddCronMsgTask)
		api.POST("/cronmessages/edit", EditCronMsgTask)
		api.POST("/cronmessages/delete", DeleteCronMsgTask)
		api.POST("/cronmessages/sendnow", SendNowCronMsg)

		// 消息模板
		api.GET("/templates/get", GetMessageTemplate)
		api.POST("/templates/add", AddMessageTemplate)
		api.GET("/templates/ins/get", GetTemplateWithIns)

		// 系统设置
		api.GET("/settings/getsetting", GetUserSetting)

		// 发信
		api.POST("/message/send", DoSendMassage)

		// 统计
		api.GET("/statistic/task", GetSendStatsByTask)
	}

	apiV2 := router.Group("/api/v2")
	apiV2.Use(middleware.JWT())
	{
		apiV2.POST("/message/send", v2.DoSendMessageByTemplate)
	}

	return &securityTestFixture{
		router:      router,
		userAID:     userAID,
		userAToken:  tokenA,
		userBID:     userBID,
		userBToken:  tokenB,
		adminID:     adminID,
		adminToken:  tokenAdmin,
		wayAID:      wayA.ID,
		taskAID:     taskA.ID,
		taskBID:     taskB.ID,
		cronAID:     cronA.ID,
		insAID:      insA.ID,
		templateAID: templateA.ID,
	}
}

func (f *securityTestFixture) exec(method, path, token, impersonateID, body string) (*httptest.ResponseRecorder, modelsResponse) {
	var bodyReader *strings.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	} else {
		bodyReader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if impersonateID != "" {
		req.Header.Set("X-Tenant-User-Id", impersonateID)
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	var resp modelsResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return w, resp
}

// -----------------------------------------------------------------------------
// 1. 发信渠道防水平越权与凭据保护测试
// -----------------------------------------------------------------------------

func TestSendWaysTenantIsolationAndCredentialDefense(t *testing.T) {
	f := newSecurityTestFixture(t)

	// User A 可以读取自己的渠道（应能看到配置）
	w, resp := f.exec(http.MethodGet, "/api/v1/sendways/get?id="+f.wayAID, f.userAToken, "", "")
	if resp.Code != http.StatusOK || !strings.Contains(w.Body.String(), "SUPER_SECRET_A_PASSWORD") {
		t.Fatalf("User A should read their own channel credentials: %+v, body=%s", resp, w.Body.String())
	}

	// 攻击：User B 试图根据 ID 读取 User A 的渠道凭据（BOLA / IDOR 探测）
	wB, respB := f.exec(http.MethodGet, "/api/v1/sendways/get?id="+f.wayAID, f.userBToken, "", "")
	if respB.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B read User A's channel info! code=%d body=%s", respB.Code, wB.Body.String())
	}
	if strings.Contains(wB.Body.String(), "SUPER_SECRET_A_PASSWORD") {
		t.Fatal("CRITICAL: User A's channel secret password leaked to User B!")
	}

	// 攻击：User B 试图修改 User A 的渠道
	_, respEdit := f.exec(http.MethodPost, "/api/v1/sendways/edit", f.userBToken, "",
		`{"id":"`+f.wayAID+`","name":"Hacked Name","type":"email","auth":"{}"}`)
	if respEdit.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B was able to call edit on User A's channel! code=%d", respEdit.Code)
	}

	// 攻击：User B 试图删除 User A 的渠道
	_, respDel := f.exec(http.MethodPost, "/api/v1/sendways/delete", f.userBToken, "", `{"id":"`+f.wayAID+`"}`)
	if respDel.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B was able to call delete on User A's channel! code=%d", respDel.Code)
	}

	// 列表隔离：User B 渠道列表绝不能包含 User A 的渠道
	wList, _ := f.exec(http.MethodGet, "/api/v1/sendways/list", f.userBToken, "", "")
	if strings.Contains(wList.Body.String(), f.wayAID) || strings.Contains(wList.Body.String(), "User A Private SMTP") {
		t.Fatalf("Tenant isolation breach: User B saw User A's channel in list: %s", wList.Body.String())
	}

	// 管理员正常能看
	_, respAdmin := f.exec(http.MethodGet, "/api/v1/sendways/get?id="+f.wayAID, f.adminToken, "", "")
	if respAdmin.Code != http.StatusOK {
		t.Fatalf("Global admin should be able to read channel: %+v", respAdmin)
	}

	// 管理员穿越 User B 视角：受限于 User B，读 User A 渠道也必须返回 404
	_, respAdminImperB := f.exec(http.MethodGet, "/api/v1/sendways/get?id="+f.wayAID, f.adminToken, f.userBID, "")
	if respAdminImperB.Code != http.StatusNotFound {
		t.Fatalf("Impersonating admin should be constrained by target tenant B: %+v", respAdminImperB)
	}
}

// -----------------------------------------------------------------------------
// 2. 发信任务与任务实例防水平越权测试
// -----------------------------------------------------------------------------

func TestSendTasksAndInstancesIsolation(t *testing.T) {
	f := newSecurityTestFixture(t)

	// 攻击：User B 试图读取 User A 的任务详情
	_, respTaskGet := f.exec(http.MethodGet, "/api/v1/sendtasks/get?id="+f.taskAID, f.userBToken, "", "")
	if respTaskGet.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B accessed User A's task by ID! code=%d", respTaskGet.Code)
	}

	// 攻击：User B 试图读取 User A 任务下的所有实例（含敏感配置）
	_, respInsGet := f.exec(http.MethodGet, "/api/v1/sendtasks/ins/gettask?id="+f.taskAID, f.userBToken, "", "")
	if respInsGet.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B accessed User A's task instances! code=%d", respInsGet.Code)
	}

	// 攻击：User B 试图将实例挂载到 User A 的任务下
	_, respAddOtherTask := f.exec(http.MethodPost, "/api/v1/sendtasks/ins/addone", f.userBToken, "",
		`{"id":"IN_ATTACK_01","task_id":"`+f.taskAID+`","way_id":"`+f.wayAID+`","way_type":"email","content_type":"text","config":"{}"}`)
	if respAddOtherTask.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B was able to attach instance to User A's task! code=%d", respAddOtherTask.Code)
	}

	// 攻击：User B 在自己的任务 taskB 下，盗用 User A 的渠道 wayA（白嫖付费渠道漏洞）
	_, respStealWay := f.exec(http.MethodPost, "/api/v1/sendtasks/ins/addone", f.userBToken, "",
		`{"id":"IN_ATTACK_02","task_id":"`+f.taskBID+`","way_id":"`+f.wayAID+`","way_type":"email","content_type":"text","config":"{}"}`)
	if respStealWay.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B was able to bind User A's send way! code=%d", respStealWay.Code)
	}

	// 攻击：批量添加接口同样必须防御盗用他人渠道
	_, respManySteal := f.exec(http.MethodPost, "/api/v1/sendtasks/ins/addmany", f.userBToken, "",
		`{"id":"TK_ATTACK_01","name":"Attack Task","ins_data":[{"id":"IN_ATTACK_03","task_id":"TK_ATTACK_01","way_id":"`+f.wayAID+`","way_type":"email","content_type":"text","config":"{}"}]}`)
	if respManySteal.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B was able to bind User A's way via addmany! code=%d", respManySteal.Code)
	}

	// 攻击：User B 试图删除 User A 的任务实例
	_, respDelIns := f.exec(http.MethodPost, "/api/v1/sendtasks/ins/delete", f.userBToken, "", `{"id":"`+f.insAID+`"}`)
	if respDelIns.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B called delete on User A's instance! code=%d", respDelIns.Code)
	}
}

// -----------------------------------------------------------------------------
// 3. 定时消息防跨租户借用与任意发信测试
// -----------------------------------------------------------------------------

func TestCronMessagesSecurityAndUnauthorizedSendNow(t *testing.T) {
	f := newSecurityTestFixture(t)

	// 攻击：User B 试图通过 sendnow 冒用 User A 的 taskA 立即发信
	_, respSendNow := f.exec(http.MethodPost, "/api/v1/cronmessages/sendnow", f.userBToken, "",
		`{"task_id":"`+f.taskAID+`","title":"Spoofed Message","content":"Hacked via sendnow"}`)
	if respSendNow.Code != http.StatusNotFound {
		t.Fatalf("CRITICAL SECURITY BREACH: User B triggered unauthorized message send using User A's task! code=%d", respSendNow.Code)
	}

	// 攻击：User B 新建定时任务时绑定 User A 的 taskA
	_, respAddCron := f.exec(http.MethodPost, "/api/v1/cronmessages/addone", f.userBToken, "",
		`{"name":"Malicious Cron","task_id":"`+f.taskAID+`","cron":"0 0 * * *","title":"Malicious","content":"Payload"}`)
	if respAddCron.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B created cron message bound to User A's task! code=%d", respAddCron.Code)
	}

	// 攻击：User B 编辑定时任务时篡改目标为 User A 的 taskA
	_, respEditCron := f.exec(http.MethodPost, "/api/v1/cronmessages/edit", f.userBToken, "",
		`{"id":"`+f.cronAID+`","name":"Hacked Cron","task_id":"`+f.taskAID+`","cron":"0 0 * * *","title":"Hacked","content":"Payload","enable":1}`)
	if respEditCron.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B edited User A's cron message! code=%d", respEditCron.Code)
	}

	// 攻击：User B 试图删除 User A 的定时任务
	_, respDelCron := f.exec(http.MethodPost, "/api/v1/cronmessages/delete", f.userBToken, "", `{"id":"`+f.cronAID+`"}`)
	if respDelCron.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B deleted User A's cron message! code=%d", respDelCron.Code)
	}

	// 列表隔离：User B 定时任务列表不得出现 User A 的 cron
	wList, _ := f.exec(http.MethodGet, "/api/v1/cronmessages/list", f.userBToken, "", "")
	if strings.Contains(wList.Body.String(), f.cronAID) || strings.Contains(wList.Body.String(), "User A Secret Cron") {
		t.Fatalf("Tenant isolation breach: User B saw User A's cron in list: %s", wList.Body.String())
	}
}

// -----------------------------------------------------------------------------
// 4. 统计报表防跨租户探查测试
// -----------------------------------------------------------------------------

func TestStatisticTaskOwnerIsolation(t *testing.T) {
	f := newSecurityTestFixture(t)

	// 攻击：User B 试图读取 User A 的任务发送统计
	_, respStat := f.exec(http.MethodGet, "/api/v1/statistic/task?task_id="+f.taskAID, f.userBToken, "", "")
	if respStat.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B viewed statistics of User A's task! code=%d", respStat.Code)
	}

	// User A 可以读取自己的统计
	_, respStatA := f.exec(http.MethodGet, "/api/v1/statistic/task?task_id="+f.taskAID, f.userAToken, "", "")
	if respStatA.Code != http.StatusOK {
		t.Fatalf("User A should be able to view their own task statistics: %+v", respStatA)
	}
}

// -----------------------------------------------------------------------------
// 5. 消息模板防跨租户克隆与实例越权盗用测试
// -----------------------------------------------------------------------------

func TestMessageTemplateCloneTenantDefense(t *testing.T) {
	f := newSecurityTestFixture(t)

	// 攻击：User B 试图通过克隆模板窃取 User A 的模板及其绑定的发信实例（包含 User A 的渠道 wayA）
	_, respClone := f.exec(http.MethodPost, "/api/v1/templates/add", f.userBToken, "",
		`{"name":"User B Cloned Template","text_template":"Hacked text","source_template_id":"`+f.templateAID+`","copy_instances":true}`)
	if respClone.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B was able to clone User A's template instances! code=%d", respClone.Code)
	}

	// 验证：数据库中绝不应存在归属于 User B 且关联 wayA 的实例
	var leakedInsCount int64
	models.GetDB().Model(&models.SendTasksIns{}).
		Where("created_by = ? AND way_id = ?", f.userBID, f.wayAID).
		Count(&leakedInsCount)
	if leakedInsCount > 0 {
		t.Fatalf("CRITICAL: User A's channel instance was leaked/cloned to User B! count=%d", leakedInsCount)
	}

	// User A 自己克隆自己的模板正常允许
	_, respCloneA := f.exec(http.MethodPost, "/api/v1/templates/add", f.userAToken, "",
		`{"name":"User A Legitimate Clone","text_template":"Self clone","source_template_id":"`+f.templateAID+`","copy_instances":true}`)
	if respCloneA.Code != http.StatusOK {
		t.Fatalf("User A should be able to clone their own template: %+v", respCloneA)
	}
}

// -----------------------------------------------------------------------------
// 6. 系统配置防垂直越权与凭据保护测试
// -----------------------------------------------------------------------------

func TestUserSettingsVerticalDefense(t *testing.T) {
	f := newSecurityTestFixture(t)

	// 攻击：普通用户 B 试图获取 security 敏感配置（探查 OIDC / OAuth2 Client Secret 等系统机密）
	wB, respB := f.exec(http.MethodGet, "/api/v1/settings/getsetting?section=security", f.userBToken, "", "")
	if respB.Code != http.StatusForbidden {
		t.Fatalf("Vertical privilege escalation: Regular user B accessed system security settings! code=%d body=%s", respB.Code, wB.Body.String())
	}
	if strings.Contains(wB.Body.String(), "SUPER_ADMIN_OIDC_SECRET_999") {
		t.Fatal("CRITICAL: Administrator OIDC secret leaked to regular user B!")
	}

	// 允许：普通用户 B 访问自己的日志清理配置或公开站点配置
	_, respSite := f.exec(http.MethodGet, "/api/v1/settings/getsetting?section=site_config", f.userBToken, "", "")
	if respSite.Code != http.StatusOK {
		t.Fatalf("Regular user B should be able to read site_config: %+v", respSite)
	}

	// 允许：管理员访问 security 配置
	wAdmin, respAdmin := f.exec(http.MethodGet, "/api/v1/settings/getsetting?section=security", f.adminToken, "", "")
	if respAdmin.Code != http.StatusOK || !strings.Contains(wAdmin.Body.String(), "SUPER_ADMIN_OIDC_SECRET_999") {
		t.Fatalf("Admin should be able to read security settings: %+v, body=%s", respAdmin, wAdmin.Body.String())
	}
}

// -----------------------------------------------------------------------------
// 7. 发信接口与 Token 防跨租户盗用测试
// -----------------------------------------------------------------------------

func TestSendMessageTenantBOLAAndTokenDefense(t *testing.T) {
	f := newSecurityTestFixture(t)

	// 攻击 1：User B 调用 /api/v1/message/send，传参 task_id=taskA（试图盗用 User A 的任务发信）
	_, respSendTaskID := f.exec(http.MethodPost, "/api/v1/message/send", f.userBToken, "",
		`{"task_id":"`+f.taskAID+`","title":"Attack Title","text":"Attack Text"}`)
	if respSendTaskID.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B triggered send on User A's task via task_id! code=%d", respSendTaskID.Code)
	}

	// 攻击 2：User B 将 User A 的 taskA 编码为 token，调用 /api/v1/message/send
	encTokenTaskA := util.EncryptTokenHex(f.taskAID, 71)
	_, respSendToken := f.exec(http.MethodPost, "/api/v1/message/send", f.userBToken, "",
		`{"token":"`+encTokenTaskA+`","title":"Attack Title","text":"Attack Text"}`)
	if respSendToken.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B triggered send on User A's task via encrypted token! code=%d", respSendToken.Code)
	}

	// 攻击 3：User B 调用 /api/v2/message/send，传入用 User A 模板 ID 加密的 token（试图盗用 User A 的模板发信）
	encTokenTmplA := util.EncryptTokenHex(f.templateAID, 71)
	_, respSendTmplToken := f.exec(http.MethodPost, "/api/v2/message/send", f.userBToken, "",
		`{"token":"`+encTokenTmplA+`","title":"Attack Title","placeholders":{"name":"hacker"}}`)
	if respSendTmplToken.Code != http.StatusNotFound {
		t.Fatalf("Tenant security breach: User B triggered send on User A's template via v2 token! code=%d", respSendTmplToken.Code)
	}
}
