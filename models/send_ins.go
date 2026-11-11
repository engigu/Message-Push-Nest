package models

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

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

// InsQyWeiXinAppConfig 实例里面的企业微信自建应用config
type InsQyWeiXinAppConfig struct {
	ToUser  string `json:"to_user" validate:"required" label:"接收成员ID"`
	ToParty string `json:"to_party" label:"接收部门ID"`
	ToTag   string `json:"to_tag" label:"接收标签ID"`
}

// InsFeishuConfig 实例里面的飞书config
type InsFeishuConfig struct {
}

// InsCustomConfig 实例里面的自定义config
type InsCustomConfig struct {
}

// InsMessageNestConfig 实例里面的托管消息config
type InsMessageNestConfig struct {
	BlacklistRegex string `json:"blacklist_regex" label:"黑名单正则表达式"`
}

// InsAliyunSMSConfig 实例里面的阿里云短信config
type InsAliyunSMSConfig struct {
	PhoneNumber        string            `json:"phone_number" validate:"required" label:"手机号码"`
	TemplateCode       string            `json:"template_code" validate:"required" label:"短信模板CODE"`
	TemplateParamMode  string            `json:"template_param_mode,omitempty"`
	SelectedParams     []string          `json:"selected_params,omitempty"`
	ParamNameOverrides map[string]string `json:"param_name_overrides,omitempty"`
}

// ValidateTemplateParamConfig validates the optional SMS parameter contract.
// A missing mode remains valid for historical content-mode instances.
func (c InsAliyunSMSConfig) ValidateTemplateParamConfig() error {
	switch c.TemplateParamMode {
	case "", "content":
		return nil
	case "structured":
	default:
		return fmt.Errorf("不支持的阿里云短信参数模式: %q", c.TemplateParamMode)
	}
	if c.SelectedParams == nil {
		return fmt.Errorf("structured 模式必须显式提供 selected_params")
	}

	selected := make(map[string]struct{}, len(c.SelectedParams))
	resolvedTargets := make(map[string]string, len(c.SelectedParams))
	for _, sourceName := range c.SelectedParams {
		if strings.TrimSpace(sourceName) == "" || strings.TrimSpace(sourceName) != sourceName {
			return fmt.Errorf("selected_params 中的参数名不能为空或包含首尾空格")
		}
		if _, exists := selected[sourceName]; exists {
			return fmt.Errorf("selected_params 中存在重复参数 %q", sourceName)
		}
		selected[sourceName] = struct{}{}

		targetName := sourceName
		if override, exists := c.ParamNameOverrides[sourceName]; exists {
			if strings.TrimSpace(override) == "" || strings.TrimSpace(override) != override {
				return fmt.Errorf("参数 %q 的阿里云映射名称不能为空或包含首尾空格", sourceName)
			}
			if override == sourceName {
				return fmt.Errorf("参数 %q 是同名映射，不应保存为覆盖项", sourceName)
			}
			targetName = override
		}
		if previousSource, exists := resolvedTargets[targetName]; exists {
			return fmt.Errorf("参数 %q 与 %q 映射到重复的阿里云参数 %q", previousSource, sourceName, targetName)
		}
		resolvedTargets[targetName] = sourceName
	}

	for sourceName := range c.ParamNameOverrides {
		if _, exists := selected[sourceName]; !exists {
			return fmt.Errorf("参数 %q 有映射，但不在 selected_params 中", sourceName)
		}
	}
	return nil
}

// InsTelegramConfig 实例里面的Telegram config
type InsTelegramConfig struct {
}

// InsBarkConfig 实例里面的Bark config
type InsBarkConfig struct {
}

// InsPushMeConfig 实例里面的PushMe config
type InsPushMeConfig struct {
}

// InsNtfyConfig 实例里面的Ntfy config
type InsNtfyConfig struct {
}

// InsGotifyConfig 实例里面的Gotify config
type InsGotifyConfig struct {
}

// ManyAddTaskIns 批量添加实例（批量排重 + 批量单次插入，消灭 N+1 查询）
func ManyAddTaskIns(taskIns []SendTasksIns) error {
	if len(taskIns) == 0 {
		return nil
	}

	// 1. 批量收集待插入的 ID
	ids := make([]string, 0, len(taskIns))
	for _, ins := range taskIns {
		if ins.ID != "" {
			ids = append(ids, ins.ID)
		}
	}

	// 2. 一次性批量查询数据库中已存在的 ID
	existingIDs := make(map[string]bool)
	if len(ids) > 0 {
		var foundIDs []string
		if err := db.Model(&SendTasksIns{}).Where("id IN ?", ids).Pluck("id", &foundIDs).Error; err == nil {
			for _, id := range foundIDs {
				existingIDs[id] = true
			}
		}
	}

	// 3. 过滤出真正需要新增的记录
	toInsert := make([]SendTasksIns, 0, len(taskIns))
	for _, ins := range taskIns {
		if ins.ID != "" && existingIDs[ins.ID] {
			continue
		}
		toInsert = append(toInsert, ins)
	}

	if len(toInsert) == 0 {
		return nil
	}

	// 4. 单次批量事务写入 (分批大小为 100)
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.CreateInBatches(toInsert, 100).Error
	})
}

