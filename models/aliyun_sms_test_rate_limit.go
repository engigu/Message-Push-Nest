package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	AliyunSMSTestUserLimit = 10
	AliyunSMSTestWayLimit  = 30
	AliyunSMSTestWindow    = time.Minute
	AliyunSMSTestRepeat    = 20 * time.Second
	aliyunSMSTestRetryMax  = 2
	aliyunSMSTestBatchSize = 500
)

// AliyunSMSTestScopeLock serializes reservations for one pseudonymous user or Way.
type AliyunSMSTestScopeLock struct {
	ScopeHash   string `gorm:"type:char(64);primaryKey"`
	ScopeType   string `gorm:"type:varchar(8);not null"`
	TouchedAtMS int64  `gorm:"not null"`
}

// AliyunSMSTestEvent contains only keyed fingerprints and the database timestamp.
type AliyunSMSTestEvent struct {
	ID          uint   `gorm:"primaryKey;autoIncrement;index:idx_alisms_evt_created_id,priority:2"`
	UserHash    string `gorm:"type:char(64);not null;index:idx_alisms_evt_user_time,priority:1"`
	WayHash     string `gorm:"type:char(64);not null;index:idx_alisms_evt_way_time,priority:1"`
	RequestHash string `gorm:"type:char(64);not null;index:idx_alisms_evt_request_time,priority:1"`
	CreatedAtMS int64  `gorm:"not null;index:idx_alisms_evt_user_time,priority:2;index:idx_alisms_evt_way_time,priority:2;index:idx_alisms_evt_request_time,priority:2;index:idx_alisms_evt_created_id,priority:1"`
}

type AliyunSMSTestLimitDecision struct {
	Allowed           bool
	Reason            string
	RetryAfterSeconds int
}

// MigrateAliyunSMSTestRateLimitTables only creates or upgrades the two dedicated
// tables. It intentionally does not invoke the application's model-signature migration.
func MigrateAliyunSMSTestRateLimitTables(database *gorm.DB) error {
	if database == nil {
		return errors.New("database is not initialized")
	}
	migrationDB := database
	if database.Dialector.Name() == "mysql" {
		migrationDB = database.Set("gorm:table_options", "ENGINE=InnoDB")
	}
	if err := migrationDB.AutoMigrate(&AliyunSMSTestScopeLock{}, &AliyunSMSTestEvent{}); err != nil {
		return fmt.Errorf("migrate Aliyun SMS test rate-limit tables: %w", err)
	}
	for _, check := range []struct {
		model any
		index string
	}{
		{&AliyunSMSTestScopeLock{}, "PRIMARY"},
		{&AliyunSMSTestEvent{}, "idx_alisms_evt_user_time"},
		{&AliyunSMSTestEvent{}, "idx_alisms_evt_way_time"},
		{&AliyunSMSTestEvent{}, "idx_alisms_evt_request_time"},
		{&AliyunSMSTestEvent{}, "idx_alisms_evt_created_id"},
	} {
		if !database.Migrator().HasTable(check.model) {
			return fmt.Errorf("rate-limit table missing after migration")
		}
		if check.index != "PRIMARY" && !database.Migrator().HasIndex(check.model, check.index) {
			return fmt.Errorf("rate-limit index %s missing after migration", check.index)
		}
	}
	if database.Dialector.Name() == "mysql" {
		for _, model := range []any{&AliyunSMSTestScopeLock{}, &AliyunSMSTestEvent{}} {
			statement := &gorm.Statement{DB: database}
			if err := statement.Parse(model); err != nil {
				return fmt.Errorf("resolve rate-limit table name: %w", err)
			}
			table := statement.Table
			var engine string
			if err := database.Raw("SELECT ENGINE FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", table).Scan(&engine).Error; err != nil {
				return fmt.Errorf("verify limiter table storage engine: %w", err)
			}
			if !strings.EqualFold(engine, "InnoDB") {
				return fmt.Errorf("rate-limit table %s must use InnoDB", table)
			}
		}
	}
	return nil
}

