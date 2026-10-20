package models

import (
	"context"
	"message-nest/pkg/app"
	"testing"
)

// MockItem 用于测试审计基类和 Hook
type MockItem struct {
	UUIDModel
	Title string `json:"title" gorm:"type:varchar(100)"`
}

func (MockItem) TableName() string {
	return "test_mock_items"
}

func TestAuditModelHookAndTenantScope(t *testing.T) {
	InitTestDB("audit_test")
	_ = db.AutoMigrate(&MockItem{})

	userA := app.UserContextInfo{UserID: "UR_USER_AAAA", Username: "userA", Role: "user"}
	userB := app.UserContextInfo{UserID: "UR_USER_BBBB", Username: "userB", Role: "user"}
	admin := app.UserContextInfo{UserID: "UR_ADMIN_000", Username: "admin", Role: "admin"}

	ctxA := app.WithUserContext(context.Background(), userA)
	ctxB := app.WithUserContext(context.Background(), userB)
	ctxAdmin := app.WithUserContext(context.Background(), admin)

	// 1. 测试 BeforeCreate 自动注入 CreatedBy 和 ModifiedBy
	itemA := MockItem{
		UUIDModel: UUIDModel{ID: "ITEM_001"},
		Title:     "Item by A",
	}
	err := db.WithContext(ctxA).Create(&itemA).Error
	if err != nil {
		t.Fatalf("Create itemA failed: %v", err)
	}

	if itemA.CreatedBy != userA.UserID {
		t.Fatalf("Expected CreatedBy = %s, got %s", userA.UserID, itemA.CreatedBy)
	}
	if itemA.ModifiedBy != userA.UserID {
		t.Fatalf("Expected ModifiedBy = %s, got %s", userA.UserID, itemA.ModifiedBy)
	}

	// 2. 测试 BeforeUpdate 自动注入 ModifiedBy
	// 由管理员更新该记录
	err = db.WithContext(ctxAdmin).Model(&MockItem{}).Where("id = ?", itemA.ID).Updates(MockItem{
		Title: "Item by A (updated by admin)",
	}).Error
	if err != nil {
		t.Fatalf("Update itemA by admin failed: %v", err)
	}

	var updatedItem MockItem
	_ = db.First(&updatedItem, "id = ?", itemA.ID)
	if updatedItem.CreatedBy != userA.UserID {
		t.Fatalf("CreatedBy should remain %s, got %s", userA.UserID, updatedItem.CreatedBy)
	}
	if updatedItem.ModifiedBy != admin.UserID {
		t.Fatalf("Expected ModifiedBy = %s, got %s", admin.UserID, updatedItem.ModifiedBy)
	}

	// 3. 测试 TenantScope 自动租户数据过滤
	var listForB []MockItem
	err = db.WithContext(ctxB).Scopes(TenantScope(ctxB)).Find(&listForB).Error
	if err != nil {
		t.Fatalf("Query list for B failed: %v", err)
	}
	if len(listForB) != 0 {
		t.Fatalf("Tenant isolation failed: User B can see User A's item")
	}

	var listForAdmin []MockItem
	err = db.WithContext(ctxAdmin).Scopes(TenantScope(ctxAdmin)).Find(&listForAdmin).Error
	if err != nil {
		t.Fatalf("Query list for Admin failed: %v", err)
	}
	if len(listForAdmin) != 1 {
		t.Fatalf("Admin should see 1 item, got %d", len(listForAdmin))
	}

	// 4. 测试 CheckOwnership 防越权守卫
	err = CheckOwnership(ctxB, "test_mock_items", itemA.ID)
	if err == nil {
		t.Fatalf("Tenant security failure: User B passed ownership check on User A's item")
	}

	err = CheckOwnership(ctxA, "test_mock_items", itemA.ID)
	if err != nil {
		t.Fatalf("User A should pass ownership check: %v", err)
	}

	err = CheckOwnership(ctxAdmin, "test_mock_items", itemA.ID)
	if err != nil {
		t.Fatalf("Admin should pass ownership check: %v", err)
	}

	// 5. 测试管理员穿越 Context（Impersonation）
	// 5.1 管理员穿越 User B 视角并代客创建资产
	adminImpersonateB := app.UserContextInfo{
		UserID:            admin.UserID,
		Username:          admin.Username,
		Role:              admin.Role,
		ImpersonateUserID: userB.UserID, // 穿越为 User B
	}
	ctxAdminImpersonateB := app.WithUserContext(context.Background(), adminImpersonateB)

	itemBByAdmin := MockItem{
		UUIDModel: UUIDModel{ID: "ITEM_002"},
		Title:     "Item created by admin for B",
	}
	err = db.WithContext(ctxAdminImpersonateB).Create(&itemBByAdmin).Error
	if err != nil {
		t.Fatalf("Admin impersonating B create failed: %v", err)
	}

	// 资产所有权属于被代建的 User B，审计操作人必须追溯至真实管理员
	if itemBByAdmin.CreatedBy != userB.UserID {
		t.Fatalf("Impersonate asset ownership failed: Expected CreatedBy = %s, got %s", userB.UserID, itemBByAdmin.CreatedBy)
	}
	if itemBByAdmin.ModifiedBy != admin.UserID {
		t.Fatalf("Impersonate audit traceability failed: Expected ModifiedBy = %s, got %s", admin.UserID, itemBByAdmin.ModifiedBy)
	}

	// 5.2 穿越模式下的列表查询（TenantScope 应精准过滤为目标租户 B 的视角）
	var listUnderBView []MockItem
	err = db.WithContext(ctxAdminImpersonateB).Scopes(TenantScope(ctxAdminImpersonateB)).Find(&listUnderBView).Error
	if err != nil {
		t.Fatalf("Query under impersonate context failed: %v", err)
	}
	if len(listUnderBView) != 1 || listUnderBView[0].ID != "ITEM_002" {
		t.Fatalf("Impersonate TenantScope failed: expected 1 item for B, got %d", len(listUnderBView))
	}

	// 5.3 退出穿越模式后，管理员恢复全局全量视角
	var listUnderGlobalAdmin []MockItem
	err = db.WithContext(ctxAdmin).Scopes(TenantScope(ctxAdmin)).Find(&listUnderGlobalAdmin).Error
	if err != nil {
		t.Fatalf("Query under global admin failed: %v", err)
	}
	if len(listUnderGlobalAdmin) != 2 {
		t.Fatalf("Global admin should see all 2 items, got %d", len(listUnderGlobalAdmin))
	}

	// 5.4 普通用户伪造穿越字段测试（非 Admin 调用 IsImpersonating 必须为 false）
	fakeImpersonateUser := app.UserContextInfo{
		UserID:            userA.UserID,
		Username:          userA.Username,
		Role:              "user",
		ImpersonateUserID: userB.UserID,
	}
	if fakeImpersonateUser.IsImpersonating() {
		t.Fatalf("Security failure: non-admin should NOT be able to impersonate")
	}
	if fakeImpersonateUser.GetEffectiveUserID() != userA.UserID {
		t.Fatalf("Security failure: non-admin effective user id should remain self")
	}
}
