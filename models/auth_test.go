package models

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	return InitTestDB("auth_test")
}

func TestGenerateUserUniqueID(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateUserUniqueID()

		// 检查前缀
		if !strings.HasPrefix(id, "UR") {
			t.Fatalf("expected prefix 'UR', got %s", id)
		}

		// 检查总长度为 12
		if len(id) != 12 {
			t.Fatalf("expected length 12, got %d (%s)", len(id), id)
		}

		// 检查无重复
		if seen[id] {
			t.Fatalf("collision detected on iteration %d: %s", i, id)
		}
		seen[id] = true
	}
}

func TestAuthModelCRUD(t *testing.T) {
	setupTestDB()

	// 1. 添加用户
	username := "test_user_01"
	password := "Secret123"
	role := "admin"
	creator := "system"

	user, err := AddUser(username, password, role, creator)
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	// 验证 ID 格式
	if !strings.HasPrefix(user.ID, "UR") || len(user.ID) != 12 {
		t.Fatalf("unexpected user ID format: %s", user.ID)
	}

	// 验证密码字段已设置
	if user.Password != password {
		t.Fatalf("expected password to match %s, got %s", password, user.Password)
	}

	// 2. 密码校验
	ok, err := CheckAuth(username, password)
	if err != nil || !ok {
		t.Fatalf("CheckAuth failed for correct credentials: %v", err)
	}
	ok, err = CheckAuth(username, "WrongPassword")
	if err != nil || ok {
		t.Fatalf("CheckAuth should return false for incorrect password")
	}

	// 3. 通过 ID 查询
	fetchedUser, err := GetUserByID(user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if fetchedUser.Username != username || fetchedUser.Role != role {
		t.Fatalf("unexpected user data: %+v", fetchedUser)
	}

	// 4. 通过 Username 查询
	fetchedByUsername, err := GetUserByUsername(username)
	if err != nil {
		t.Fatalf("GetUserByUsername failed: %v", err)
	}
	if fetchedByUsername.ID != user.ID {
		t.Fatalf("expected user ID %s, got %s", user.ID, fetchedByUsername.ID)
	}

	// 5. 编辑用户
	newEmail := "user01@example.com"
	err = EditUserByID(user.ID, map[string]interface{}{
		"email":       newEmail,
		"role":        "user",
		"status":      0,
		"modified_by": "admin",
	})
	if err != nil {
		t.Fatalf("EditUserByID failed: %v", err)
	}

	updatedUser, _ := GetUserByID(user.ID)
	if updatedUser.Email != newEmail || updatedUser.Role != "user" || updatedUser.Status != 0 {
		t.Fatalf("user was not updated correctly: %+v", updatedUser)
	}

	// 6. 用户列表与总数
	total, err := GetUsersTotal("test_user", nil)
	if err != nil || total != 1 {
		t.Fatalf("expected total 1, got %d, err: %v", total, err)
	}

	users, err := GetUsers(0, 10, "test_user", nil)
	if err != nil || len(users) != 1 {
		t.Fatalf("expected 1 user in list, got %d, err: %v", len(users), err)
	}
	if users[0].ID != user.ID {
		t.Fatalf("expected user in list to be %s, got %s", user.ID, users[0].ID)
	}

	// 7. 删除用户
	err = DeleteUser(user.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	_, err = GetUserByID(user.ID)
	if err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestCheckUsernameAvailable(t *testing.T) {
	InitTestDB("auth_check_username_test")

	// 1. 保留字测试
	avail, reason := CheckUsernameAvailable("system")
	if avail || !strings.Contains(reason, "保留") {
		t.Fatalf("expected 'system' to be unavailable as reserved, got avail=%v, reason=%s", avail, reason)
	}
	avail, _ = CheckUsernameAvailable("ROOT")
	if avail {
		t.Fatalf("expected 'ROOT' to be unavailable as reserved")
	}

	// 2. 长度边界测试
	avail, _ = CheckUsernameAvailable("a")
	if avail {
		t.Fatalf("expected 1-char username to be unavailable")
	}
	avail, _ = CheckUsernameAvailable("   ")
	if avail {
		t.Fatalf("expected empty username to be unavailable")
	}

	// 3. 不存在时返回可用
	avail, reason = CheckUsernameAvailable("valid_username")
	if !avail {
		t.Fatalf("expected 'valid_username' to be available, got reason: %s", reason)
	}

	// 4. 插入用户并测试大小写不敏感排重
	created, err := AddUser("Charlie", "password", "user", "system")
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}

	// 4.1 完全同名
	avail, _ = CheckUsernameAvailable("Charlie")
	if avail {
		t.Fatalf("expected 'Charlie' to be unavailable")
	}

	// 4.2 全小写
	avail, _ = CheckUsernameAvailable("charlie")
	if avail {
		t.Fatalf("expected 'charlie' to be unavailable due to case-insensitivity")
	}

	// 4.3 全大写
	avail, _ = CheckUsernameAvailable("CHARLIE")
	if avail {
		t.Fatalf("expected 'CHARLIE' to be unavailable due to case-insensitivity")
	}

	// 4.4 排除自身 ID 时可用
	avail, _ = CheckUsernameAvailable("Charlie", created.ID)
	if !avail {
		t.Fatalf("expected 'Charlie' to be available when excluding own ID")
	}

	// 5. 验证底层并发/直接 AddUser 重复时的友好错误提示
	_, err = AddUser("Charlie", "pass", "user", "system")
	if err == nil || !strings.Contains(err.Error(), "已存在") {
		t.Fatalf("expected friendly duplicate error from AddUser, got %v", err)
	}
}

