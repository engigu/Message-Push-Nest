package models

import (
	"context"
	"errors"
	"gorm.io/gorm"
	"message-nest/pkg/app"
)

var (
	// ErrNoPermissionOrNotFound 无权操作或资源不存在错误
	ErrNoPermissionOrNotFound = errors.New("无权操作此资源或资源不存在")
)

// TenantScope 自动租户过滤 Scope：
// 1. 如果是管理员且指定了穿越目标租户，自动附加 WHERE created_by = ?，呈现目标客户视角
// 2. 如果是普通用户，自动附加 WHERE created_by = ?，强制租户隔离
// 3. 如果是管理员且未穿越，放行全量数据
func TenantScope(ctx context.Context) func(db *gorm.DB) *gorm.DB {
	return func(d *gorm.DB) *gorm.DB {
		if ctx == nil {
			return d
		}
		if user, ok := app.GetUserFromContext(ctx); ok {
			// 管理员处于穿越状态：限定目标租户视角
			if user.IsImpersonating() {
				return d.Where("created_by = ?", user.ImpersonateUserID)
			}
			// 普通用户：强制限定自身租户视角
			if !user.IsAdmin() && user.UserID != "" {
				return d.Where("created_by = ?", user.UserID)
			}
		}
		return d
	}
}

// CheckAccess 统一防越权检查：管理员直接放行，普通用户校验数据归属（ID 与 created_by）
func CheckAccess(isAdmin bool, userID string, tableOrModel interface{}, id interface{}) error {
	if isAdmin || userID == "" {
		return nil
	}
	var tableName string
	switch v := tableOrModel.(type) {
	case string:
		tableName = v
	default:
		tableName = GetSchema(v)
	}

	var count int64
	err := db.Table(tableName).Where("id = ? AND created_by = ?", id, userID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNoPermissionOrNotFound
	}
	return nil
}

// CheckOwnership 检查指定表记录的所有权，防止 IDOR 越权操作（管理员直接放行）
func CheckOwnership(ctx context.Context, table string, id interface{}) error {
	if ctx == nil {
		return nil
	}
	user, ok := app.GetUserFromContext(ctx)
	if !ok || user.UserID == "" {
		return ErrNoPermissionOrNotFound
	}
	return CheckAccess(user.IsAdmin(), user.GetEffectiveUserID(), table, id)
}

// WithAuditUpdate 为 map 更新自动填充 modified_by 审计信息（真实操作者）
func WithAuditUpdate(ctx context.Context, data map[string]interface{}) map[string]interface{} {
	if data == nil {
		data = make(map[string]interface{})
	}
	if ctx != nil {
		if uid := app.GetAuditUserIDFromContext(ctx); uid != "" {
			data["modified_by"] = uid
		}
	}
	return data
}

// TenantMap 统一生成/合并租户过滤条件 map（管理员放行，普通用户追加 created_by）
func TenantMap(isAdmin bool, userID string, extra ...map[string]interface{}) map[string]interface{} {
	maps := make(map[string]interface{})
	if len(extra) > 0 && extra[0] != nil {
		for k, v := range extra[0] {
			maps[k] = v
		}
	}
	if !isAdmin && userID != "" {
		maps["created_by"] = userID
	}
	return maps
}
