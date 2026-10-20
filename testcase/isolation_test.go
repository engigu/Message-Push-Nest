package testcase

import (
	"context"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/service/cron_msg_service"
	"message-nest/service/hosted_message_service"
	"message-nest/service/message_template_service"
	"message-nest/service/send_logs_service"
	"message-nest/service/send_task_service"
	"message-nest/service/send_way_service"
	"message-nest/service/settings_service"
	"message-nest/service/statistic_service"
	"testing"
)

func setupIsolationDB(testName string) (userA, userB, admin *models.Auth) {
	models.InitTestDB("isolation_test_" + testName)

	userA, _ = models.AddUser("userA_"+testName, "passA", "user", "system")
	userB, _ = models.AddUser("userB_"+testName, "passB", "user", "system")
	admin, _ = models.AddUser("admin_"+testName, "adminPass", "admin", "system")
	return
}

func TestDataIsolationAcrossTenants(t *testing.T) {
	userA, userB, admin := setupIsolationDB("basic")

	// ==========================================
	// 1. 发信渠道隔离测试
	// ==========================================
	wayASvc := send_way_service.SendWay{
		Name:       "Email Channel A",
		Type:       "email",
		Auth:       `{"smtp_host":"smtp.a.com"}`,
		CreatedBy:  userA.ID,
		ModifiedBy: userA.ID,
		IsAdmin:    false,
	}
	err := wayASvc.Add()
	if err != nil {
		t.Fatalf("User A add send way failed: %v", err)
	}
	wayAObj, err := models.GetWayByName("Email Channel A")
	if err != nil || wayAObj.ID == "" {
		t.Fatalf("GetWayByName failed: %v", err)
	}
	wayAID := wayAObj.ID

	// User B 查询渠道列表：应该看不到 User A 的渠道
	wayBSvc := send_way_service.SendWay{
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	waysB, err := wayBSvc.GetAll()
	if err != nil {
		t.Fatalf("User B get ways failed: %v", err)
	}
	for _, w := range waysB {
		if w.ID == wayAID {
			t.Fatalf("Tenant isolation breach: User B can see User A's channel %s", wayAID)
		}
	}

	// User B 试图修改 User A 的渠道：应被拦截
	wayBHackSvc := send_way_service.SendWay{
		ID:         wayAID,
		Name:       "Hacked Name",
		ModifiedBy: userB.ID,
		IsAdmin:    false,
	}
	err = wayBHackSvc.Edit()
	if err == nil {
		t.Fatalf("Tenant security breach: User B was able to edit User A's channel")
	}

	// User B 试图删除 User A 的渠道：应被拦截
	wayBHackDelete := send_way_service.SendWay{
		ID:         wayAID,
		ModifiedBy: userB.ID,
		IsAdmin:    false,
	}
	err = wayBHackDelete.Delete()
	if err == nil {
		t.Fatalf("Tenant security breach: User B was able to delete User A's channel")
	}

	// 管理员查询：能够看到 User A 的渠道
	wayAdminSvc := send_way_service.SendWay{
		CreatedBy: admin.ID,
		IsAdmin:   true,
	}
	waysAdmin, err := wayAdminSvc.GetAll()
	if err != nil {
		t.Fatalf("Admin get ways failed: %v", err)
	}
	foundWayA := false
	for _, w := range waysAdmin {
		if w.ID == wayAID {
			foundWayA = true
			break
		}
	}
	if !foundWayA {
		t.Fatalf("Admin should be able to see all channels, but Way A not found")
	}

	// ==========================================
	// 2. 发信任务隔离测试
	// ==========================================
	taskAID := models.GenerateTaskUniqueID()
	taskASvc := send_task_service.SendTaskService{
		ID:         taskAID,
		Name:       "Task for User A",
		CreatedBy:  userA.ID,
		ModifiedBy: userA.ID,
		IsAdmin:    false,
	}
	err = taskASvc.AddWithID()
	if err != nil {
		t.Fatalf("User A add task failed: %v", err)
	}

	// User B 查询任务列表：应该看不到 User A 的任务
	taskBSvc := send_task_service.SendTaskService{
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	tasksB, err := taskBSvc.GetAll()
	if err != nil {
		t.Fatalf("User B get tasks failed: %v", err)
	}
	for _, tk := range tasksB {
		if tk.ID == taskAID {
			t.Fatalf("Tenant isolation breach: User B can see User A's task %s", taskAID)
		}
	}

	// User B 试图修改 User A 的任务：应被拦截
	taskBHackSvc := send_task_service.SendTaskService{
		ID:         taskAID,
		ModifiedBy: userB.ID,
		IsAdmin:    false,
	}
	err = taskBHackSvc.Edit(map[string]interface{}{"name": "Hacked Task"})
	if err == nil {
		t.Fatalf("Tenant security breach: User B was able to edit User A's task")
	}

	// User B 试图删除 User A 的任务：应被拦截
	err = taskBHackSvc.Delete()
	if err == nil {
		t.Fatalf("Tenant security breach: User B was able to delete User A's task")
	}

	// ==========================================
	// 3. 消息模板隔离测试
	// ==========================================
	tmplASvc := message_template_service.TemplateService{
		Name:         "Template A",
		TextTemplate: "Hello {{name}}",
		CreatedBy:    userA.ID,
		ModifiedBy:   userA.ID,
		IsAdmin:      false,
	}
	err = tmplASvc.Add()
	if err != nil {
		t.Fatalf("User A add template failed: %v", err)
	}
	tmplAID := tmplASvc.ID

	// User B 查询模板列表：应该看不到 User A 的模板
	tmplBSvc := message_template_service.TemplateService{
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	tmplsB, err := tmplBSvc.GetAll()
	if err != nil {
		t.Fatalf("User B get templates failed: %v", err)
	}
	for _, tp := range tmplsB {
		if tp.ID == tmplAID {
			t.Fatalf("Tenant isolation breach: User B can see User A's template %s", tmplAID)
		}
	}

	// User B 试图修改 User A 的模板：应被拦截
	tmplBHackSvc := message_template_service.TemplateService{
		ID:           tmplAID,
		Name:         "Hacked Template",
		TextTemplate: "Hacked",
		ModifiedBy:   userB.ID,
		IsAdmin:      false,
	}
	err = tmplBHackSvc.Update()
	if err == nil {
		t.Fatalf("Tenant security breach: User B was able to edit User A's template")
	}

	// User B 试图删除 User A 的模板：应被拦截
	tmplBDeleteSvc := message_template_service.TemplateService{
		ID:         tmplAID,
		ModifiedBy: userB.ID,
		IsAdmin:    false,
	}
	err = tmplBDeleteSvc.Delete()
	if err == nil {
		t.Fatalf("Tenant security breach: User B was able to delete User A's template")
	}

	// ==========================================
	// 4. 定时消息隔离测试
	// ==========================================
	cronASvc := cron_msg_service.CronMsgService{
		Name:      "Cron A",
		TaskID:    taskAID,
		Cron:      "0 * * * *",
		Title:     "Hourly Msg",
		Content:   "Time to check",
		CreatedBy: userA.ID,
		IsAdmin:   false,
	}
	cronAID, err := cronASvc.Add()
	if err != nil {
		t.Fatalf("User A add cron msg failed: %v", err)
	}

	// User B 查询定时消息：应该看不到 User A 的定时消息
	cronBSvc := cron_msg_service.CronMsgService{
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	cronsB, err := cronBSvc.GetAll()
	if err != nil {
		t.Fatalf("User B get cron msgs failed: %v", err)
	}
	for _, cm := range cronsB {
		if cm.ID == cronAID {
			t.Fatalf("Tenant isolation breach: User B can see User A's cron msg %s", cronAID)
		}
	}

	// User B 试图编辑 User A 的定时消息：应被拦截
	cronBHackSvc := cron_msg_service.CronMsgService{
		ID:        cronAID,
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	err = cronBHackSvc.Edit(map[string]interface{}{"title": "Hacked Title"})
	if err == nil {
		t.Fatalf("Tenant security breach: User B was able to edit User A's cron msg")
	}

	// User B 试图删除 User A 的定时消息：应被拦截
	err = cronBHackSvc.Delete()
	if err == nil {
		t.Fatalf("Tenant security breach: User B was able to delete User A's cron msg")
	}

	// ==========================================
	// 5. 发信日志与托管消息隔离测试
	// ==========================================
	logA := models.SendTasksLogs{
		TaskID:    taskAID,
		Type:      "task",
		Name:      "Task for User A",
		Log:       "Sending log for User A",
		CreatedBy: userA.ID,
	}
	_ = logA.Add()

	msgHostA := models.HostedMessage{
		Title:     "Hosted Msg for A",
		Content:   "Secret content for A",
		CreatedBy: userA.ID,
	}
	_ = msgHostA.Add()

	// User B 查询日志：应看不到 User A 的日志
	logBSvc := send_logs_service.SendTaskLogsService{
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	logsB, _ := logBSvc.GetAll()
	if len(logsB) != 0 {
		t.Fatalf("Tenant isolation breach: User B can see %d logs (expected 0)", len(logsB))
	}

	// User B 查询托管消息：应看不到 User A 的托管消息
	hostBSvc := hosted_message_service.HostMessageService{
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	hostedB, _ := hostBSvc.GetAll()
	if len(hostedB) != 0 {
		t.Fatalf("Tenant isolation breach: User B can see %d hosted messages (expected 0)", len(hostedB))
	}

	// Admin 查询日志与托管消息：能够看到
	logAdminSvc := send_logs_service.SendTaskLogsService{
		CreatedBy: admin.ID,
		IsAdmin:   true,
	}
	logsAdmin, _ := logAdminSvc.GetAll()
	if len(logsAdmin) != 1 {
		t.Fatalf("Admin expected to see 1 log, got %d", len(logsAdmin))
	}

	// ==========================================
	// 6. 数据大屏统计隔离测试
	// ==========================================
	_ = models.IncrementSendStats(taskAID, "task", "2026-10-20", "success", 10, userA.ID)

	statBSvc := statistic_service.StatisticService{
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	statB, err := statBSvc.GetSendStatsData()
	if err != nil {
		t.Fatalf("User B get stats failed: %v", err)
	}
	if statB.TotalNum != 0 {
		t.Fatalf("Tenant isolation breach: User B saw total num %d, expected 0", statB.TotalNum)
	}

	statAdminSvc := statistic_service.StatisticService{
		CreatedBy: admin.ID,
		IsAdmin:   true,
	}
	statAdmin, err := statAdminSvc.GetSendStatsData()
	if err != nil {
		t.Fatalf("Admin get stats failed: %v", err)
	}
	if statAdmin.TotalNum != 10 {
		t.Fatalf("Admin expected total num 10, got %d", statAdmin.TotalNum)
	}

	// 针对具体任务统计隔离测试：User B 试图获取 User A 的 taskAID 统计
	taskStatBSvc := statistic_service.StatisticService{
		TaskID:    taskAID,
		CreatedBy: userB.ID,
		IsAdmin:   false,
	}
	taskStatB, err := taskStatBSvc.GetSendStatsByTask()
	if err != nil {
		t.Fatalf("User B GetSendStatsByTask failed: %v", err)
	}
	if taskStatB.TotalNum != 0 {
		t.Fatalf("Tenant isolation breach: User B saw task A's total num %d, expected 0", taskStatB.TotalNum)
	}

	// Admin 获取 taskAID 的任务统计：应能看到 10
	taskStatAdminSvc := statistic_service.StatisticService{
		TaskID:    taskAID,
		CreatedBy: admin.ID,
		IsAdmin:   true,
	}
	taskStatAdmin, err := taskStatAdminSvc.GetSendStatsByTask()
	if err != nil {
		t.Fatalf("Admin GetSendStatsByTask failed: %v", err)
	}
	if taskStatAdmin.TotalNum != 10 {
		t.Fatalf("Admin expected task A's total num 10, got %d", taskStatAdmin.TotalNum)
	}

	// ==========================================
	// 7. 配置表 (message_settings) 隔离与系统级回退测试
	// ==========================================
	settingASvc := settings_service.UserSettings{
		UserID:  userA.ID,
		IsAdmin: false,
	}
	_ = settingASvc.EditSettings("user_pref", "color", "red", userA.ID)

	settingBSvc := settings_service.UserSettings{
		UserID:  userB.ID,
		IsAdmin: false,
	}
	_ = settingBSvc.EditSettings("user_pref", "color", "blue", userB.ID)

	// User A 读取自己的 user_pref: 得到 color = red
	prefA, err := settingASvc.GetUserSetting("user_pref")
	if err != nil || prefA["color"] != "red" {
		t.Fatalf("expected User A color = 'red', got %v", prefA["color"])
	}

	// User B 读取自己的 user_pref: 得到 color = blue
	prefB, err := settingBSvc.GetUserSetting("user_pref")
	if err != nil || prefB["color"] != "blue" {
		t.Fatalf("expected User B color = 'blue', got %v", prefB["color"])
	}

	// Admin 设置全局系统配置
	settingAdminSvc := settings_service.UserSettings{
		UserID:  admin.ID,
		IsAdmin: true,
	}
	_ = settingAdminSvc.EditSettings("global_notice", "banner", "System Maintenance Soon", admin.ID)

	// 普通用户 User A 读取全局配置，查不到私有配置时应平滑回退读取全局系统配置
	globalNoticeForA, err := settingASvc.GetUserSetting("global_notice")
	if err != nil || globalNoticeForA["banner"] != "System Maintenance Soon" {
		t.Fatalf("User A failed to fallback-read global notice: %v", globalNoticeForA)
	}
}

// TestAdminTenantImpersonation 测试管理员穿越 Context（租户模拟与代客运维全生命周期）
func TestAdminTenantImpersonation(t *testing.T) {
	userA, userB, admin := setupIsolationDB("impersonation")

	// 1. 管理员普通模式：全局全量
	adminContext := app.WithUserContext(context.Background(), app.UserContextInfo{
		UserID:   admin.ID,
		Username: admin.Username,
		Role:     admin.Role,
	})

	// 2. 管理员穿越模式：穿越到 User A
	adminImpersonateAContext := app.WithUserContext(context.Background(), app.UserContextInfo{
		UserID:            admin.ID,
		Username:          admin.Username,
		Role:              admin.Role,
		ImpersonateUserID: userA.ID, // 穿越为 User A
	})

	// 2.1 管理员代客为 User A 创建渠道
	impersonateWay := models.SendWays{
		UUIDModel: models.UUIDModel{ID: "WAY_IMPER_A"},
		Name:      "Channel For A By Admin",
		Type:      "email",
		Auth:      `{}`,
	}
	err := models.GetDB().WithContext(adminImpersonateAContext).Create(&impersonateWay).Error
	if err != nil {
		t.Fatalf("Admin impersonating A create way failed: %v", err)
	}

	// 验证资产所有权归被代建人 User A，而审计溯源真实记录管理员
	var savedWay models.SendWays
	_ = models.GetDB().First(&savedWay, "id = ?", impersonateWay.ID)
	if savedWay.CreatedBy != userA.ID {
		t.Fatalf("Impersonate ownership mismatch: expected %s, got %s", userA.ID, savedWay.CreatedBy)
	}
	if savedWay.ModifiedBy != admin.ID {
		t.Fatalf("Impersonate audit mismatch: expected %s, got %s", admin.ID, savedWay.ModifiedBy)
	}

	// 2.2 租户 User B 自建一条渠道
	userBContext := app.WithUserContext(context.Background(), app.UserContextInfo{
		UserID:   userB.ID,
		Username: userB.Username,
		Role:     userB.Role,
	})
	wayB := models.SendWays{
		UUIDModel: models.UUIDModel{ID: "WAY_USER_B"},
		Name:      "Channel By B",
		Type:      "webhook",
		Auth:      `{}`,
	}
	_ = models.GetDB().WithContext(userBContext).Create(&wayB)

	// 2.3 管理员在穿越视角下查询（模拟 User A 的视角）：仅能看到 User A 的渠道，看不到 User B
	var waysUnderA []models.SendWays
	err = models.GetDB().WithContext(adminImpersonateAContext).
		Scopes(models.TenantScope(adminImpersonateAContext)).
		Find(&waysUnderA).Error
	if err != nil {
		t.Fatalf("Query ways under Impersonate A failed: %v", err)
	}
	for _, w := range waysUnderA {
		if w.ID == "WAY_USER_B" {
			t.Fatalf("Impersonate TenantScope breach: Admin simulating A saw User B's channel")
		}
	}

	// 2.4 管理员切回全局视角：能够看到所有渠道（包括 A 和 B）
	var waysAll []models.SendWays
	err = models.GetDB().WithContext(adminContext).
		Scopes(models.TenantScope(adminContext)).
		Find(&waysAll).Error
	if err != nil {
		t.Fatalf("Query ways under global admin failed: %v", err)
	}
	if len(waysAll) < 2 {
		t.Fatalf("Global admin expected >= 2 ways, got %d", len(waysAll))
	}
}
