package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"message-nest/pkg/util"
)

const (
	RateLimitRetryMax  = 2
	RateLimitBatchSize = 500

	AliyunSMSTestUserLimit = 10
	AliyunSMSTestWayLimit  = 30
	AliyunSMSTestWindow    = time.Minute
	AliyunSMSTestRepeat    = 20 * time.Second
)

// RateLimitScopeLock 通用限流作用域排他锁表
type RateLimitScopeLock struct {
	BizType     string `gorm:"type:varchar(32);primaryKey"`
	ScopeHash   string `gorm:"type:char(64);primaryKey"`
	ScopeType   string `gorm:"type:varchar(16);not null"`
	TouchedAtMS int64  `gorm:"not null"`
}

func (s *RateLimitScopeLock) BeforeCreate(tx *gorm.DB) error {
	if s.BizType == "" {
		s.BizType = "default"
	}
	return nil
}

// RateLimitEvent 通用滑动窗口限流事件表（方案 A：标准维度-事件表）
type RateLimitEvent struct {
	ID          string `gorm:"type:varchar(12);primaryKey;index:idx_rl_evt_created_id,priority:2"`
	BizType     string `gorm:"type:varchar(32);not null;index:idx_rl_evt_scope_time,priority:1;index:idx_rl_evt_req_time,priority:1"`
	ScopeType   string `gorm:"type:varchar(16);not null;index:idx_rl_evt_scope_time,priority:2"`
	ScopeHash   string `gorm:"type:char(64);not null;index:idx_rl_evt_scope_time,priority:3"`
	RequestHash string `gorm:"type:char(64);not null;index:idx_rl_evt_req_time,priority:2"`
	CreatedAtMS int64  `gorm:"not null;index:idx_rl_evt_scope_time,priority:4;index:idx_rl_evt_req_time,priority:3;index:idx_rl_evt_created_id,priority:1"`
}

func (e *RateLimitEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = fmt.Sprintf("EV%s", util.GenerateUniqueID())
	}
	if e.BizType == "" {
		e.BizType = "default"
	}
	return nil
}

// RateLimitScopeRule 作用域滑动窗口限流规则
type RateLimitScopeRule struct {
	ScopeType string
	Limit     int64
	Window    time.Duration
}

// RateLimitScopeTarget 限流判定的作用域目标
type RateLimitScopeTarget struct {
	ScopeType string
	ScopeHash string
}

// RateLimitConfig 通用限流配置
type RateLimitConfig struct {
	BizType      string
	Rules        []RateLimitScopeRule
	RepeatWindow time.Duration
}

// RateLimitDecision 限流判定结果
type RateLimitDecision struct {
	Allowed           bool
	Reason            string
	RetryAfterSeconds int
}

// 阿里云短信测试限流默认配置
var DefaultAliyunSMSTestRateLimitConfig = RateLimitConfig{
	BizType: "aliyun_sms_test",
	Rules: []RateLimitScopeRule{
		{ScopeType: "user", Limit: AliyunSMSTestUserLimit, Window: AliyunSMSTestWindow},
		{ScopeType: "way", Limit: AliyunSMSTestWayLimit, Window: AliyunSMSTestWindow},
	},
	RepeatWindow: AliyunSMSTestRepeat,
}

// ReserveRateLimit 通用滑动窗口与防重限流决策
func ReserveRateLimit(ctx context.Context, cfg RateLimitConfig, targets []RateLimitScopeTarget, requestHash string) (RateLimitDecision, error) {
	if db == nil {
		return RateLimitDecision{}, errors.New("database is not initialized")
	}
	if cfg.BizType == "" {
		return RateLimitDecision{}, errors.New("rate-limit biz_type is required")
	}
	if len(targets) == 0 {
		return RateLimitDecision{}, errors.New("rate-limit scope targets are required")
	}
	for _, target := range targets {
		if target.ScopeType == "" || target.ScopeHash == "" {
			return RateLimitDecision{}, errors.New("rate-limit scope target type and hash are required")
		}
	}

	database := db.WithContext(ctx)
	options := &sql.TxOptions{}
	switch database.Dialector.Name() {
	case "mysql", "tidb", "postgres":
		options.Isolation = sql.LevelReadCommitted
	case "sqlite":
	case "":
		return RateLimitDecision{}, errors.New("unknown database dialect")
	default:
		return RateLimitDecision{}, fmt.Errorf("unsupported rate-limit database dialect: %s", database.Dialector.Name())
	}

	var decision RateLimitDecision
	var lastErr error
	for attempt := 0; attempt <= RateLimitRetryMax; attempt++ {
		decision = RateLimitDecision{}
		lastErr = database.Transaction(func(tx *gorm.DB) error {
			return reserveRateLimitTx(tx, database.Dialector.Name(), cfg, targets, requestHash, &decision)
		}, options)
		if lastErr == nil {
			return decision, nil
		}
		if attempt == RateLimitRetryMax || !IsRetryableRateLimitError(lastErr) {
			break
		}
		delay := time.Duration(15*(attempt+1)) * time.Millisecond
		select {
		case <-ctx.Done():
			return RateLimitDecision{}, ctx.Err()
		case <-time.After(delay):
		}
	}
	return RateLimitDecision{}, lastErr
}

