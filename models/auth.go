package models

import (
	"errors"
	"fmt"
	"strings"

	"message-nest/pkg/util"

	"gorm.io/gorm"
)

type Auth struct {
	ID           string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	Username     string    `json:"username" gorm:"type:varchar(100);default:'';uniqueIndex"`
	Password     string    `json:"password" gorm:"type:varchar(100);default:'';"`
	Email        string    `json:"email" gorm:"type:varchar(255);default:'';"`
	Role         string    `json:"role" gorm:"type:varchar(50);default:'user';"` // admin / user
	Status       int       `json:"status" gorm:"type:int;default:1;"`            // 1: 启用, 0: 禁用
	OidcSubject  string    `json:"oidc_subject" gorm:"type:varchar(255);default:'';index"`
	AuthProvider string    `json:"auth_provider" gorm:"type:varchar(50);default:'';index"` // oidc / oauth2 等
	CreatedBy    string    `json:"created_by" gorm:"type:varchar(36);default:'';"`
	ModifiedBy   string    `json:"modified_by" gorm:"type:varchar(36);default:'';"`
	CreatedAt    util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedAt    util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime"`
}

// GenerateUserUniqueID 生成用户唯一随机字符串ID (如 UR + 10位随机字符)
func GenerateUserUniqueID() string {
	newUUID := util.GenerateUniqueID()
	return fmt.Sprintf("UR%s", newUUID)
}

// CheckAuth 检查用户信息
func CheckAuth(username, password string) (bool, error) {
	var auth Auth
	err := db.Select("id, status").Where("username = ? AND password = ?", username, password).Limit(1).Find(&auth).Error
	if err != nil {
		return false, err
	}

	if auth.ID != "" && auth.Status == 1 {
		return true, nil
	}

	return false, nil
}

// GetUserByUsername 根据用户名获取用户 (大小写不敏感)
func GetUserByUsername(username string) (*Auth, error) {
	var auth Auth
	name := strings.TrimSpace(username)
	err := db.Where("LOWER(username) = LOWER(?)", name).Limit(1).Find(&auth).Error
	if err != nil {
		return nil, err
	}
	if auth.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &auth, nil
}

// CheckUsernameAvailable 检查用户名是否可用 (大小写无关排重及保留标识拦截)
func CheckUsernameAvailable(username string, excludeUserID ...string) (bool, string) {
	name := strings.TrimSpace(username)
	if name == "" {
		return false, "用户名不能为空"
	}
	if len(name) < 2 || len(name) > 30 {
		return false, "用户名长度须在 2 到 30 个字符之间"
	}
	reserved := []string{"system", "anonymous", "root"}
	for _, r := range reserved {
		if strings.EqualFold(name, r) {
			return false, fmt.Sprintf("用户名 [%s] 为系统保留标识，不可使用", name)
		}
	}

	query := db.Model(&Auth{}).Where("LOWER(username) = LOWER(?)", name)
	if len(excludeUserID) > 0 && excludeUserID[0] != "" {
		query = query.Where("id != ?", excludeUserID[0])
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, "校验用户名失败: " + err.Error()
	}
	if count > 0 {
		return false, fmt.Sprintf("用户名 [%s] 已被使用，请更换", name)
	}
	return true, ""
}

// GetUserByID 根据用户ID获取用户
func GetUserByID(id string) (*Auth, error) {
	var auth Auth
	err := db.Where("id = ?", id).Limit(1).Find(&auth).Error
	if err != nil {
		return nil, err
	}
	if auth.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &auth, nil
}

