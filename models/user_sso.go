package models

import (
	"errors"
	"fmt"
	"message-nest/pkg/util"

	"gorm.io/gorm"
)

// UserSSOIdentity 用户与第三方单点登录身份绑定关系模型 (支持 1:N 多身份源绑定)
type UserSSOIdentity struct {
	ID        string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:varchar(36);not null;index:idx_user_provider,unique;index:idx_user_id"`
	Provider  string    `json:"provider" gorm:"type:varchar(50);not null;index:idx_provider_subject,unique;index:idx_user_provider,unique"`
	Subject   string    `json:"subject" gorm:"type:varchar(255);not null;index:idx_provider_subject,unique"`
	Email     string    `json:"email" gorm:"type:varchar(255);default:'';"`
	ExtraData string    `json:"extra_data" gorm:"type:text;"`
	CreatedAt util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedAt util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime"`
}

// GenerateSSOIdentityID 生成唯一标识
func GenerateSSOIdentityID() string {
	return fmt.Sprintf("ID%s", util.GenerateUniqueID())
}

// GetUserBySSO 根据 Provider 和 Subject 查询绑定的系统账号 (优先独立表，平滑兜底存量 Auth 表)
func GetUserBySSO(provider, subject string) (*Auth, error) {
	if subject == "" {
		return nil, errors.New("subject 标识不能为空")
	}

	// 1. 查询独立关联表
	var identity UserSSOIdentity
	query := db.Where("subject = ?", subject)
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	err := query.Limit(1).Find(&identity).Error
	if err == nil && identity.ID != "" && identity.UserID != "" {
		user, uErr := GetUserByID(identity.UserID)
		if uErr == nil && user != nil && user.ID != "" {
			return user, nil
		}
	}

	// 2. 兜底查询历史 Auth 表
	user, err := GetUserByOidcSubject(subject)
	if err == nil && user != nil && user.ID != "" {
		// 静默回填至新表
		p := provider
		if p == "" {
			p = user.AuthProvider
			if p == "" {
				p = "oidc"
			}
		}
		_ = db.Create(&UserSSOIdentity{
			ID:       GenerateSSOIdentityID(),
			UserID:   user.ID,
			Provider: p,
			Subject:  subject,
			Email:    user.Email,
		}).Error
		return user, nil
	}

	// 兼容带前缀格式 (如 oidc:sub_123)
	if provider != "" {
		qualified := fmt.Sprintf("%s:%s", provider, subject)
		user, err = GetUserByOidcSubject(qualified)
		if err == nil && user != nil && user.ID != "" {
			_ = db.Create(&UserSSOIdentity{
				ID:       GenerateSSOIdentityID(),
				UserID:   user.ID,
				Provider: provider,
				Subject:  subject,
				Email:    user.Email,
			}).Error
			return user, nil
		}
	}

	return nil, gorm.ErrRecordNotFound
}

// GetUserSSOIdentities 获取指定用户绑定的所有第三方单点登录身份列表
func GetUserSSOIdentities(userID string) ([]UserSSOIdentity, error) {
	if userID == "" {
		return nil, errors.New("用户ID不能为空")
	}

	var identities []UserSSOIdentity
	err := db.Where("user_id = ?", userID).Order("created_on ASC").Find(&identities).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 如果关联表为空，检查历史 Auth 表是否存在绑定，若有则进行平滑迁移
	if len(identities) == 0 {
		user, uErr := GetUserByID(userID)
		if uErr == nil && user != nil && user.OidcSubject != "" {
			p := user.AuthProvider
			if p == "" {
				p = "oidc"
			}
			newIdentity := UserSSOIdentity{
				ID:       GenerateSSOIdentityID(),
				UserID:   user.ID,
				Provider: p,
				Subject:  user.OidcSubject,
				Email:    user.Email,
			}
			_ = db.Create(&newIdentity).Error
			identities = append(identities, newIdentity)
		}
	}

	return identities, nil
}

// BindUserSSO 为指定系统账号绑定某个提供商的 SSO 身份 (支持一人多绑定)
func BindUserSSO(userID, provider, subject, email string) error {
	if userID == "" || provider == "" || subject == "" {
		return errors.New("参数不完整 (userID, provider, subject 均不能为空)")
	}

	// 1. 检查该 (provider, subject) 是否已被其他用户绑定
	var existOther UserSSOIdentity
	err := db.Where("provider = ? AND subject = ? AND user_id != ?", provider, subject, userID).Limit(1).Find(&existOther).Error
	if err == nil && existOther.ID != "" {
		otherUser, _ := GetUserByID(existOther.UserID)
		name := existOther.UserID
		if otherUser != nil {
			name = otherUser.Username
		}
		return fmt.Errorf("该单点登录身份已绑定至账号 [%s]，无法重复绑定", name)
	}

	// 2. 检查该用户在该 provider 下是否已有绑定记录
	var myIdentity UserSSOIdentity
	err = db.Where("user_id = ? AND provider = ?", userID, provider).Limit(1).Find(&myIdentity).Error
	if err == nil && myIdentity.ID != "" {
		// 更新现有绑定
		data := map[string]interface{}{
			"subject": subject,
		}
		if email != "" {
			data["email"] = email
		}
		if err := db.Model(&UserSSOIdentity{}).Where("id = ?", myIdentity.ID).Updates(data).Error; err != nil {
			return err
		}
	} else {
		// 新建绑定
		newIdentity := UserSSOIdentity{
			ID:       GenerateSSOIdentityID(),
			UserID:   userID,
			Provider: provider,
			Subject:  subject,
			Email:    email,
		}
		if err := db.Create(&newIdentity).Error; err != nil {
			return err
		}
	}

	// 3. 同步回写 Auth 表主字段 (兼容旧版查询与徽章显示)
	var auth Auth
	if err := db.Where("id = ?", userID).Limit(1).Find(&auth).Error; err == nil && auth.ID != "" {
		if auth.OidcSubject == "" || auth.AuthProvider == "" || auth.AuthProvider == provider {
			_ = UpdateUserSSO(userID, subject, provider)
		}
		if auth.Email == "" && email != "" {
			_ = EditUserByID(userID, map[string]interface{}{"email": email})
		}
	}

	return nil
}

// UnbindUserSSOByProvider 解除指定系统账号在某个提供商（或全部）的 SSO 绑定
func UnbindUserSSOByProvider(userID, provider string) error {
	if userID == "" {
		return errors.New("用户ID不能为空")
	}

	// 1. 删除关联表中的记录
	query := db.Where("user_id = ?", userID)
	if provider != "" {
		query = query.Where("provider = ?", provider)
	}
	if err := query.Delete(&UserSSOIdentity{}).Error; err != nil {
		return err
	}

	// 2. 同步更新 Auth 表
	var remaining []UserSSOIdentity
	_ = db.Where("user_id = ?", userID).Order("created_on ASC").Find(&remaining).Error
	if len(remaining) > 0 {
		// 切换主绑定为剩余的其中一个
		_ = UpdateUserSSO(userID, remaining[0].Subject, remaining[0].Provider)
	} else {
		// 全部已解绑，清空 Auth 表
		_ = UnbindUserSSO(userID)
	}

	return nil
}