// ReserveAliyunSMSTestRate atomically checks both sliding windows and the repeat
// window, then records one event. The provider call must happen after this returns.
func ReserveAliyunSMSTestRate(ctx context.Context, userHash, wayHash, requestHash string) (AliyunSMSTestLimitDecision, error) {
	if db == nil {
		return AliyunSMSTestLimitDecision{}, errors.New("database is not initialized")
	}
	if userHash == "" || wayHash == "" || requestHash == "" {
		return AliyunSMSTestLimitDecision{}, errors.New("rate-limit fingerprints are required")
	}

	database := db.WithContext(ctx)
	options := &sql.TxOptions{}
	switch database.Dialector.Name() {
	case "mysql":
		// TiDB supports READ COMMITTED as well. In pessimistic mode this gives
		// post-lock event queries a current view; optimistic conflicts fail at
		// commit and are retried below.
		options.Isolation = sql.LevelReadCommitted
	case "postgres":
		options.Isolation = sql.LevelReadCommitted
	case "sqlite":
		// The first INSERT/UPDATE takes SQLite's database-level writer lock.
	case "":
		return AliyunSMSTestLimitDecision{}, errors.New("unknown database dialect")
	default:
		return AliyunSMSTestLimitDecision{}, fmt.Errorf("unsupported rate-limit database dialect: %s", database.Dialector.Name())
	}

	var decision AliyunSMSTestLimitDecision
	var lastErr error
	for attempt := 0; attempt <= aliyunSMSTestRetryMax; attempt++ {
		decision = AliyunSMSTestLimitDecision{}
		lastErr = database.Transaction(func(tx *gorm.DB) error {
			return reserveAliyunSMSTestRateTx(tx, database.Dialector.Name(), userHash, wayHash, requestHash, &decision)
		}, options)
		if lastErr == nil {
			return decision, nil
		}
		if attempt == aliyunSMSTestRetryMax || !isRetryableAliyunRateLimitError(lastErr) {
			break
		}
		delay := time.Duration(15*(attempt+1)) * time.Millisecond
		select {
		case <-ctx.Done():
			return AliyunSMSTestLimitDecision{}, ctx.Err()
		case <-time.After(delay):
		}
	}
	return AliyunSMSTestLimitDecision{}, lastErr
}

func reserveAliyunSMSTestRateTx(tx *gorm.DB, dialect, userHash, wayHash, requestHash string, decision *AliyunSMSTestLimitDecision) error {
	lockRows := []AliyunSMSTestScopeLock{
		{ScopeHash: userHash, ScopeType: "user"},
		{ScopeHash: wayHash, ScopeType: "way"},
	}
	if userHash == wayHash {
		return errors.New("user and Way rate-limit scopes must be distinct")
	}
	if lockRows[0].ScopeHash > lockRows[1].ScopeHash {
		lockRows[0], lockRows[1] = lockRows[1], lockRows[0]
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&lockRows).Error; err != nil {
		return err
	}

	scopes := []string{userHash, wayHash}
	if scopes[0] > scopes[1] {
		scopes[0], scopes[1] = scopes[1], scopes[0]
	}
	lockNow, err := databaseNowMillis(tx, dialect)
	if err != nil {
		return err
	}
	// A write-locking update works in optimistic TiDB transactions too: a concurrent
	// commit to either same scope conflicts and is retried by the outer function.
	for _, scopeHash := range scopes {
		result := tx.Model(&AliyunSMSTestScopeLock{}).
			Where("scope_hash = ?", scopeHash).
			UpdateColumn("touched_at_ms", gorm.Expr("CASE WHEN touched_at_ms >= ? THEN touched_at_ms + 1 ELSE ? END", lockNow, lockNow))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("rate-limit scope lock row missing")
		}
	}
	if dialect == "mysql" || dialect == "postgres" {
		var locked []AliyunSMSTestScopeLock
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("scope_hash IN ?", scopes).Order("scope_hash ASC").Find(&locked).Error; err != nil {
			return err
		}
		if len(locked) != 2 {
			return errors.New("rate-limit scope lock rows unavailable")
		}
	}

	nowMS, err := databaseNowMillis(tx, dialect)
	if err != nil {
		return err
	}
	windowCutoff := nowMS - AliyunSMSTestWindow.Milliseconds()
	repeatCutoff := nowMS - AliyunSMSTestRepeat.Milliseconds()

	var duplicateCount int64
	if err := tx.Model(&AliyunSMSTestEvent{}).
		Where("request_hash = ? AND created_at_ms > ?", requestHash, repeatCutoff).
		Count(&duplicateCount).Error; err != nil {
		return err
	}
	if duplicateCount > 0 {
		var oldest int64
		if err := tx.Model(&AliyunSMSTestEvent{}).Select("MIN(created_at_ms)").
			Where("request_hash = ? AND created_at_ms > ?", requestHash, repeatCutoff).
			Scan(&oldest).Error; err != nil {
			return err
		}
		decision.Allowed = false
		decision.Reason = "duplicate"
		decision.RetryAfterSeconds = retryAfterSeconds(oldest, AliyunSMSTestRepeat.Milliseconds(), nowMS)
		return nil
	}

	var userCount, wayCount int64
	if err := tx.Model(&AliyunSMSTestEvent{}).
		Where("user_hash = ? AND created_at_ms > ?", userHash, windowCutoff).
		Count(&userCount).Error; err != nil {
		return err
	}
	if err := tx.Model(&AliyunSMSTestEvent{}).
		Where("way_hash = ? AND created_at_ms > ?", wayHash, windowCutoff).
		Count(&wayCount).Error; err != nil {
		return err
	}
	if userCount >= AliyunSMSTestUserLimit || wayCount >= AliyunSMSTestWayLimit {
		retryAfter := 0
		if userCount >= AliyunSMSTestUserLimit {
			retry, err := oldestEventRetry(tx, "user_hash", userHash, windowCutoff, AliyunSMSTestWindow.Milliseconds(), nowMS)
			if err != nil {
				return err
			}
			retryAfter = max(retryAfter, retry)
		}
		if wayCount >= AliyunSMSTestWayLimit {
			retry, err := oldestEventRetry(tx, "way_hash", wayHash, windowCutoff, AliyunSMSTestWindow.Milliseconds(), nowMS)
			if err != nil {
				return err
			}
			retryAfter = max(retryAfter, retry)
		}
		decision.Allowed = false
		decision.Reason = "rate_limited"
		decision.RetryAfterSeconds = retryAfter
		return nil
	}

	if err := tx.Create(&AliyunSMSTestEvent{
		UserHash: userHash, WayHash: wayHash, RequestHash: requestHash, CreatedAtMS: nowMS,
	}).Error; err != nil {
		return err
	}
	decision.Allowed = true
	return nil
}

