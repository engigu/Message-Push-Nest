package models

import (
	"fmt"
	"message-nest/pkg/util"
)

type HostedMessage struct {
	ID      int    `gorm:"primaryKey" json:"id" `
	Title   string `json:"title" gorm:"type:text ;"`
	Content string `json:"content" gorm:"type:text ;"`
	Type    string `json:"type" gorm:"type:varchar(100) ;default:'';index"`

	CreatedAt util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime "`
	UpdatedAt util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime ;"`
}

// Add 添加托管消息
func (message *HostedMessage) Add() error {
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

// DeleteOutDateHostedMessages 删除过期的托管消息，保留最新的 keepNum 条
func DeleteOutDateHostedMessages(keepNum int) (int, error) {
	var affectedRows int
	
	// 优化方案：使用GORM的Offset和Limit找到临界ID，兼容多种数据库
	// 1. 获取第 keepNum 条记录的ID作为临界值
	var threshold HostedMessage
	result := db.Model(&HostedMessage{}).
		Select("id").
		Order("created_on DESC").
		Offset(keepNum - 1).
		Limit(1).
		First(&threshold)
	
	// 如果记录总数不足keepNum条，则不需要删除
	if result.Error != nil {
		return 0, nil
	}
	
	// 2. 删除ID小于临界值的记录
	deleteResult := db.Where("id < ?", threshold.ID).Delete(&HostedMessage{})
	if deleteResult.Error != nil {
		return affectedRows, deleteResult.Error
	}
	
	affectedRows = int(deleteResult.RowsAffected)
	return affectedRows, nil
}
