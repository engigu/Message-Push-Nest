package models

import (
	"testing"
)

func TestSettingsSectionQueryAndParameterization(t *testing.T) {
	InitTestDB("settings_test")

	// 1. 插入管理员与普通用户的同 section 设置
	adminID := "UR_ADMIN_S01"
	userID := "UR_USER_S01"

	sAdmin := Settings{
		Section:    "test_sec",
		Key:        "timeout",
		Value:      "30",
		CreatedBy:  adminID,
		ModifiedBy: adminID,
	}
	sUser := Settings{
		Section:    "test_sec",
		Key:        "timeout",
		Value:      "60",
		CreatedBy:  userID,
		ModifiedBy: userID,
	}

	if err := AddOneSetting(sAdmin); err != nil {
		t.Fatalf("failed to add admin setting: %v", err)
	}
	if err := AddOneSetting(sUser); err != nil {
		t.Fatalf("failed to add user setting: %v", err)
	}

	// 2. 查询该 section 下配置，带入 userID（触发参数化 Order 子句）
	settings, err := GetSettingBySection("test_sec", userID)
	if err != nil {
		t.Fatalf("GetSettingBySection failed: %v", err)
	}
	if len(settings) == 0 {
		t.Fatalf("expected non-empty settings")
	}

	// 3. 测试通过 Key 查找配置
	foundSetting, err := GetSettingByKey("test_sec", "timeout", userID)
	if err != nil {
		t.Fatalf("GetSettingByKey failed: %v", err)
	}
	if foundSetting.Value != "60" {
		t.Fatalf("expected user private value '60', got: %s", foundSetting.Value)
	}
}