// ReserveAliyunSMSTestRate 阿里云短信测试限流入口
func ReserveAliyunSMSTestRate(ctx context.Context, userHash, wayHash, requestHash string) (RateLimitDecision, error) {
	targets := []RateLimitScopeTarget{
		{ScopeType: "user", ScopeHash: userHash},
		{ScopeType: "way", ScopeHash: wayHash},
	}
	return ReserveRateLimit(ctx, DefaultAliyunSMSTestRateLimitConfig, targets, requestHash)
}

func reserveRateLimitTx(tx *gorm.DB, dialect string, cfg RateLimitConfig, targets []RateLimitScopeTarget, requestHash string, decision *RateLimitDecision) error {
	targetsMap := make(map[string]RateLimitScopeTarget, len(targets))
	for _, t := range targets {
		key := t.ScopeType + ":" + t.ScopeHash
		targetsMap[key] = t
	}
	normalizedTargets := make([]RateLimitScopeTarget, 0, len(targetsMap))
	for _, t := range targetsMap {
		normalizedTargets = append(normalizedTargets, t)
	}
	sort.Slice(normalizedTargets, func(i, j int) bool {
		if normalizedTargets[i].ScopeHash == normalizedTargets[j].ScopeHash {
			return normalizedTargets[i].ScopeType < normalizedTargets[j].ScopeType
		}
		return normalizedTargets[i].ScopeHash < normalizedTargets[j].ScopeHash
	})

	lockMap := make(map[string]RateLimitScopeLock, len(normalizedTargets))
	for _, t := range normalizedTargets {
		if _, exists := lockMap[t.ScopeHash]; !exists {
			lockMap[t.ScopeHash] = RateLimitScopeLock{
				BizType:   cfg.BizType,
				ScopeHash: t.ScopeHash,
				ScopeType: t.ScopeType,
			}
		}
	}
	lockRows := make([]RateLimitScopeLock, 0, len(lockMap))
	for _, row := range lockMap {
		lockRows = append(lockRows, row)
	}
	sort.Slice(lockRows, func(i, j int) bool {
		return lockRows[i].ScopeHash < lockRows[j].ScopeHash
	})
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&lockRows).Error; err != nil {
		return err
	}

	nowMS, err := databaseNowMillis(tx, dialect)
	if err != nil {
		return err
	}

	for _, row := range lockRows {
		result := tx.Model(&RateLimitScopeLock{}).
			Where("biz_type = ? AND scope_hash = ?", row.BizType, row.ScopeHash).
			UpdateColumn("touched_at_ms", gorm.Expr("CASE WHEN touched_at_ms >= ? THEN touched_at_ms + 1 ELSE ? END", nowMS, nowMS))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return errors.New("rate-limit scope lock row missing")
		}
	}

	if cfg.RepeatWindow > 0 && requestHash != "" {
		repeatCutoff := nowMS - cfg.RepeatWindow.Milliseconds()
		var duplicateCount int64
		if err := tx.Model(&RateLimitEvent{}).
			Where("biz_type = ? AND request_hash = ? AND created_at_ms > ?", cfg.BizType, requestHash, repeatCutoff).
			Count(&duplicateCount).Error; err != nil {
			return err
		}
		if duplicateCount > 0 {
			var oldest int64
			if err := tx.Model(&RateLimitEvent{}).Select("MIN(created_at_ms)").
				Where("biz_type = ? AND request_hash = ? AND created_at_ms > ?", cfg.BizType, requestHash, repeatCutoff).
				Scan(&oldest).Error; err != nil {
				return err
			}
			decision.Allowed = false
			decision.Reason = "duplicate"
			decision.RetryAfterSeconds = retryAfterSeconds(oldest, cfg.RepeatWindow.Milliseconds(), nowMS)
			return nil
		}
	}

	ruleMap := make(map[string]RateLimitScopeRule, len(cfg.Rules))
	for _, r := range cfg.Rules {
		ruleMap[r.ScopeType] = r
	}

	var maxRetry int
	var rateLimited bool
	for _, target := range normalizedTargets {
		rule, ok := ruleMap[target.ScopeType]
		if !ok || rule.Limit <= 0 || rule.Window <= 0 {
			continue
		}
		windowCutoff := nowMS - rule.Window.Milliseconds()
		var count int64
		if err := tx.Model(&RateLimitEvent{}).
			Where("biz_type = ? AND scope_type = ? AND scope_hash = ? AND created_at_ms > ?", cfg.BizType, target.ScopeType, target.ScopeHash, windowCutoff).
			Count(&count).Error; err != nil {
			return err
		}
		if count >= rule.Limit {
			var oldest int64
			if err := tx.Model(&RateLimitEvent{}).Select("MIN(created_at_ms)").
				Where("biz_type = ? AND scope_type = ? AND scope_hash = ? AND created_at_ms > ?", cfg.BizType, target.ScopeType, target.ScopeHash, windowCutoff).
				Scan(&oldest).Error; err != nil {
				return err
			}
			retry := retryAfterSeconds(oldest, rule.Window.Milliseconds(), nowMS)
			if retry > maxRetry {
				maxRetry = retry
			}
			rateLimited = true
		}
	}

	if rateLimited {
		decision.Allowed = false
		decision.Reason = "rate_limited"
		decision.RetryAfterSeconds = maxRetry
		return nil
	}

	events := make([]RateLimitEvent, len(normalizedTargets))
	for i, target := range normalizedTargets {
		events[i] = RateLimitEvent{
			ID:          fmt.Sprintf("EV%s", util.GenerateUniqueID()),
			BizType:     cfg.BizType,
			ScopeType:   target.ScopeType,
			ScopeHash:   target.ScopeHash,
			RequestHash: requestHash,
			CreatedAtMS: nowMS,
		}
	}
	if err := tx.Create(&events).Error; err != nil {
		return err
	}
	decision.Allowed = true
	return nil
}

