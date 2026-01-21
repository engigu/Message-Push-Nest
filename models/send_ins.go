package models

import "fmt"

type SendTasksIns struct {
	UUIDModel

	TaskID      string `json:"task_id"  gorm:"type:varchar(12) ;default:'';index"`
	TemplateID  string `json:"template_id"  gorm:"type:varchar(12) ;default:'';index"` // 模板ID
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

// InsFeishuConfig 实例里面的飞书config
type InsFeishuConfig struct {
}

// InsCustomConfig 实例里面的自定义config
type InsCustomConfig struct {
}

// InsMessageNestConfig 实例里面的托管消息config
type InsMessageNestConfig struct {
}

// InsAliyunSMSConfig 实例里面的阿里云短信config
type InsAliyunSMSConfig struct {
	PhoneNumber  string `json:"phone_number" validate:"required" label:"手机号码"`
	TemplateCode string `json:"template_code" validate:"required" label:"短信模板CODE"`
}

// InsTelegramConfig 实例里面的Telegram config
type InsTelegramConfig struct {
}

// InsBarkConfig 实例里面的Bark config
type InsBarkConfig struct {
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

// GetTemplateInsList 获取模板关联的实例列表（包含渠道名称）
func GetTemplateInsList(templateID string) ([]SendTasksInsRes, error) {
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})
	var insList []SendTasksInsRes

	err := db.
		Table(insTable).
		Select(fmt.Sprintf("%s.*, %s.name as way_name", insTable, waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Where(fmt.Sprintf("%s.template_id = ?", insTable), templateID).
		Order(fmt.Sprintf("%s.created_on DESC", insTable)).
		Scan(&insList).Error

	if err != nil {
		return nil, err
	}
	return insList, nil
}
