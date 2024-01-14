package models

type SendTasksIns struct {
	UUIDModel

	TaskID      string `json:"task_id"  gorm:"type:varchar(36) comment '任务id';default:'';index:task_id"`
	WayID       string `json:"way_id" gorm:"type:varchar(36) comment '渠道id';default:'';index:way_id"`
	WayType     string `json:"way_type" gorm:"type:varchar(100) comment '渠道类型';default:'';index:way_type"`
	ContentType string `json:"content_type" gorm:"type:varchar(100) comment '实例类型';default:'';index:content_type"`
	Config      string `json:"config" gorm:"type:text comment '实例配置';"`
	Extra       string `json:"extra" gorm:"type:text comment '额外信息';"`
	Enable      int    `json:"enable" gorm:"type:int comment '开启、暂停状态';default:1;"`
}

// InsEmailConfig 实例里面的邮箱config
type InsEmailConfig struct {
	ToAccount string `json:"to_account" validate:"required,email" label:"收件邮箱"`
	//Title     string `json:"title" validate:"required,max=150" label:"邮箱标题"`
}

// InsEmailConfig 实例里面的邮箱config
type InsDtalkConfig struct {
}

// InsCustomConfig 实例里面的自定义config
type InsCustomConfig struct {
}

// ManyAddTaskIns 批量添加实例
func ManyAddTaskIns(taskIns []SendTasksIns) error {
	tx := db.Begin()
	for _, ins := range taskIns {
		// 存在就跳过这条ins记录
		err := db.Where("id = ?", ins.ID).Find(&SendTasksIns{}).Error
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