func retryAfterSeconds(oldest, windowMS, nowMS int64) int {
	remaining := oldest + windowMS - nowMS
	if remaining <= 0 {
		return 1
	}
	return max(1, int(math.Ceil(float64(remaining)/float64(time.Second.Milliseconds()))))
}

func databaseNowMillis(tx *gorm.DB, dialect string) (int64, error) {
	if tx == nil {
		return 0, errors.New("database transaction is nil")
	}
	if dialect == "" && tx.Dialector != nil {
		dialect = tx.Dialector.Name()
	}

	var now int64
	query := ""
	switch dialect {
	case "mysql", "tidb":
		query = "SELECT CAST(TIMESTAMPDIFF(MICROSECOND, '1970-01-01 00:00:00', UTC_TIMESTAMP(3)) / 1000 AS SIGNED)"
	case "postgres":
		query = "SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000)::bigint"
	case "sqlite":
		query = "SELECT CAST((julianday('now') - 2440587.5) * 86400000 AS INTEGER)"
	default:
		return 0, fmt.Errorf("unsupported database dialect for rate limiting: %s", dialect)
	}

	if err := tx.Raw(query).Scan(&now).Error; err != nil {
		return 0, fmt.Errorf("query database now millis failed: %w", err)
	}
	if now <= 0 {
		return 0, errors.New("database returned invalid rate-limit timestamp")
	}
	return now, nil
}

// CleanupRateLimitEvents 批量清理过期事件数据，bizType 为空时清理全部业务
func CleanupRateLimitEvents(ctx context.Context, bizType string, retention time.Duration, maxBatches int) (int64, error) {
	if db == nil {
		return 0, errors.New("database is not initialized")
	}
	if retention <= 0 || maxBatches <= 0 {
		return 0, errors.New("invalid rate-limit event cleanup settings")
	}
	database := db.WithContext(ctx)
	dialect := database.Dialector.Name()
	cutoffDB, err := databaseNowMillis(database, dialect)
	if err != nil {
		return 0, err
	}
	cutoff := cutoffDB - retention.Milliseconds()
	var deleted int64
	for batch := 0; batch < maxBatches; batch++ {
		var ids []string
		query := database.Model(&RateLimitEvent{}).Where("created_at_ms < ?", cutoff)
		if bizType != "" {
			query = query.Where("biz_type = ?", bizType)
		}
		if err := query.Order("created_at_ms ASC, id ASC").Limit(RateLimitBatchSize).Pluck("id", &ids).Error; err != nil {
			return deleted, err
		}
		if len(ids) == 0 {
			break
		}
		delQuery := database.Where("id IN ? AND created_at_ms < ?", ids, cutoff)
		if bizType != "" {
			delQuery = delQuery.Where("biz_type = ?", bizType)
		}
		result := delQuery.Delete(&RateLimitEvent{})
		if result.Error != nil {
			return deleted, result.Error
		}
		deleted += result.RowsAffected
		if len(ids) < RateLimitBatchSize {
			break
		}
	}
	return deleted, nil
}

func IsRetryableRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "database is locked") || strings.Contains(message, "sqlite_busy") || strings.Contains(message, "database table is locked") {
		return true
	}
	var stateErr interface{ SQLState() string }
	if errors.As(err, &stateErr) {
		state := stateErr.SQLState()
		if state == "40001" || state == "40P01" {
			return true
		}
	}
	return strings.Contains(message, "write conflict") ||
		strings.Contains(message, "deadlock found") ||
		strings.Contains(message, "lock wait timeout") ||
		strings.Contains(message, "error 9007") ||
		strings.Contains(message, "error 1213") ||
		strings.Contains(message, "error 1205")
}
