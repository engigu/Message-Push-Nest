package models

import "github.com/jinzhu/gorm"

type Auth struct {
	ID       int    `gorm:"primary_key" json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// CheckAuth 检查用户信息
func CheckAuth(username, password string) (bool, error) {
	var auth Auth
	err := db.Select("id").Where(Auth{Username: username, Password: password}).First(&auth).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	if auth.ID > 0 {
		return true, nil
	}

	return false, nil
}

// EditUser 编辑用户信息
func EditUser(username string, data interface{}) error {
	if err := db.Model(&Auth{}).Where("username = ? ", username).Updates(data).Error; err != nil {
		return err
	}
	return nil
}
