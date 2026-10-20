package auth_service

import (
	"message-nest/models"
	"testing"
)

func setupTestDB() {
	models.InitTestDB("user_service_test")
}

func TestUserService(t *testing.T) {
	setupTestDB()

	// 0. 初始化系统内置 admin 账号
	adminUser, err := models.AddUser("admin", "admin123", "admin", "system")
	if err != nil {
		t.Fatalf("setup admin failed: %v", err)
	}

	// 1. 测试添加新用户
	svc := UserService{
		Username:  "bob",
		Password:  "bob123456",
		Email:     "bob@example.com",
		Role:      "user",
		CreatedBy: adminUser.ID,
	}

	createdBob, err := svc.AddUser()
	if err != nil {
		t.Fatalf("AddUser failed: %v", err)
	}
	if createdBob.ID == "" || createdBob.Role != "user" {
		t.Fatalf("unexpected bob data: %+v", createdBob)
	}

	// 2. 测试重复用户名添加应失败
	duplicateSvc := UserService{
		Username: "bob",
		Password: "password",
	}
	_, err = duplicateSvc.AddUser()
	if err == nil {
		t.Fatalf("expected error when adding duplicate username, got nil")
	}

	// 3. 测试编辑用户
	editSvc := UserService{
		ID:         createdBob.ID,
		Email:      "bob_new@example.com",
		Role:       "user",
		Status:     0,
		ModifiedBy: adminUser.ID,
	}
	err = editSvc.EditUser()
	if err != nil {
		t.Fatalf("EditUser failed: %v", err)
	}

	updatedBob, _ := models.GetUserByID(createdBob.ID)
	if updatedBob.Email != "bob_new@example.com" || updatedBob.Status != 0 {
		t.Fatalf("EditUser did not update fields correctly: %+v", updatedBob)
	}

	// 4. 测试非法角色应被拒绝
	invalidRoleSvc := UserService{
		ID:     createdBob.ID,
		Role:   "super_hacker",
		Status: 1,
	}
	err = invalidRoleSvc.EditUser()
	if err == nil {
		t.Fatalf("expected error for invalid role 'super_hacker', got nil")
	}

	// 5. 测试重置密码
	resetSvc := UserService{
		ID:         createdBob.ID,
		Password:   "newSecretPasswd",
		ModifiedBy: adminUser.ID,
	}
	err = resetSvc.ResetPassword()
	if err != nil {
		t.Fatalf("ResetPassword failed: %v", err)
	}

	ok, _ := models.CheckAuth("bob", "newSecretPasswd")
	if !ok {
		// 注意 bob 状态当前是 0 (禁用)，CheckAuth 要求 status == 1
		// 先恢复状态为 1
		_ = models.EditUserByID(createdBob.ID, map[string]interface{}{"status": 1})
		ok, _ = models.CheckAuth("bob", "newSecretPasswd")
		if !ok {
			t.Fatalf("CheckAuth with new password failed")
		}
	}

	// 6. 测试删除安全校验
	// 6.1 不能删除自身
	deleteSelfSvc := UserService{ID: adminUser.ID}
	err = deleteSelfSvc.DeleteUser(adminUser.ID)
	if err == nil || err.Error() != "不能删除当前登录的管理员账号自身" {
		t.Fatalf("expected cannot delete self error, got: %v", err)
	}

	// 6.2 不能删除系统内置 admin
	deleteAdminSvc := UserService{ID: adminUser.ID}
	err = deleteAdminSvc.DeleteUser("another_admin_id")
	if err == nil || err.Error() != "不能删除系统内置 admin 账号" {
		t.Fatalf("expected cannot delete admin account error, got: %v", err)
	}

	// 6.3 正常删除普通用户
	deleteBobSvc := UserService{ID: createdBob.ID}
	err = deleteBobSvc.DeleteUser(adminUser.ID)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	deletedCheck, _ := models.GetUserByID(createdBob.ID)
	if deletedCheck != nil {
		t.Fatalf("expected user to be deleted, but still found")
	}

	// 7. 测试集成登录（OIDC/OAuth2）用户来源类型标记识别
	oidcUser, err := models.AddSSOUser("sso_alice", "secret", "alice@sso.com", "sub_123456", "oidc")
	if err != nil {
		t.Fatalf("AddSSOUser oidc failed: %v", err)
	}
	oauth2User, err := models.AddSSOUser("sso_github", "secret", "gh@github.com", "gh_99999", "oauth2")
	if err != nil {
		t.Fatalf("AddSSOUser oauth2 failed: %v", err)
	}

	listSvc := UserService{PageNum: 0, PageSize: 10}
	userList, _, err := listSvc.GetUserList()
	if err != nil {
		t.Fatalf("GetUserList failed: %v", err)
	}

	var foundOidc, foundOAuth2, foundLocal bool
	for _, u := range userList {
		if u.ID == oidcUser.ID {
			foundOidc = true
			if !u.IsOidc || u.LoginType != "sso" || u.AuthProvider != "oidc" || u.OidcSubject != "sub_123456" {
				t.Fatalf("expected sso_alice to have IsOidc=true, AuthProvider='oidc', got %+v", u)
			}
		}
		if u.ID == oauth2User.ID {
			foundOAuth2 = true
			if !u.IsOidc || u.LoginType != "sso" || u.AuthProvider != "oauth2" || u.OidcSubject != "gh_99999" {
				t.Fatalf("expected sso_github to have IsOidc=true, AuthProvider='oauth2', got %+v", u)
			}
		}
		if u.ID == adminUser.ID {
			foundLocal = true
			if u.IsOidc || u.LoginType != "local" || u.AuthProvider != "" {
				t.Fatalf("expected admin to have IsOidc=false, LoginType='local', AuthProvider='', got %+v", u)
			}
		}
	}
	if !foundOidc || !foundOAuth2 || !foundLocal {
		t.Fatalf("did not find expected users in list: oidc=%v, oauth2=%v, local=%v", foundOidc, foundOAuth2, foundLocal)
	}

	// 8. 测试自主注册功能
	// 8.1 默认未开启注册，注册应被拒绝
	regSvc := UserService{
		Username: "newbie",
		Password: "password123",
		Email:    "newbie@example.com",
	}
	_, err = regSvc.Register()
	if err == nil {
		t.Fatalf("expected error when registration is disabled, got nil")
	}

	// 8.2 开启注册开关
	_ = models.AddOneSetting(models.Settings{
		Section: "site_config",
		Key:     "allow_registration",
		Value:   "true",
	})
	registeredUser, err := regSvc.Register()
	if err != nil {
		t.Fatalf("Register failed when enabled: %v", err)
	}
	if registeredUser.Role != "user" || registeredUser.CreatedBy != "self_register" {
		t.Fatalf("unexpected registered user properties: %+v", registeredUser)
	}

	// 8.3 重复注册同一用户名应被拒绝
	_, err = regSvc.Register()
	if err == nil {
		t.Fatalf("expected error for duplicate registration, got nil")
	}

	// 9. 测试管理员及用户 SSO 绑定信息获取与解除绑定
	// 9.1 查询绑定信息
	ssoProfileSvc := UserService{ID: oidcUser.ID}
	profile, err := ssoProfileSvc.GetSSOProfile()
	if err != nil {
		t.Fatalf("GetSSOProfile failed: %v", err)
	}
	if !profile.IsOidc || profile.OidcSubject != "sub_123456" {
		t.Fatalf("unexpected sso profile: %+v", profile)
	}

	// 9.2 解除绑定
	err = ssoProfileSvc.UnbindSSO("")
	if err != nil {
		t.Fatalf("UnbindSSO failed: %v", err)
	}

	unboundProfile, err := ssoProfileSvc.GetSSOProfile()
	if err != nil {
		t.Fatalf("GetSSOProfile after unbind failed: %v", err)
	}
	if unboundProfile.IsOidc || unboundProfile.OidcSubject != "" || unboundProfile.AuthProvider != "" {
		t.Fatalf("expected sso to be cleared after unbind, got %+v", unboundProfile)
	}

	// 10. 测试“一人绑定多个 SSO 身份源”
	// 10.1 为同一个 adminUser 同时绑定 OIDC 和 OAuth2 (如企业统一认证 + GitHub)
	err = models.BindUserSSO(adminUser.ID, "oidc", "corp_sub_admin_001", "admin@corp.com")
	if err != nil {
		t.Fatalf("BindUserSSO oidc failed: %v", err)
	}
	err = models.BindUserSSO(adminUser.ID, "oauth2", "github_sub_admin_999", "admin@github.com")
	if err != nil {
		t.Fatalf("BindUserSSO oauth2 failed: %v", err)
	}

	// 10.2 查询该用户的绑定列表，必须包含两个 Provider
	adminBindings, err := models.GetUserSSOIdentities(adminUser.ID)
	if err != nil {
		t.Fatalf("GetUserSSOIdentities failed: %v", err)
	}
	if len(adminBindings) != 2 {
		t.Fatalf("expected admin to have 2 sso bindings, got %d", len(adminBindings))
	}

	// 10.3 验证使用任一 Provider 均可精准识别出该管理员
	foundByOidc, err := models.GetUserBySSO("oidc", "corp_sub_admin_001")
	if err != nil || foundByOidc.ID != adminUser.ID {
		t.Fatalf("GetUserBySSO with oidc failed: user=%+v, err=%v", foundByOidc, err)
	}
	foundByOAuth2, err := models.GetUserBySSO("oauth2", "github_sub_admin_999")
	if err != nil || foundByOAuth2.ID != adminUser.ID {
		t.Fatalf("GetUserBySSO with oauth2 failed: user=%+v, err=%v", foundByOAuth2, err)
	}

	// 10.4 验证已被绑定的身份不能被他人重复绑定
	err = models.BindUserSSO("other_user_id", "oidc", "corp_sub_admin_001", "fake@corp.com")
	if err == nil {
		t.Fatalf("expected error when another user tries to bind an existing subject, got nil")
	}

	// 10.5 测试单独解绑 OIDC，OAuth2 仍然保持有效
	err = models.UnbindUserSSOByProvider(adminUser.ID, "oidc")
	if err != nil {
		t.Fatalf("UnbindUserSSOByProvider oidc failed: %v", err)
	}
	afterUnbindOidc, _ := models.GetUserSSOIdentities(adminUser.ID)
	if len(afterUnbindOidc) != 1 || afterUnbindOidc[0].Provider != "oauth2" {
		t.Fatalf("expected only oauth2 to remain after unbinding oidc, got %+v", afterUnbindOidc)
	}
	// OAuth2 依然能登录
	foundAfter, err := models.GetUserBySSO("oauth2", "github_sub_admin_999")
	if err != nil || foundAfter.ID != adminUser.ID {
		t.Fatalf("GetUserBySSO after single unbind failed: %+v", foundAfter)
	}
}
