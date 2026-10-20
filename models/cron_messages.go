package models

import (
	"fmt"
	"gorm.io/gorm"
	"message-nest/pkg/util"
)

type CronMessages struct {
	UUIDModel

	Name    string `json:"name" gorm:"type:varchar(200) ;default:'';"`
	TaskID  string `json:"task_id" gorm:"type:varchar(36) ;default:'';"`
	Cron    string `json:"cron" gorm:"type:varchar(4096) ;default:'';"`
	Title   string `json:"title" gorm:"type:varchar(1000) ;default:'';"`
	Content string `json:"content" gorm:"type:varchar(4096) ;default:'';"`
	//MarkDown string `json:"markdown" gorm:"type:varchar(4096) ;default:'';"`
	Url    string `json:"url" gorm:"type:varchar(4096) ;default:'';"`
	Enable int    `json:"enable" gorm:"type:int ;default:1;"`
}

func GenerateMsgUniqueID() string {
	newUUID := util.GenerateUniqueID()
	return fmt.Sprintf("CM%s", newUUID)
}

func AddSendCronMsg(
	name string,
	task_id string,
	cron string,
	title string,
	content string,
	url string,
	createdBy string,
) (string, error) {
	newUUID := GenerateMsgUniqueID()
	msg := CronMessages{
		UUIDModel: UUIDModel{
			ID:         newUUID,
			CreatedBy:  createdBy,
			ModifiedBy: createdBy,
		},
		Name:    name,
		TaskID:  task_id,
		Cron:    cron,
		Title:   title,
		Content: content,
		Url:     url,
		Enable:  1,
	}
	if err := db.Create(&msg).Error; err != nil {
		return newUUID, err
	}
	return newUUID, nil
}

// GetCronMessages 获取所有任务
func GetCronMessages(pageNum int, pageSize int, name string, maps interface{}) ([]CronMessages, error) {
	var (
		msgs []CronMessages
		err  error
	)
	query := db.Where(maps)
	if name != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", name))
	}
	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	err = query.Find(&msgs).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return msgs, nil
}

// GetCronMessagesTotal 获取所有任务总数
func GetCronMessagesTotal(name string, maps interface{}) (int64, error) {
	var (
		err   error
		total int64
	)
	query := db.Model(&CronMessages{}).Where(maps)
	if name != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", name))
	}

	err = query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func DeleteCronMsg(id string, createdBy ...string) error {
	query := db.Where("id = ?", id)
	if len(createdBy) > 0 && createdBy[0] != "" {
		query = query.Where("created_by = ?", createdBy[0])
	}
	result := query.Delete(&CronMessages{})
	if result.Error != nil {
		return result.Error
	}
	if len(createdBy) > 0 && createdBy[0] != "" && result.RowsAffected == 0 {
		return ErrNoPermissionOrNotFound
	}
	return nil
}

func EditCronMsg(id string, data interface{}, createdBy ...string) error {
	query := db.Model(&CronMessages{}).Where("id = ?", id)
	if len(createdBy) > 0 && createdBy[0] != "" {
		query = query.Where("created_by = ?", createdBy[0])
	}
	result := query.Updates(data)
	if result.Error != nil {
		return result.Error
	}
	if len(createdBy) > 0 && createdBy[0] != "" && result.RowsAffected == 0 {
		return ErrNoPermissionOrNotFound
	}
	return nil
}

func GetCronMsgByID(id string) (CronMessages, error) {
	var msg CronMessages
	err := db.Where("id = ? ", id).Limit(1).Find(&msg).Error
	if err != nil {
		return msg, err
	}
	if msg.ID == "" {
		return msg, gorm.ErrRecordNotFound
	}
	return msg, nil
}
