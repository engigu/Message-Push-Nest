package models

import (
	"fmt"
	"gorm.io/gorm"
	"message-nest/pkg/util"
)

type HostedMessage struct {
	ID        int    `gorm:"primaryKey" json:"id"`
	Title     string `json:"title" gorm:"type:text;"`
	Content   string `json:"content" gorm:"type:text;"`
	Type      string `json:"type" gorm:"type:varchar(100);default:'';index"`
	UniqueKey string `json:"unique_key" gorm:"type:varchar(50);uniqueIndex;default:''"`

	CreatedBy  string `json:"created_by" gorm:"type:varchar(36);default:'';index"`
	ModifiedBy string `json:"modified_by" gorm:"type:varchar(36);default:'';"`

	CreatedAt util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedAt util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime;"`
}

// Add 添加托管消息
func (message *HostedMessage) Add() error {
	if message.UniqueKey == "" {
		message.UniqueKey = util.GenerateRandomString(16)
	}
	if err := db.Create(&message).Error; err != nil {
		return err
	}
	return nil
}

// 托管消息的结果
type HostMessageResult struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Type       string    `json:"type"`
	UniqueKey  string    `json:"unique_key"`
	CreatedBy  string    `json:"created_by"`
	ModifiedBy string    `json:"modified_by"`
	CreatedOn  util.Time `json:"created_on"`
	ModifiedOn util.Time `json:"modified_on"`
}

// GetHostMessages 获取所有托管消息记录
func GetHostMessages(pageNum int, pageSize int, text string, maps map[string]interface{}) ([]HostMessageResult, error) {
	var datas []HostMessageResult
	hostMessageT := GetSchema(HostedMessage{})

	query := db.Table(hostMessageT)

	//dayVal, ok := maps["day_created_on"]
	//if ok {
	//	delete(maps, "day_created_on")
	//	query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
	//}

	query = query.Where(maps)
	if text != "" {
		query = query.Where("title like ? or content like ?", fmt.Sprintf("%%%s%%", text), fmt.Sprintf("%%%s%%", text))
	}
	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	query.Scan(&datas)

	return datas, nil
}

// GetHostMessagesTotal 获取托管消息总数
func GetHostMessagesTotal(text string, maps map[string]interface{}) (int64, error) {
	var total int64
	hostMessageT := GetSchema(HostedMessage{})

	query := db.Table(hostMessageT)

	//dayVal, ok := maps["day_created_on"]
	//if ok {
	//	delete(maps, "day_created_on")
	//	query = query.Where(fmt.Sprintf("DATE(%s.created_on) = ?", logt), dayVal)
	//}

	query = query.Where(maps)
	if text != "" {
		query = query.Where("title like ?  or content like ?", fmt.Sprintf("%%%s%%", text), fmt.Sprintf("%%%s%%", text))
	}
	query.Count(&total)
	return total, nil
}

// DeleteOutDateHostedMessages 删除过期的托管消息，保留最新的 keepNum 条，支持按用户过滤
func DeleteOutDateHostedMessages(keepNum int, createdBy ...string) (int, error) {
	var affectedRows int
	if keepNum <= 0 {
		return 0, nil
	}

	userID := ""
	if len(createdBy) > 0 {
		userID = createdBy[0]
	}
	
	// 优化方案：使用GORM的Offset和Limit找到临界ID，兼容多种数据库
	// 1. 获取第 keepNum 条记录的ID作为临界值
	var threshold HostedMessage
	query := db.Model(&HostedMessage{}).Select("id")
	if userID != "" {
		isAdmin := false
		if u, err := GetUserByID(userID); err == nil && u != nil && u.Role == "admin" {
			isAdmin = true
		}
		if isAdmin {
			query = query.Where("created_by = ? OR created_by = '' OR created_by IS NULL", userID)
		} else {
			query = query.Where("created_by = ?", userID)
		}
	} else if len(createdBy) > 0 && createdBy[0] == "" {
		query = query.Where("created_by = '' OR created_by IS NULL")
	}

	result := query.Order("created_on DESC, id DESC").
		Offset(keepNum - 1).
		Limit(1).
		Find(&threshold)
	
	// 如果记录总数不足keepNum条，或者查找失败，则不需要删除
	if result.Error != nil || threshold.ID == 0 {
		return 0, nil
	}
	
	// 2. 删除ID小于临界值的记录（限定用户范围）
	deleteQuery := db.Where("id < ?", threshold.ID)
	if userID != "" {
		isAdmin := false
		if u, err := GetUserByID(userID); err == nil && u != nil && u.Role == "admin" {
			isAdmin = true
		}
		if isAdmin {
			deleteQuery = deleteQuery.Where("created_by = ? OR created_by = '' OR created_by IS NULL", userID)
		} else {
			deleteQuery = deleteQuery.Where("created_by = ?", userID)
		}
	} else if len(createdBy) > 0 && createdBy[0] == "" {
		deleteQuery = deleteQuery.Where("created_by = '' OR created_by IS NULL")
	}

	deleteResult := deleteQuery.Delete(&HostedMessage{})
	if deleteResult.Error != nil {
		return affectedRows, deleteResult.Error
	}
	
	affectedRows = int(deleteResult.RowsAffected)
	return affectedRows, nil
}

// GetHostMessageByUniqueKey 根据 unique_key 获取托管消息
func GetHostMessageByUniqueKey(uniqueKey string) (*HostMessageResult, error) {
	var data HostMessageResult
	hostMessageT := GetSchema(HostedMessage{})
	err := db.Table(hostMessageT).Where("unique_key = ?", uniqueKey).Limit(1).Find(&data).Error
	if err != nil {
		return nil, err
	}
	if data.UniqueKey == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return &data, nil
}
