package app

import (
	"context"
)

type userContextKey struct{}

// UserContextInfo 保存请求上下文中的用户身份信息
type UserContextInfo struct {
	UserID            string // 真实登录用户ID（真实操作者）
	Username          string // 真实用户名
	Role              string // 真实角色
	ImpersonateUserID string // 管理员穿越的目标租户ID（为空表示未穿越）
}

// IsAdmin 判断当前用户是否具备管理员权限
func (u UserContextInfo) IsAdmin() bool {
	return u.Role == "admin" || u.Username == "admin"
}

// IsImpersonating 判断当前是否处于管理员穿越状态
func (u UserContextInfo) IsImpersonating() bool {
	return u.IsAdmin() && u.ImpersonateUserID != ""
}

// GetEffectiveUserID 获取生效的租户用户ID：穿越时为目标租户ID，否则为自身ID
func (u UserContextInfo) GetEffectiveUserID() string {
	if u.IsImpersonating() {
		return u.ImpersonateUserID
	}
	return u.UserID
}

// GetAuditUserID 永远返回真实登录操作者ID（用于 modified_by 审计溯源）
func (u UserContextInfo) GetAuditUserID() string {
	return u.UserID
}

// WithUserContext 将用户信息注入到 context.Context 中
func WithUserContext(ctx context.Context, info UserContextInfo) context.Context {
	return context.WithValue(ctx, userContextKey{}, info)
}

// GetUserFromContext 从 context.Context 中获取用户信息
func GetUserFromContext(ctx context.Context) (UserContextInfo, bool) {
	if ctx == nil {
		return UserContextInfo{}, false
	}
	info, ok := ctx.Value(userContextKey{}).(UserContextInfo)
	return info, ok
}

// GetCurrentUserIDFromContext 从 Context 获取当前生效租户用户ID（穿越时为目标租户ID）
func GetCurrentUserIDFromContext(ctx context.Context) string {
	if info, ok := GetUserFromContext(ctx); ok {
		return info.GetEffectiveUserID()
	}
	return ""
}

// GetAuditUserIDFromContext 从 Context 获取真实操作者用户ID（用于审计溯源）
func GetAuditUserIDFromContext(ctx context.Context) string {
	if info, ok := GetUserFromContext(ctx); ok {
		return info.GetAuditUserID()
	}
	return ""
}

// IsAdminContext 判断当前 Context 是否具备管理员权限
func IsAdminContext(ctx context.Context) bool {
	if info, ok := GetUserFromContext(ctx); ok {
		return info.IsAdmin()
	}
	return false
}
