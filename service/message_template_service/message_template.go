package message_template_service

import (
	"encoding/json"
	"errors"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/util"
	"strings"
)

type TemplateService struct {
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
	SourceTemplateID string
	CopyInstances    bool
	Text             string
	CreatedBy        string
	ModifiedBy       string
	IsAdmin          bool

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
func (s *TemplateService) Add() error {
	if err := s.validatePlaceholders(); err != nil {
		return err
	}

	newUUID := models.GenerateTemplateUniqueID()
	model := models.Template{
		UUIDModel: models.UUIDModel{
			ID:         newUUID,
			CreatedBy:  s.CreatedBy,
			ModifiedBy: s.ModifiedBy,
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

	err := model.Add()
	if err != nil {
		return err
	}

	s.ID = newUUID

	if s.CopyInstances && s.SourceTemplateID != "" {
		insList, err := models.GetRawTemplateInsList(s.SourceTemplateID)
		if err == nil && len(insList) > 0 {
			for i := range insList {
				insList[i].ID = fmt.Sprintf("IN%s", util.GenerateUniqueID())
				insList[i].TemplateID = s.ID
				insList[i].TaskID = ""
				insList[i].CreatedBy = s.CreatedBy
				insList[i].ModifiedBy = s.ModifiedBy
			}
			_ = models.ManyAddTaskIns(insList)
		}
	}

	return nil
}

// Update 更新消息模板
func (s *TemplateService) Update() error {
	tenantID := ""
	if !s.IsAdmin && s.ModifiedBy != "" {
		tenantID = s.ModifiedBy
	}

	if err := s.validatePlaceholders(); err != nil {
		return err
	}

	model := models.Template{
		UUIDModel: models.UUIDModel{
			ID:         s.ID,
			ModifiedBy: s.ModifiedBy,
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

	return model.Update(tenantID)
}

// Delete 删除消息模板
func (s *TemplateService) Delete() error {
	tenantID := ""
	if !s.IsAdmin && s.ModifiedBy != "" {
		tenantID = s.ModifiedBy
	}

	model := models.Template{
		UUIDModel: models.UUIDModel{
			ID: s.ID,
		},
	}
	return model.Delete(tenantID)
}

// Get 获取单个消息模板
func (s *TemplateService) Get() (*models.TemplateResult, error) {
	return models.GetTemplateByID(s.ID)
}

// GetAll 获取消息模板列表
func (s *TemplateService) GetAll() ([]models.TemplateResult, error) {
	templates, err := models.GetTemplates(s.PageNum, s.PageSize, s.Text, s.getMaps())
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// Count 获取消息模板总数
func (s *TemplateService) Count() (int64, error) {
	return models.GetTemplatesTotal(s.Text, s.getMaps())
}

// ExistByID 检查模板是否存在
func (s *TemplateService) ExistByID() (bool, error) {
	return models.ExistTemplateByID(s.ID)
}

// RenderTemplate 渲染模板（替换占位符）
func (s *TemplateService) RenderTemplate(templateContent string, params map[string]string) string {
	result := templateContent

	for key, value := range params {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// PreviewTemplate 预览模板效果
func (s *TemplateService) PreviewTemplate(params map[string]string) (map[string]string, error) {
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
func (s *TemplateService) validatePlaceholders() error {
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
func (s *TemplateService) getMaps() map[string]interface{} {
	extra := make(map[string]interface{})
	if s.Status != "" {
		extra["status"] = s.Status
	}
	return models.TenantMap(s.IsAdmin, s.CreatedBy, extra)
}
