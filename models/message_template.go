package models

import (
	"fmt"
	"message-nest/pkg/util"
)

// GenerateTemplateUniqueID 生成模板唯一ID
func GenerateTemplateUniqueID() string {
	newUUID := util.GenerateUniqueID()
	return fmt.Sprintf("TP%s", newUUID)
}

// MessageTemplate 消息模板
type MessageTemplate struct {
	UUIDModel
	
	Name        string `json:"name" gorm:"type:varchar(200);not null;index" binding:"required"`
	Description string `json:"description" gorm:"type:text"`
	
	// 模板内容（带占位符）
	TextTemplate     string `json:"text_template" gorm:"type:text"`
	HTMLTemplate     string `json:"html_template" gorm:"type:text"`
	MarkdownTemplate string `json:"markdown_template" gorm:"type:text"`
	
	// 占位符定义（JSON格式）
	Placeholders string `json:"placeholders" gorm:"type:text"`
	
	// @提醒配置
	AtMobiles string `json:"at_mobiles" gorm:"type:text;comment:'@手机号列表，逗号分隔'"`
	AtUserIds string `json:"at_user_ids" gorm:"type:text;comment:'@用户ID列表，逗号分隔'"`
	IsAtAll   bool   `json:"is_at_all" gorm:"default:false;comment:'是否@所有人'"`
	
	// 状态：enabled/disabled
	Status string `json:"status" gorm:"type:varchar(20);default:'enabled';index"`
}

// Add 添加消息模板
func (t *MessageTemplate) Add() error {
	if err := db.Create(&t).Error; err != nil {
		return err
	}
	return nil
}

// Update 更新消息模板
func (t *MessageTemplate) Update() error {
	if err := db.Model(&MessageTemplate{}).Where("id = ?", t.ID).Updates(t).Error; err != nil {
		return err
	}
	return nil
}

// Delete 删除消息模板
func (t *MessageTemplate) Delete() error {
	if err := db.Where("id = ?", t.ID).Delete(&MessageTemplate{}).Error; err != nil {
		return err
	}
	return nil
}

// MessageTemplateResult 消息模板查询结果
type MessageTemplateResult struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	TextTemplate     string    `json:"text_template"`
	HTMLTemplate     string    `json:"html_template"`
	MarkdownTemplate string    `json:"markdown_template"`
	Placeholders     string    `json:"placeholders"`
	AtMobiles        string    `json:"at_mobiles"`
	AtUserIds        string    `json:"at_user_ids"`
	IsAtAll          bool      `json:"is_at_all"`
	Status           string    `json:"status"`
	CreatedBy        string    `json:"created_by"`
	ModifiedBy       string    `json:"modified_by"`
	CreatedOn        util.Time `json:"created_on"`
	ModifiedOn       util.Time `json:"modified_on"`
}

// GetMessageTemplates 获取消息模板列表
func GetMessageTemplates(pageNum int, pageSize int, text string, maps map[string]interface{}) ([]MessageTemplateResult, error) {
	var datas []MessageTemplateResult
	templateT := GetSchema(MessageTemplate{})
	
	query := db.Table(templateT)
	query = query.Where(maps)
	
	if text != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", 
			fmt.Sprintf("%%%s%%", text), 
			fmt.Sprintf("%%%s%%", text))
	}
	
	query = query.Order("created_on DESC")
	
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	
	query.Scan(&datas)
	return datas, nil
}

// GetMessageTemplatesTotal 获取消息模板总数
func GetMessageTemplatesTotal(text string, maps map[string]interface{}) (int64, error) {
	var total int64
	templateT := GetSchema(MessageTemplate{})
	
	query := db.Table(templateT)
	query = query.Where(maps)
	
	if text != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", 
			fmt.Sprintf("%%%s%%", text), 
			fmt.Sprintf("%%%s%%", text))
	}
	
	query.Count(&total)
	return total, nil
}

// GetMessageTemplateByID 根据ID获取消息模板
func GetMessageTemplateByID(id string) (*MessageTemplateResult, error) {
	var data MessageTemplateResult
	templateT := GetSchema(MessageTemplate{})
	
	err := db.Table(templateT).Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}
	
	return &data, nil
}

// ExistMessageTemplateByID 检查模板是否存在
func ExistMessageTemplateByID(id string) (bool, error) {
	var template MessageTemplate
	err := db.Select("id").Where("id = ?", id).First(&template).Error
	if err != nil {
		return false, err
	}
	
	if template.ID != "" {
		return true, nil
	}
	
	return false, nil
}
