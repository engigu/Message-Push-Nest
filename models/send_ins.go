package models

type SendTasksIns struct {
	UUIDModel

	TaskID      string `json:"task_id"`
	WayID       string `json:"way_id"`
	WayType     string `json:"way_type"`
	ContentType string `json:"content_type"`
	Config      string `json:"config"`
	Extra       string `json:"extra"`
}

// InsEmailConfig 实例里面的邮箱config
type InsEmailConfig struct {
	ToAccount string `json:"to_account" validate:"required,email" label:"收件邮箱"`
	Title     string `json:"title" validate:"required,max=150" label:"邮箱标题"`
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
