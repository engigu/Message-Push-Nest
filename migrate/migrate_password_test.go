package migrate

import (
	"message-nest/models"
	"testing"
)

func TestMigrateUserPasswordsToBcrypt(t *testing.T) {
	db := models.InitTestDB("migrate_password_test")

	// 1. 模拟存量数据库：插入多个明文密码用户及一个已加盐用户
	alreadyBcrypt, err := models.HashPassword("ModernSecPass789")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	legacyUsers := []models.Auth{
		{
			ID:       "UR_MIG_001",
			Username: "user_plain_1",
			Password: "PlainPassOne123",
			Role:     "user",
			Status:   1,
		},
		{
			ID:       "UR_MIG_002",
			Username: "user_plain_2",
			Password: "PlainPassTwo456",
			Role:     "user",
			Status:   1,
		},
		{
			ID:       "UR_MIG_003",
			Username: "user_bcrypt_3",
			Password: alreadyBcrypt,
			Role:     "user",
			Status:   1,
		},
	}

	for _, u := range legacyUsers {
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("failed to seed user %s: %v", u.Username, err)
		}
	}

	// 2. 确保初始时没有升级完成标记
	settingO, _ := models.GetSettingByKey("init", "user_password_bcrypt_migration")
	if settingO.Value == "1" {
		t.Fatalf("expected migration setting to not exist yet")
	}

	// 3. 执行迁移逻辑
	MigrateUserPasswordsToBcrypt()

	// 4. 验证 setting 标记已写入且值为 "1"
	settingAfter, err := models.GetSettingByKey("init", "user_password_bcrypt_migration")
	if err != nil || settingAfter.Value != "1" {
		t.Fatalf("expected user_password_bcrypt_migration setting to be '1', got: %v (err: %v)", settingAfter.Value, err)
	}

	// 5. 验证原明文密码用户均已升级为 bcrypt，且能通过明文比对
	user1, err := models.GetUserByID("UR_MIG_001")
	if err != nil {
		t.Fatalf("failed to query user1: %v", err)
	}
	if !models.IsBcryptHash(user1.Password) {
		t.Fatalf("expected user1 password to be bcrypt hash, got: %s", user1.Password)
	}
	if !models.CheckPassword("PlainPassOne123", user1.Password) {
		t.Fatalf("user1 password verification failed against original plaintext")
	}

	user2, err := models.GetUserByID("UR_MIG_002")
	if err != nil {
		t.Fatalf("failed to query user2: %v", err)
	}
	if !models.IsBcryptHash(user2.Password) {
		t.Fatalf("expected user2 password to be bcrypt hash, got: %s", user2.Password)
	}
	if !models.CheckPassword("PlainPassTwo456", user2.Password) {
		t.Fatalf("user2 password verification failed against original plaintext")
	}

	// 6. 验证原本就是 bcrypt 的用户未被破坏
	user3, err := models.GetUserByID("UR_MIG_003")
	if err != nil {
		t.Fatalf("failed to query user3: %v", err)
	}
	if user3.Password != alreadyBcrypt {
		t.Fatalf("user3 password was unexpectedly modified: %s vs %s", user3.Password, alreadyBcrypt)
	}

	// 7. 幂等性测试：再次调用应直接跳过
	MigrateUserPasswordsToBcrypt()
}