func oldestEventRetry(tx *gorm.DB, column, hash string, cutoff, windowMS, nowMS int64) (int, error) {
	if column != "user_hash" && column != "way_hash" {
		return 0, errors.New("invalid rate-limit scope column")
	}
	var oldest int64
	if err := tx.Model(&AliyunSMSTestEvent{}).Select("MIN(created_at_ms)").
		Where(column+" = ? AND created_at_ms > ?", hash, cutoff).Scan(&oldest).Error; err != nil {
		return 0, err
	}
	return retryAfterSeconds(oldest, windowMS, nowMS), nil
}

func retryAfterSeconds(oldest, windowMS, nowMS int64) int {
	remaining := oldest + windowMS - nowMS
	if remaining <= 0 {
		return 1
	}
	return max(1, int(math.Ceil(float64(remaining)/float64(time.Second.Milliseconds()))))
}

func databaseNowMillis(tx *gorm.DB, dialect string) (int64, error) {
	var now int64
	query := ""
	switch dialect {
	case "mysql":
		query = "SELECT CAST(TIMESTAMPDIFF(MICROSECOND, '1970-01-01 00:00:00', UTC_TIMESTAMP(3)) / 1000 AS SIGNED)"
	case "postgres":
		query = "SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000)::bigint"
	case "sqlite":
		query = "SELECT CAST((julianday('now') - 2440587.5) * 86400000 AS INTEGER)"
	default:
		return 0, fmt.Errorf("unsupported database dialect for rate limiting: %s", dialect)
	}
	if err := tx.Raw(query).Scan(&now).Error; err != nil {
		return 0, err
	}
	if now <= 0 {
		return 0, errors.New("database returned invalid rate-limit time")
	}
	return now, nil
}

// CleanupAliyunSMSTestEvents deletes only events older than retention in bounded batches.
func CleanupAliyunSMSTestEvents(ctx context.Context, retention time.Duration, maxBatches int) (int64, error) {
	if db == nil {
		return 0, errors.New("database is not initialized")
	}
	if retention < AliyunSMSTestWindow || maxBatches <= 0 {
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
		var ids []uint
		if err := database.Model(&AliyunSMSTestEvent{}).
			Where("created_at_ms < ?", cutoff).
			Order("created_at_ms ASC, id ASC").Limit(aliyunSMSTestBatchSize).
			Pluck("id", &ids).Error; err != nil {
			return deleted, err
		}
		if len(ids) == 0 {
			break
		}
		result := database.Where("id IN ? AND created_at_ms < ?", ids, cutoff).Delete(&AliyunSMSTestEvent{})
		if result.Error != nil {
			return deleted, result.Error
		}
		deleted += result.RowsAffected
		if len(ids) < aliyunSMSTestBatchSize {
			break
		}
	}
	return deleted, nil
}

func isRetryableAliyunRateLimitError(err error) bool {
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
