package models

import (
	"fmt"
	"gorm.io/gorm"
	"message-nest/pkg/util"
)

type Settings struct {
	ID uint `gorm:"autoIncrement;type:integer;primaryKey" json:"id"`

	Section string `json:"section" gorm:"column:section;type:varchar(100);default:'';index:idx_settings_sec_key;index:idx_settings_lookup,priority:1"`
	Key     string `json:"key" gorm:"column:key;type:varchar(100);default:'';index:idx_settings_sec_key;index:idx_settings_lookup,priority:2"`
	Value   string `json:"value" gorm:"column:value;type:text;"`

	CreatedBy  string `json:"created_by" gorm:"column:created_by;type:varchar(36);default:'';index:idx_settings_created_by;index:idx_settings_lookup,priority:3"`
	ModifiedBy string `json:"modified_by" gorm:"column:modified_by;type:varchar(36);default:'';"`

	CreatedAt util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedAt util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime;"`
}

// AddOneSetting 添加一条设置
func AddOneSetting(setting Settings) error {
	if err := db.Create(&setting).Error; err != nil {
		return err
	}
	return nil
}

// EditSetting 编辑设置
func EditSetting(id uint, data interface{}) error {
	if err := db.Model(&Settings{}).Where("id = ?", id).Updates(data).Error; err != nil {
		return err
	}
	return nil
}

// DeleteSettingByKey 删除一条设置
func DeleteSettingByKey(section string, key string, createdBy ...string) error {
	query := db.Where(&Settings{Section: section, Key: key})
	if len(createdBy) > 0 && createdBy[0] != "" {
		query = query.Where("created_by = ?", createdBy[0])
	}
	if err := query.Delete(&Settings{}).Error; err != nil {
		return err
	}
	return nil
}

// GetSettingByUser 精确查找属于指定用户的某项配置
func GetSettingByUser(section string, key string, createdBy string) (Settings, error) {
	var setting Settings
	err := db.Where(&Settings{Section: section, Key: key, CreatedBy: createdBy}).Limit(1).Find(&setting).Error
	if err != nil {
		return setting, err
	}
	if setting.ID == 0 {
		return setting, gorm.ErrRecordNotFound
	}
	return setting, nil
}

// GetSettingByKey 获取单项配置，支持按用户过滤，查不到时可降级查找全局配置
func GetSettingByKey(section string, key string, createdBy ...string) (Settings, error) {
	var setting Settings
	authTable := GetSchema(Auth{})

	if len(createdBy) > 0 && createdBy[0] != "" {
		err := db.Where(&Settings{Section: section, Key: key, CreatedBy: createdBy[0]}).Limit(1).Find(&setting).Error
		if err == nil && setting.ID > 0 {
			return setting, nil
		}
	}

	// 查找全局配置 (created_by 为空或归属系统管理员)
	err := db.Where(&Settings{Section: section, Key: key}).
		Where(fmt.Sprintf("(created_by = '' OR created_by IS NULL OR created_by IN (SELECT id FROM %s WHERE role = 'admin'))", authTable)).
		Order("CASE WHEN created_by = '' OR created_by IS NULL THEN 0 ELSE 1 END DESC").
		Limit(1).Find(&setting).Error
	if err == nil && setting.ID > 0 {
		return setting, nil
	}

	err = db.Where(&Settings{Section: section, Key: key}).Limit(1).Find(&setting).Error
	if err != nil {
		return setting, err
	}
	if setting.ID == 0 {
		return setting, gorm.ErrRecordNotFound
	}
	return setting, nil
}

// GetSettingBySection 获取某 Section 下的配置列表，支持用户私有配置优先覆盖全局/管理员配置
func GetSettingBySection(section string, createdBy ...string) ([]Settings, error) {
	var settings []Settings
	settingsTable := GetSchema(Settings{})
	authTable := GetSchema(Auth{})

	query := db.Table(settingsTable).Where(&Settings{Section: section})
	if len(createdBy) > 0 && createdBy[0] != "" {
		currentUserID := createdBy[0]
		query = query.Where(fmt.Sprintf(
			"(%s.created_by = ? OR %s.created_by = '' OR %s.created_by IS NULL OR %s.created_by IN (SELECT id FROM %s WHERE role = 'admin'))",
			settingsTable, settingsTable, settingsTable, settingsTable, authTable,
		), currentUserID).
			Order(fmt.Sprintf("CASE WHEN %s.created_by = '%s' THEN 1 ELSE 0 END ASC", settingsTable, currentUserID))
	}
	err := query.Scan(&settings).Error
	if err != nil {
		return settings, err
	}
	return settings, nil
}

// GetConfiguredUsersBySection 获取配置了某个 section 的所有独立用户 ID 列表
func GetConfiguredUsersBySection(section string) ([]string, error) {
	var userIDs []string
	settingsTable := GetSchema(Settings{})
	err := db.Table(settingsTable).
		Where("section = ? AND created_by != '' AND created_by IS NOT NULL", section).
		Distinct("created_by").
		Pluck("created_by", &userIDs).Error
	return userIDs, err
}
