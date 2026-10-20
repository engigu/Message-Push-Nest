package auth_service

import (
	"errors"
	"fmt"
	"strings"

	"message-nest/models"
	"message-nest/service/settings_service"
)

type UserService struct {
	ID         string
	Username   string
	Password   string
	Email      string
	Role       string
	Status     int
	Search     string
	CreatedBy  string
	ModifiedBy string
	PageNum    int
	PageSize   int
}

// GetUserList 获取用户列表
func (s *UserService) GetUserList() ([]models.UserResult, int64, error) {
	maps := make(map[string]interface{})
	total, err := models.GetUsersTotal(s.Search, maps)
	if err != nil {
		return nil, 0, err
	}
	users, err := models.GetUsers(s.PageNum, s.PageSize, s.Search, maps)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// Register 访客自主注册
func (s *UserService) Register() (*models.Auth, error) {
	// 1. 检查注册功能是否开启
	if !settings_service.IsRegistrationAllowed() {
		return nil, errors.New("系统当前暂未开放自主注册，请联系管理员开通账号")
	}

	username := strings.TrimSpace(s.Username)
	if username == "" {
		return nil, errors.New("用户名不能为空")
	}
	if len(username) < 2 || len(username) > 30 {
		return nil, errors.New("用户名长度需在 2 到 30 个字符之间")
	}
	if len(s.Password) < 4 || len(s.Password) > 50 {
		return nil, errors.New("密码长度需在 4 到 50 个字符之间")
	}

	// 2. 检查用户名唯一性与格式
	if available, reason := models.CheckUsernameAvailable(username); !available {
		return nil, errors.New(reason)
	}

	// 3. 强制分配角色为普通用户 (user)，禁止提权，标记来源为 self_register
	user, err := models.AddUser(username, s.Password, "user", "self_register")
	if err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	if s.Email != "" {
		_ = models.EditUserByID(user.ID, map[string]interface{}{"email": s.Email})
		user.Email = s.Email
	}

	return user, nil
}

// AddUser 添加新用户
func (s *UserService) AddUser() (*models.Auth, error) {
	username := strings.TrimSpace(s.Username)
	if s.Password == "" {
		return nil, errors.New("密码不能为空")
	}

	// 检查用户名唯一性与格式
	if available, reason := models.CheckUsernameAvailable(username); !available {
		return nil, errors.New(reason)
	}

	role := s.Role
	if role != "admin" && role != "user" {
		role = "user"
	}

	user, err := models.AddUser(s.Username, s.Password, role, s.CreatedBy)
	if err != nil {
		return nil, err
	}
	if s.Email != "" {
		_ = models.EditUserByID(user.ID, map[string]interface{}{"email": s.Email})
		user.Email = s.Email
	}
	return user, nil
}

// EditUser 编辑用户
func (s *UserService) EditUser() error {
	if s.ID == "" {
		return errors.New("用户ID不能为空")
	}

	exist, err := models.GetUserByID(s.ID)
	if err != nil || exist == nil {
		return errors.New("用户不存在")
	}

	data := make(map[string]interface{})
	if s.Email != "" {
		data["email"] = s.Email
	}
	if s.Role != "" {
		if s.Role != "admin" && s.Role != "user" {
			return errors.New("角色仅支持 admin 或 user")
		}
		data["role"] = s.Role
	}
	data["status"] = s.Status
	data["modified_by"] = s.ModifiedBy

	return models.EditUserByID(s.ID, data)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(currentUserID string) error {
	if s.ID == "" {
		return errors.New("用户ID不能为空")
	}
	if s.ID == currentUserID {
		return errors.New("不能删除当前登录的管理员账号自身")
	}

	target, err := models.GetUserByID(s.ID)
	if err != nil || target == nil {
		return errors.New("用户不存在")
	}
	if target.Username == "admin" {
		return errors.New("不能删除系统内置 admin 账号")
	}

	return models.DeleteUser(s.ID)
}

// ResetPassword 重置用户密码
func (s *UserService) ResetPassword() error {
	if s.ID == "" {
		return errors.New("用户ID不能为空")
	}
	if s.Password == "" {
		return errors.New("新密码不能为空")
	}

	data := map[string]interface{}{
		"password":    s.Password,
		"modified_by": s.ModifiedBy,
	}
	return models.EditUserByID(s.ID, data)
}

// SSOBindingDTO 单个第三方身份源绑定详情
type SSOBindingDTO struct {
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Subject   string `json:"subject"`
	Email     string `json:"email"`
	CreatedOn string `json:"created_on"`
}

// SSOProfileDTO 用户 SSO 关联信息 (包含所有绑定的第三方身份列表)
type SSOProfileDTO struct {
	ID           string          `json:"id"`
	Username     string          `json:"username"`
	Email        string          `json:"email"`
	Role         string          `json:"role"`
	IsOidc       bool            `json:"is_oidc"`
	AuthProvider string          `json:"auth_provider"`
	OidcSubject  string          `json:"oidc_subject"`
	Bindings     []SSOBindingDTO `json:"bindings"`
}

// GetSSOProfile 获取当前用户的 SSO 绑定信息 (含多身份列表)
func (s *UserService) GetSSOProfile() (*SSOProfileDTO, error) {
	if s.ID == "" {
		return nil, errors.New("用户ID不能为空")
	}
	user, err := models.GetUserByID(s.ID)
	if err != nil || user == nil {
		return nil, errors.New("用户不存在")
	}

	// 从独立表查询多绑定列表
	rawIdentities, _ := models.GetUserSSOIdentities(s.ID)
	bindings := make([]SSOBindingDTO, 0, len(rawIdentities))
	for _, ident := range rawIdentities {
		bindings = append(bindings, SSOBindingDTO{
			ID:        ident.ID,
			Provider:  ident.Provider,
			Subject:   ident.Subject,
			Email:     ident.Email,
			CreatedOn: ident.CreatedAt.String(),
		})
	}

	hasBinding := len(bindings) > 0 || user.OidcSubject != "" || user.AuthProvider != ""

	return &SSOProfileDTO{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Role:         user.Role,
		IsOidc:       hasBinding,
		AuthProvider: user.AuthProvider,
		OidcSubject:  user.OidcSubject,
		Bindings:     bindings,
	}, nil
}

// UnbindSSO 解除指定用户的 SSO 绑定 (支持指定特定 Provider，若 provider 为空则解绑所有)
func (s *UserService) UnbindSSO(provider string) error {
	if s.ID == "" {
		return errors.New("用户ID不能为空")
	}
	user, err := models.GetUserByID(s.ID)
	if err != nil || user == nil {
		return errors.New("用户不存在")
	}
	return models.UnbindUserSSOByProvider(s.ID, provider)
}