// AddTaskInsOne 添加一条实例
func AddTaskInsOne(ins SendTasksIns) error {
	if err := db.Create(&ins).Error; err != nil {
		return err
	}
	return nil
}

// DeleteMsgTaskInsForOwner 在指定所有者权限范围内删除实例
func DeleteMsgTaskInsForOwner(id, ownerID string) error {
	query := db.Where("id = ?", id)
	if ownerID != "" {
		query = query.Where("(created_by = ? OR created_by = '')", ownerID)
	}
	res := query.Delete(&SendTasksIns{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNoPermissionOrNotFound
	}
	return nil
}

// DeleteMsgTaskIns 删除一条实例
func DeleteMsgTaskIns(id string) error {
	return DeleteMsgTaskInsForOwner(id, "")
}

// UpdateMsgTaskInsForOwner 在指定所有者权限范围内更新实例
func UpdateMsgTaskInsForOwner(id, ownerID string, data map[string]interface{}) error {
	query := db.Model(&SendTasksIns{}).Where("id = ?", id)
	if ownerID != "" {
		query = query.Where("(created_by = ? OR created_by = '')", ownerID)
	}
	res := query.Updates(data)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var count int64
		check := db.Model(&SendTasksIns{}).Where("id = ?", id)
		if ownerID != "" {
			check = check.Where("(created_by = ? OR created_by = '')", ownerID)
		}
		if err := check.Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNoPermissionOrNotFound
		}
	}
	return nil
}

// UpdateMsgTaskIns 更新实例
func UpdateMsgTaskIns(id string, data map[string]interface{}) error {
	return UpdateMsgTaskInsForOwner(id, "", data)
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

// GetTemplateInsListForOwner limits template instance listings to the
// effective owner scope used by the authenticated API caller.
func GetTemplateInsListForOwner(templateID, ownerID string) ([]SendTasksInsRes, error) {
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})
	var insList []SendTasksInsRes
	query := db.Table(insTable).
		Select(fmt.Sprintf("%s.*, %s.name as way_name", insTable, waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Where(fmt.Sprintf("%s.template_id = ?", insTable), templateID)
	if ownerID != "" {
		query = query.Where(fmt.Sprintf("(%s.created_by = ? OR %s.created_by = '')", insTable, insTable), ownerID).
			Where(fmt.Sprintf("%s.created_by = ?", waysTable), ownerID)
	}
	err := query.Order(fmt.Sprintf("%s.created_on DESC", insTable)).Scan(&insList).Error
	if err != nil {
		return nil, err
	}
	return insList, nil
}

// GetRawTemplateInsList 获取模板关联的原始实例列表
func GetRawTemplateInsList(templateID string) ([]SendTasksIns, error) {
	var insList []SendTasksIns
	err := db.Where("template_id = ?", templateID).Find(&insList).Error
	if err != nil {
		return nil, err
	}
	return insList, nil
}

// GetRawTemplateInsListForOwner 获取租户权限范围内模板关联的原始实例列表（校验实例及渠道所有权）
func GetRawTemplateInsListForOwner(templateID, ownerID string) ([]SendTasksIns, error) {
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})
	var insList []SendTasksIns
	query := db.Table(insTable).
		Select(fmt.Sprintf("%s.*", insTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Where(fmt.Sprintf("%s.template_id = ?", insTable), templateID)
	if ownerID != "" {
		query = query.Where(fmt.Sprintf("(%s.created_by = ? OR %s.created_by = '')", insTable, insTable), ownerID).
			Where(fmt.Sprintf("%s.created_by = ?", waysTable), ownerID)
	}
	err := query.Order(fmt.Sprintf("%s.created_on DESC", insTable)).Scan(&insList).Error
	if err != nil {
		return nil, err
	}
	return insList, nil
}

// GetTemplateInstanceByIDForOwner loads only a template-bound instance in the
// caller's effective owner scope. An empty owner is reserved for global admins.
func GetTemplateInstanceByIDForOwner(id, ownerID string) (SendTasksIns, error) {
	var ins SendTasksIns
	query := db.Where("id = ? AND template_id <> ? AND task_id = ?", id, "", "")
	if ownerID != "" {
		query = query.Where("(created_by = ? OR created_by = '')", ownerID)
	}
	err := query.First(&ins).Error
	return ins, err
}

// UpdateTemplateInstanceConfig updates only Config and its audit actor while
// applying the owner predicate in the same write query.
func UpdateTemplateInstanceConfig(id, ownerID, actorID, config string) error {
	query := db.Model(&SendTasksIns{}).Where("id = ? AND template_id <> ? AND task_id = ?", id, "", "")
	if ownerID != "" {
		query = query.Where("(created_by = ? OR created_by = '')", ownerID)
	}
	result := query.Updates(map[string]interface{}{
		"config":      config,
		"modified_by": actorID,
	})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		check := db.Model(&SendTasksIns{}).Where("id = ? AND template_id <> ? AND task_id = ?", id, "", "")
		if ownerID != "" {
			check = check.Where("(created_by = ? OR created_by = '')", ownerID)
		}
		if err := check.Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrNoPermissionOrNotFound
		}
	}
	return nil
}
