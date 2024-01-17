package models

import (
	"errors"
	"github.com/jinzhu/gorm"
)

type Settings struct {
	IDModel

	Section string `json:"section" gorm:"type:varchar(100) comment '实例类型';default:'';index:section"`
	Key     string `json:"key" gorm:"type:varchar(100) comment '实例类型';default:'';"`
	Value   string `json:"value" gorm:"type:text comment '实例类型';default:'';"`
}

// AddOneSetting 添加一条设置
func AddOneSetting(setting Settings) error {
	if err := db.Create(&setting).Error; err != nil {
		return err
	}
	return nil
}

// EditSetting 编辑设置
func EditSetting(id int, data interface{}) error {
	if err := db.Model(&Settings{}).Where("id = ? ", id).Updates(data).Error; err != nil {
		return err
	}

	return nil
}

// DeleteMsgTaskIns 删除一条设置
func DeleteSettingByKey(section string, key string) error {
	if err := db.Where("`section` = ? and `key` = ? ", section, key).Delete(&Settings{}).Error; err != nil {
		return err
	}
	return nil
}

func GetSettingByKey(section string, key string) (Settings, error) {
	var setting Settings
	err := db.Where("`section` = ? and `key` = ? ", section, key).Find(&setting).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return setting, err
	}
	return setting, nil
}

func GetSettingBySection(section string) ([]Settings, error) {
	var settings []Settings
	err := db.Table(db.NewScope(Settings{}).TableName()).Where("`section` = ? ", section).Scan(&settings).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return settings, err
	}
	return settings, nil
}
