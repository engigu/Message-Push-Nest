package models

type SendTasksIns struct {
	UUIDModel

	TaskID      string `json:"task_id"  gorm:"type:varchar(12) ;default:'';index"`
	WayID       string `json:"way_id" gorm:"type:varchar(12) ;default:'';index"`
	WayType     string `json:"way_type" gorm:"type:varchar(100) ;default:'';index"`
	ContentType string `json:"content_type" gorm:"type:varchar(100) ;default:'';index"`
	Config      string `json:"config" gorm:"type:text ;"`
	Extra       string `json:"extra" gorm:"type:text ;"`
	Enable      int    `json:"enable" gorm:"type:int ;default:1;"`
}

// InsEmailConfig 实例里面的邮箱config
type InsEmailConfig struct {
	ToAccount string `json:"to_account" validate:"required,email" label:"收件邮箱"`
}

// InsWeChatAccountConfig 实例里面的邮箱config
type InsWeChatAccountConfig struct {
	ToAccount string `json:"to_account" validate:"required" label:"收件微信Openid"`
}

// InsEmailConfig 实例里面的邮箱config
type InsDtalkConfig struct {
}

// InsQyWeiXinConfig 实例里面的企业微信config
type InsQyWeiXinConfig struct {
}

// InsCustomConfig 实例里面的自定义config
type InsCustomConfig struct {
}

// InsMessageNestConfig 实例里面的托管消息config
type InsMessageNestConfig struct {
}

// ManyAddTaskIns 批量添加实例
func ManyAddTaskIns(taskIns []SendTasksIns) error {
	tx := db.Begin()
	for _, ins := range taskIns {
		// 存在就跳过这条ins记录
		err := db.Where("id = ?", ins.ID).Take(&SendTasksIns{}).Error
		if err == nil {
			continue
		}
		if err := tx.Create(&ins).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

// AddTaskInsOne 添加一条实例
func AddTaskInsOne(ins SendTasksIns) error {
	if err := db.Create(&ins).Error; err != nil {
		return err
	}
	return nil
}

// DeleteMsgTaskIns 删除一条实例
func DeleteMsgTaskIns(id string) error {
	if err := db.Where("id = ?", id).Delete(&SendTasksIns{}).Error; err != nil {
		return err
	}
	return nil
}

// UpdateMsgTaskIns 更新实例
func UpdateMsgTaskIns(id string, data map[string]interface{}) error {
	if err := db.Model(&SendTasksIns{}).Where("id = ?", id).Updates(data).Error; err != nil {
		return err
	}
	return nil
}
