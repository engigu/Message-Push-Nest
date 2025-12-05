package message_template_service

import (
	"encoding/json"
	"errors"
	"message-nest/models"
	"strings"
)

type MessageTemplateService struct {
	ID               string
	Name             string
	Description      string
	TextTemplate     string
	HTMLTemplate     string
	MarkdownTemplate string
	Placeholders     string
	AtMobiles        string
	AtUserIds        string
	IsAtAll          bool
	Status           string
	Text             string
	
	PageNum  int
	PageSize int
}

// Placeholder 占位符定义
type Placeholder struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Default string `json:"default"`
}

// Add 添加消息模板
func (s *MessageTemplateService) Add() error {
	if err := s.validatePlaceholders(); err != nil {
		return err
	}
	
	newUUID := models.GenerateTemplateUniqueID()
	model := models.MessageTemplate{
		UUIDModel: models.UUIDModel{
			ID: newUUID,
		},
		Name:             s.Name,
		Description:      s.Description,
		TextTemplate:     s.TextTemplate,
		HTMLTemplate:     s.HTMLTemplate,
		MarkdownTemplate: s.MarkdownTemplate,
		Placeholders:     s.Placeholders,
		AtMobiles:        s.AtMobiles,
		AtUserIds:        s.AtUserIds,
		IsAtAll:          s.IsAtAll,
		Status:           s.Status,
	}
	
	return model.Add()
}

// Update 更新消息模板
func (s *MessageTemplateService) Update() error {
	if err := s.validatePlaceholders(); err != nil {
		return err
	}
	
	model := models.MessageTemplate{
		UUIDModel: models.UUIDModel{
			ID: s.ID,
		},
		Name:             s.Name,
		Description:      s.Description,
		TextTemplate:     s.TextTemplate,
		HTMLTemplate:     s.HTMLTemplate,
		MarkdownTemplate: s.MarkdownTemplate,
		Placeholders:     s.Placeholders,
		AtMobiles:        s.AtMobiles,
		AtUserIds:        s.AtUserIds,
		IsAtAll:          s.IsAtAll,
		Status:           s.Status,
	}
	
	return model.Update()
}

// Delete 删除消息模板
func (s *MessageTemplateService) Delete() error {
	model := models.MessageTemplate{
		UUIDModel: models.UUIDModel{
			ID: s.ID,
		},
	}
	return model.Delete()
}

// Get 获取单个消息模板
func (s *MessageTemplateService) Get() (*models.MessageTemplateResult, error) {
	return models.GetMessageTemplateByID(s.ID)
}

// GetAll 获取消息模板列表
func (s *MessageTemplateService) GetAll() ([]models.MessageTemplateResult, error) {
	templates, err := models.GetMessageTemplates(s.PageNum, s.PageSize, s.Text, s.getMaps())
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// Count 获取消息模板总数
func (s *MessageTemplateService) Count() (int64, error) {
	return models.GetMessageTemplatesTotal(s.Text, s.getMaps())
}

// ExistByID 检查模板是否存在
func (s *MessageTemplateService) ExistByID() (bool, error) {
	return models.ExistMessageTemplateByID(s.ID)
}

// RenderTemplate 渲染模板（替换占位符）
func (s *MessageTemplateService) RenderTemplate(templateContent string, params map[string]string) string {
	result := templateContent
	
	for key, value := range params {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	return result
}

// PreviewTemplate 预览模板效果
func (s *MessageTemplateService) PreviewTemplate(params map[string]string) (map[string]string, error) {
	template, err := s.Get()
	if err != nil {
		return nil, err
	}
	
	result := make(map[string]string)
	
	if template.TextTemplate != "" {
		result["text"] = s.RenderTemplate(template.TextTemplate, params)
	}
	
	if template.HTMLTemplate != "" {
		result["html"] = s.RenderTemplate(template.HTMLTemplate, params)
	}
	
	if template.MarkdownTemplate != "" {
		result["markdown"] = s.RenderTemplate(template.MarkdownTemplate, params)
	}
	
	return result, nil
}

// validatePlaceholders 验证占位符格式
func (s *MessageTemplateService) validatePlaceholders() error {
	if s.Placeholders == "" {
		return nil
	}
	
	var placeholders []Placeholder
	if err := json.Unmarshal([]byte(s.Placeholders), &placeholders); err != nil {
		return errors.New("占位符格式错误，必须是有效的JSON数组")
	}
	
	for _, p := range placeholders {
		if p.Key == "" {
			return errors.New("占位符的key不能为空")
		}
	}
	
	return nil
}

// getMaps 获取查询条件
func (s *MessageTemplateService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	
	if s.Status != "" {
		maps["status"] = s.Status
	}
	
	return maps
}