// EditUser 编辑用户信息
func EditUser(username string, data map[string]interface{}) error {
	if err := db.Model(&Auth{}).Where("username = ?", username).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

// EditUserByID 根据ID编辑用户
func EditUserByID(id string, data map[string]interface{}) error {
	if err := db.Model(&Auth{}).Where("id = ?", id).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

// AddUser 添加新用户
func AddUser(account string, password string, role string, createdBy string) (*Auth, error) {
	if role == "" {
		role = "user"
	}
	cleanAccount := strings.TrimSpace(account)
	auth := Auth{
		ID:         GenerateUserUniqueID(),
		Username:   cleanAccount,
		Password:   password,
		Role:       role,
		Status:     1,
		CreatedBy:  createdBy,
		ModifiedBy: createdBy,
	}
	if err := db.Create(&auth).Error; err != nil {
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "Duplicate") {
			return nil, fmt.Errorf("用户名 [%s] 已存在，请更换", cleanAccount)
		}
		return nil, err
	}
	return &auth, nil
}

// GetUserByOidcSubject 根据 OIDC Subject 获取用户
func GetUserByOidcSubject(subject string) (*Auth, error) {
	var auth Auth
	err := db.Where("oidc_subject = ?", subject).Limit(1).Find(&auth).Error
	if err != nil {
		return nil, err
	}
	if auth.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &auth, nil
}

// GetUserByEmail 根据邮箱获取未绑定 OIDC 的用户
func GetUserByEmail(email string) (*Auth, error) {
	var auth Auth
	err := db.Where("email = ? AND (oidc_subject = '' OR oidc_subject IS NULL)", email).Limit(1).Find(&auth).Error
	if err != nil {
		return nil, err
	}
	if auth.ID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &auth, nil
}

// UpdateUserOidcSubject 更新用户的 OIDC Subject
func UpdateUserOidcSubject(id string, subject string) error {
	return db.Model(&Auth{}).Where("id = ?", id).Update("oidc_subject", subject).Error
}

// UpdateUserSSO 更新用户的 SSO 绑定信息 (Subject 与 AuthProvider)
func UpdateUserSSO(id string, subject string, provider string) error {
	data := map[string]interface{}{
		"oidc_subject":  subject,
		"auth_provider": provider,
	}
	return db.Model(&Auth{}).Where("id = ?", id).Updates(data).Error
}

// UnbindUserSSO 解除用户的 SSO 绑定
func UnbindUserSSO(id string) error {
	return db.Model(&Auth{}).Where("id = ?", id).Updates(map[string]interface{}{
		"oidc_subject":  "",
		"auth_provider": "",
	}).Error
}

// AddOidcUser 添加 OIDC 用户（默认 provider 为 oidc）
func AddOidcUser(username, password, email, oidcSubject string) (*Auth, error) {
	return AddSSOUser(username, password, email, oidcSubject, "oidc")
}

// AddSSOUser 添加指定 SSO 类型的用户
func AddSSOUser(username, password, email, oidcSubject, authProvider string) (*Auth, error) {
	if authProvider == "" {
		authProvider = "oidc"
	}
	auth := Auth{
		ID:           GenerateUserUniqueID(),
		Username:     username,
		Password:     password,
		Email:        email,
		Role:         "user",
		Status:       1,
		OidcSubject:  oidcSubject,
		AuthProvider: authProvider,
		CreatedBy:    "sso",
		ModifiedBy:   "sso",
	}
	if err := db.Create(&auth).Error; err != nil {
		return nil, err
	}
	if oidcSubject != "" {
		_ = db.Create(&UserSSOIdentity{
			ID:       GenerateSSOIdentityID(),
			UserID:   auth.ID,
			Provider: authProvider,
			Subject:  oidcSubject,
			Email:    email,
		}).Error
	}
	return &auth, nil
}

// UserResult 用户返回结构体（隐藏敏感字段）
type UserResult struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	Status       int       `json:"status"`
	OidcSubject  string    `json:"oidc_subject"`
	AuthProvider string    `json:"auth_provider"` // "oidc", "oauth2" 等，本地用户为空
	IsOidc       bool      `json:"is_oidc"`
	LoginType    string    `json:"login_type"` // "sso": 集成登录, "local": 本地账号
	CreatedBy    string    `json:"created_by"`
	ModifiedBy   string    `json:"modified_by"`
	CreatedOn    util.Time `json:"created_on"`
	ModifiedOn   util.Time `json:"modified_on"`
}

// GetUsers 获取用户列表
func GetUsers(pageNum int, pageSize int, search string, maps interface{}) ([]UserResult, error) {
	var users []UserResult
	query := db.Model(&Auth{}).Where(maps)
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	err := query.Select("id, username, email, role, status, oidc_subject, auth_provider, created_by, modified_by, created_on, modified_on").Scan(&users).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	for i := range users {
		if users[i].OidcSubject != "" || users[i].CreatedBy == "sso" || users[i].AuthProvider != "" {
			users[i].IsOidc = true
			users[i].LoginType = "sso"
			// 存量数据平滑兜底：若没有明确记录 provider，默认记为 oidc
			if users[i].AuthProvider == "" {
				users[i].AuthProvider = "oidc"
			}
		} else {
			users[i].IsOidc = false
			users[i].LoginType = "local"
			users[i].AuthProvider = ""
		}
	}
	return users, nil
}

// GetUsersTotal 获取用户总数
func GetUsersTotal(search string, maps interface{}) (int64, error) {
	var total int64
	query := db.Model(&Auth{}).Where(maps)
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}
	err := query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

// DeleteUser 删除用户
func DeleteUser(id string) error {
	return db.Where("id = ?", id).Delete(&Auth{}).Error
}
