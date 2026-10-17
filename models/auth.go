package models

import (
	"gorm.io/gorm"
)

type Auth struct {
	ID          int    `json:"id" gorm:"autoIncrement;type:integer ;primaryKey"`
	Username    string `json:"username" gorm:"type:varchar(100);default:'';"`
	Password    string `json:"password" gorm:"type:varchar(100);default:'';"`
	Email       string `json:"email" gorm:"type:varchar(255);default:'';"`
	OidcSubject string `json:"oidc_subject" gorm:"type:varchar(255);default:'';index"`
}

// CheckAuth 检查用户信息
func CheckAuth(username, password string) (bool, error) {
	var auth Auth
	err := db.Select("id").Where(Auth{Username: username, Password: password}).Limit(1).Find(&auth).Error
	if err != nil {
		return false, err
	}

	if auth.ID > 0 {
		return true, nil
	}

	return false, nil
}

func GetUserByUsername(username string) (*Auth, error) {
    var auth Auth
    err := db.Where("username = ?", username).Limit(1).Find(&auth).Error
    if err != nil {
        return nil, err
    }
	if auth.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
    return &auth, nil
}

// EditUser 编辑用户信息
func EditUser(username string, data map[string]interface{}) error {
	if err := db.Model(&Auth{}).Where("username = ? ", username).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

func AddUser(account string, password string) error {
	auth := Auth{
		Username: account,
		Password: password,
	}
	if err := db.Create(&auth).Error; err != nil {
		return err
	}
	return nil
}

func GetUserByOidcSubject(subject string) (*Auth, error) {
	var auth Auth
	err := db.Where("oidc_subject = ?", subject).Limit(1).Find(&auth).Error
	if err != nil {
		return nil, err
	}
	if auth.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &auth, nil
}

func GetUserByEmail(email string) (*Auth, error) {
	var auth Auth
	err := db.Where("email = ? AND (oidc_subject = '' OR oidc_subject IS NULL)", email).Limit(1).Find(&auth).Error
	if err != nil {
		return nil, err
	}
	if auth.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &auth, nil
}

func UpdateUserOidcSubject(id int, subject string) error {
	return db.Model(&Auth{}).Where("id = ?", id).Update("oidc_subject", subject).Error
}

func AddOidcUser(username, password, email, oidcSubject string) (*Auth, error) {
	auth := Auth{
		Username:    username,
		Password:    password,
		Email:       email,
		OidcSubject: oidcSubject,
	}
	if err := db.Create(&auth).Error; err != nil {
		return nil, err
	}
	return &auth, nil
}
