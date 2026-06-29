package models

// No imports needed for this file as it uses types from the same package and builtin errors.


type Settings struct {
	IDModel

	Section string `json:"section" gorm:"type:varchar(100) ;default:'';index"`
	Key     string `json:"key" gorm:"type:varchar(100) ;default:'';"`
	Value   string `json:"value" gorm:"type:text ;"`
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
	if err := db.Model(&Settings{}).Where("id = ? ", id).Updates(data).Error; err != nil {
		return err
	}

	return nil
}

// DeleteMsgTaskIns 删除一条设置
func DeleteSettingByKey(section string, key string) error {
	if err := db.Where(&Settings{Section: section, Key: key}).Delete(&Settings{}).Error; err != nil {
		return err
	}
	return nil
}

func GetSettingByKey(section string, key string) (Settings, error) {
	var setting Settings
	err := db.Where(&Settings{Section: section, Key: key}).Limit(1).Find(&setting).Error
	if err != nil {
		return setting, err
	}
	return setting, nil
}

func GetSettingBySection(section string) ([]Settings, error) {
	var settings []Settings
	err := db.Table(GetSchema(Settings{})).Where(&Settings{Section: section}).Scan(&settings).Error
	if err != nil {
		return settings, err
	}
	return settings, nil
}
