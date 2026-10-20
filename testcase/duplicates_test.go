package testcase

import (
	"message-nest/models"
	"message-nest/service/send_way_service"
	"testing"
)

// TestTenantNamespaceDuplicateCheck 验证多租户命名空间下的同名检测与隔离
func TestTenantNamespaceDuplicateCheck(t *testing.T) {
	models.InitTestDB("tenant_namespace_dup_test")

	userA := "UR_USER_DUP_A"
	userB := "UR_USER_DUP_B"

	// ==========================================
	// 1. 发信渠道同名检测：同租户内排重，跨租户隔离
	// ==========================================
	wayASvc := send_way_service.SendWay{
		Name:       "Channel_Notification",
		Type:       "webhook",
		Auth:       `{}`,
		CreatedBy:  userA,
		ModifiedBy: userA,
		IsAdmin:    false,
	}

	// 1.1 租户 A 首次添加 "Channel_Notification" -> 应成功
	err := wayASvc.Add()
	if err != nil {
		t.Fatalf("User A first AddSendWay failed: %v", err)
	}

	// 1.2 租户 A 再次尝试添加同名 "Channel_Notification" -> 应被 NameIsExist 拦截
	wayADupSvc := send_way_service.SendWay{
		Name:       "Channel_Notification",
		Type:       "webhook",
		Auth:       `{}`,
		CreatedBy:  userA,
		ModifiedBy: userA,
		IsAdmin:    false,
	}
	err = wayADupSvc.Add()
	if err == nil {
		t.Fatalf("Expected duplicate channel name error for User A, but succeeded")
	}

	// 1.3 租户 B 添加同名 "Channel_Notification" -> 应该成功（不同租户互不冲突）
	wayBSvc := send_way_service.SendWay{
		Name:       "Channel_Notification",
		Type:       "webhook",
		Auth:       `{}`,
		CreatedBy:  userB,
		ModifiedBy: userB,
		IsAdmin:    false,
	}
	err = wayBSvc.Add()
	if err != nil {
		t.Fatalf("User B should be able to create channel with same name as User A: %v", err)
	}

	// 1.4 租户 A 编辑当前渠道保持原名 -> 应允许通过
	wayAObj, err := models.GetWayByName("Channel_Notification", userA)
	if err != nil || wayAObj.ID == "" {
		t.Fatalf("GetWayByName for User A failed: %v", err)
	}
	wayAEditSelf := send_way_service.SendWay{
		ID:         wayAObj.ID,
		Name:       "Channel_Notification",
		Auth:       `{"token":"new"}`,
		ModifiedBy: userA,
		IsAdmin:    false,
	}
	err = wayAEditSelf.Edit()
	if err != nil {
		t.Fatalf("Editing self with unchanged name should succeed: %v", err)
	}

	// 1.5 租户 A 添加第二个渠道并试图改名为第一个的名字 -> 应被拦截
	wayA2 := send_way_service.SendWay{
		Name:       "Channel_Second",
		Type:       "webhook",
		Auth:       `{}`,
		CreatedBy:  userA,
		ModifiedBy: userA,
		IsAdmin:    false,
	}
	_ = wayA2.Add()
	wayA2Obj, _ := models.GetWayByName("Channel_Second", userA)

	wayA2EditToA1 := send_way_service.SendWay{
		ID:         wayA2Obj.ID,
		Name:       "Channel_Notification", // 试图改名为已有的 Channel_Notification
		ModifiedBy: userA,
		IsAdmin:    false,
	}
	err = wayA2EditToA1.Edit()
	if err == nil {
		t.Fatalf("Renaming to existing channel name should fail for User A")
	}
}

// TestDatabaseTableNoUnintendedDuplicates 校验数据库核心表中不存在未预期的同租户重名数据
func TestDatabaseTableNoUnintendedDuplicates(t *testing.T) {
	models.InitTestDB("db_check_duplicates_test")
	db := models.GetDB()

	// 构造测试数据：插入多个租户的不同数据
	_ = models.AddSendWay("Way_1", `{}`, "webhook", "U1", "U1")
	_ = models.AddSendWay("Way_2", `{}`, "webhook", "U1", "U1")
	_ = models.AddSendWay("Way_1", `{}`, "webhook", "U2", "U2") // 跨租户同名合法

	_ = models.AddSendTask("Task_1", "U1")
	_ = models.AddSendTask("Task_2", "U1")
	_ = models.AddSendTask("Task_1", "U2") // 跨租户同名合法

	tmpl1 := models.Template{
		UUIDModel: models.UUIDModel{ID: "TP001", CreatedBy: "U1"},
		Name:      "Tmpl_1",
	}
	_ = tmpl1.Add()

	tmpl2 := models.Template{
		UUIDModel: models.UUIDModel{ID: "TP002", CreatedBy: "U2"},
		Name:      "Tmpl_1", // 跨租户同名合法
	}
	_ = tmpl2.Add()

	type DupCount struct {
		CreatedBy string
		Name      string
		Cnt       int
	}

	// 1. 检查发信渠道表
	var wayDups []DupCount
	wayTable := models.GetSchema(models.SendWays{})
	db.Table(wayTable).
		Select("created_by, name, count(*) as cnt").
		Group("created_by, name").
		Having("count(*) > 1").
		Scan(&wayDups)
	if len(wayDups) > 0 {
		t.Errorf("Found unexpected duplicates in %s: %+v", wayTable, wayDups)
	}

	// 2. 检查发信任务表
	var taskDups []DupCount
	taskTable := models.GetSchema(models.SendTasks{})
	db.Table(taskTable).
		Select("created_by, name, count(*) as cnt").
		Group("created_by, name").
		Having("count(*) > 1").
		Scan(&taskDups)
	if len(taskDups) > 0 {
		t.Errorf("Found unexpected duplicates in %s: %+v", taskTable, taskDups)
	}

	// 3. 检查消息模板表
	var tmplDups []DupCount
	tmplTable := models.GetSchema(models.Template{})
	db.Table(tmplTable).
		Select("created_by, name, count(*) as cnt").
		Group("created_by, name").
		Having("count(*) > 1").
		Scan(&tmplDups)
	if len(tmplDups) > 0 {
		t.Errorf("Found unexpected duplicates in %s: %+v", tmplTable, tmplDups)
	}
}
