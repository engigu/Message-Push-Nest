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

// Template 消息模板
type Template struct {
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
func (t *Template) Add() error {
	if err := db.Create(&t).Error; err != nil {
		return err
	}
	return nil
}

// Update 更新消息模板
func (t *Template) Update() error {
	// 使用 Select 明确指定要更新的字段，包括布尔值字段，排除不应更新的时间戳字段
	if err := db.Model(&Template{}).Where("id = ?", t.ID).
		Select("name", "description", "text_template", "html_template", "markdown_template", 
			"placeholders", "at_mobiles", "at_user_ids", "is_at_all", "status", "modified_by").
		Updates(t).Error; err != nil {
		return err
	}
	return nil
}

// Delete 删除消息模板
func (t *Template) Delete() error {
	if err := db.Where("id = ?", t.ID).Delete(&Template{}).Error; err != nil {
		return err
	}
	return nil
}

// TemplateResult 消息模板查询结果
type TemplateResult struct {
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

// GetTemplates 获取消息模板列表
func GetTemplates(pageNum int, pageSize int, text string, maps map[string]interface{}) ([]TemplateResult, error) {
	var datas []TemplateResult
	templateT := GetSchema(Template{})
	
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

// GetTemplatesTotal 获取消息模板总数
func GetTemplatesTotal(text string, maps map[string]interface{}) (int64, error) {
	var total int64
	templateT := GetSchema(Template{})
	
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

// GetTemplateByID 根据ID获取消息模板
func GetTemplateByID(id string) (*TemplateResult, error) {
	var data TemplateResult
	templateT := GetSchema(Template{})
	
	err := db.Table(templateT).Where("id = ?", id).First(&data).Error
	if err != nil {
		return nil, err
	}
	
	return &data, nil
}

// ExistTemplateByID 检查模板是否存在
func ExistTemplateByID(id string) (bool, error) {
	var template Template
	err := db.Select("id").Where("id = ?", id).First(&template).Error
	if err != nil {
		return false, err
	}
	
	if template.ID != "" {
		return true, nil
	}
	
	return false, nil
}
