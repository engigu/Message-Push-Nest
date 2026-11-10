package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type rateLimitProcessRequest struct {
	UserHash    string `json:"user_hash"`
	WayHash     string `json:"way_hash"`
	RequestHash string `json:"request_hash"`
}

type rateLimitProcessResult struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
	Error   string `json:"error,omitempty"`
}

func TestAliyunSMSRateLimitProcessWorker(t *testing.T) {
	if os.Getenv("MN_ALIYUN_RATE_WORKER") != "1" {
		return
	}
	dbPath := os.Getenv("MN_ALIYUN_RATE_DB")
	inputPath := os.Getenv("MN_ALIYUN_RATE_INPUT")
	outputPath := os.Getenv("MN_ALIYUN_RATE_OUTPUT")
	database := openAliyunRateDBFromWorkerEnv(dbPath)
	SetDB(database)
	input, err := os.ReadFile(inputPath)
	if err != nil {
		t.Fatalf("read worker input: %v", err)
	}
	var requests []rateLimitProcessRequest
	if err := json.Unmarshal(input, &requests); err != nil {
		t.Fatalf("decode worker input: %v", err)
	}
	results := make([]rateLimitProcessResult, len(requests))
	var wg sync.WaitGroup
	for i, request := range requests {
		wg.Add(1)
		go func(i int, request rateLimitProcessRequest) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			decision, err := ReserveAliyunSMSTestRate(ctx, request.UserHash, request.WayHash, request.RequestHash)
			results[i].Allowed = decision.Allowed
			results[i].Reason = decision.Reason
			if err != nil {
				results[i].Error = err.Error()
			}
		}(i, request)
	}
	wg.Wait()
	output, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("encode worker results: %v", err)
	}
	if err := os.WriteFile(outputPath, output, 0600); err != nil {
		t.Fatalf("write worker results: %v", err)
	}
	_ = database
}

func TestAliyunSMSRateLimitSlidingWindowDuplicateAndCleanup(t *testing.T) {
	database := openAliyunRateDB(filepath.Join(t.TempDir(), "rate.db"))
	defer closeAliyunRateDB(t, database)
	SetDB(database)
	var sqliteVersion string
	if err := database.Raw("SELECT sqlite_version()").Scan(&sqliteVersion).Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("SQLite runtime version: %s", sqliteVersion)
	if err := MigrateAliyunSMSTestRateLimitTables(database); err != nil {
		t.Fatal(err)
	}
	if err := database.Migrator().CreateTable(&struct {
		ID int `gorm:"primaryKey"`
	}{}); err != nil {
		t.Fatal(err)
	}
	if err := MigrateAliyunSMSTestRateLimitTables(database); err != nil {
		t.Fatalf("dedicated migration must be repeatable: %v", err)
	}
	if !database.Migrator().HasTable(&struct {
		ID int `gorm:"primaryKey"`
	}{}) {
		t.Fatal("dedicated migration unexpectedly removed an unrelated table")
	}

	userHash, wayHash := testRateHash('a'), testRateHash('b')
	requestHash := testRateHash('c')
	decision, err := ReserveAliyunSMSTestRate(context.Background(), userHash, wayHash, requestHash)
	if err != nil || !decision.Allowed {
		t.Fatalf("first request should be allowed: %+v err=%v", decision, err)
	}
	decision, err = ReserveAliyunSMSTestRate(context.Background(), userHash, wayHash, requestHash)
	if err != nil || decision.Allowed || decision.Reason != "duplicate" || decision.RetryAfterSeconds <= 0 {
		t.Fatalf("duplicate should be rejected: %+v err=%v", decision, err)
	}
	var event AliyunSMSTestEvent
	if err := database.Where("request_hash = ?", requestHash).First(&event).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&AliyunSMSTestEvent{}).Where("id = ?", event.ID).UpdateColumn("created_at_ms", event.CreatedAtMS-AliyunSMSTestRepeat.Milliseconds()-1).Error; err != nil {
		t.Fatal(err)
	}
	decision, err = ReserveAliyunSMSTestRate(context.Background(), userHash, wayHash, requestHash)
	if err != nil || !decision.Allowed {
		t.Fatalf("same request should be allowed after repeat window while rate window remains: %+v err=%v", decision, err)
	}

	limitNow, err := databaseNowMillis(database, "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	oldUser, oldWay := testRateHash('d'), testRateHash('e')
	for i := 0; i < AliyunSMSTestUserLimit; i++ {
		if err := database.Create(&AliyunSMSTestEvent{
			UserHash: oldUser, WayHash: oldWay, RequestHash: fmt.Sprintf("%064x", i+100),
			CreatedAtMS: limitNow,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	var seededCount int64
	if err := database.Model(&AliyunSMSTestEvent{}).Where("user_hash = ? AND created_at_ms > ?", oldUser, limitNow-AliyunSMSTestWindow.Milliseconds()).Count(&seededCount).Error; err != nil || seededCount != AliyunSMSTestUserLimit {
		t.Fatalf("seeded user events=%d, want %d, now=%d err=%v", seededCount, AliyunSMSTestUserLimit, limitNow, err)
	}
	decision, err = ReserveAliyunSMSTestRate(context.Background(), oldUser, oldWay, testRateHash('f'))
	if err != nil || decision.Allowed || decision.Reason != "rate_limited" {
		t.Fatalf("user window must reject at the limit: %+v err=%v seeded=%d", decision, err, seededCount)
	}
	if err := database.Model(&AliyunSMSTestEvent{}).Where("user_hash = ?", oldUser).UpdateColumn("created_at_ms", limitNow-AliyunSMSTestWindow.Milliseconds()-1).Error; err != nil {
		t.Fatal(err)
	}
	decision, err = ReserveAliyunSMSTestRate(context.Background(), oldUser, oldWay, testRateHash('g'))
	if err != nil || !decision.Allowed {
		t.Fatalf("expired events must no longer consume quota: %+v err=%v", decision, err)
	}

	wayHashLimit := testRateHash('h')
	for i := 0; i < AliyunSMSTestWayLimit; i++ {
		if err := database.Create(&AliyunSMSTestEvent{
			UserHash: fmt.Sprintf("%064x", i+1000), WayHash: wayHashLimit,
			RequestHash: fmt.Sprintf("%064x", i+2000), CreatedAtMS: limitNow,
		}).Error; err != nil {
			t.Fatal(err)
		}
	}
	decision, err = ReserveAliyunSMSTestRate(context.Background(), testRateHash('i'), wayHashLimit, testRateHash('j'))
	if err != nil || decision.Allowed || decision.Reason != "rate_limited" {
		t.Fatalf("Way window must reject at the limit: %+v err=%v", decision, err)
	}

	stale := AliyunSMSTestEvent{UserHash: testRateHash('k'), WayHash: testRateHash('l'), RequestHash: testRateHash('m'), CreatedAtMS: limitNow - 3*time.Minute.Milliseconds()}
	recent := AliyunSMSTestEvent{UserHash: testRateHash('n'), WayHash: testRateHash('o'), RequestHash: testRateHash('p'), CreatedAtMS: limitNow}
	if err := database.Create(&[]AliyunSMSTestEvent{stale, recent}).Error; err != nil {
		t.Fatal(err)
	}
	deleted, err := CleanupAliyunSMSTestEvents(context.Background(), 2*time.Minute, 4)
	if err != nil || deleted < 1 {
		t.Fatalf("cleanup should delete stale event rows: deleted=%d err=%v", deleted, err)
	}
	var remaining int64
	if err := database.Model(&AliyunSMSTestEvent{}).Where("request_hash IN ?", []string{stale.RequestHash, recent.RequestHash}).Count(&remaining).Error; err != nil || remaining != 1 {
		t.Fatalf("cleanup should preserve recent rows; remaining=%d err=%v", remaining, err)
	}
	if !database.Migrator().HasTable(&AliyunSMSTestScopeLock{}) {
		t.Fatal("cleanup must not remove scope lock rows")
	}
}

func TestAliyunSMSRateLimitAcrossIndependentProcessesAndRestart(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "shared-rate.db")
	database := openAliyunRateDB(dbPath)
	defer closeAliyunRateDB(t, database)
	SetDB(database)
	if err := MigrateAliyunSMSTestRateLimitTables(database); err != nil {
		t.Fatal(err)
	}

	userHash, wayHash := testRateHash('q'), testRateHash('r')
	var requests [2][]rateLimitProcessRequest
	for i := 0; i < 100; i++ {
		requests[i%2] = append(requests[i%2], rateLimitProcessRequest{
			UserHash: userHash, WayHash: wayHash, RequestHash: fmt.Sprintf("%064x", i+1),
		})
	}
	results := runAliyunRateWorkers(t, dbPath, requests)
	if allowed := countAllowed(results); allowed != AliyunSMSTestUserLimit {
		t.Fatalf("100 concurrent requests across two processes: allowed=%d, want %d", allowed, AliyunSMSTestUserLimit)
	}
	var eventCount int64
	if err := database.Model(&AliyunSMSTestEvent{}).Where("user_hash = ?", userHash).Count(&eventCount).Error; err != nil || eventCount != AliyunSMSTestUserLimit {
		t.Fatalf("persisted event count=%d, want %d, err=%v", eventCount, AliyunSMSTestUserLimit, err)
	}

	sharedWay := testRateHash('s')
	requests = [2][]rateLimitProcessRequest{}
	for i := 0; i < 80; i++ {
		requests[i%2] = append(requests[i%2], rateLimitProcessRequest{
			UserHash: fmt.Sprintf("%064x", 10000+i), WayHash: sharedWay, RequestHash: fmt.Sprintf("%064x", 20000+i),
		})
	}
	results = runAliyunRateWorkers(t, dbPath, requests)
	if allowed := countAllowed(results); allowed != AliyunSMSTestWayLimit {
		t.Fatalf("80 concurrent requests across users: allowed=%d, want Way limit %d", allowed, AliyunSMSTestWayLimit)
	}

	// Requests touch overlapping sets of user and Way scopes in different hash
	// orders. They must finish without a permanent lock cycle and user quotas stay
	// atomic while Way quotas are shared by the independent processes.
	requests = [2][]rateLimitProcessRequest{}
	for userIndex := 0; userIndex < 4; userIndex++ {
		for requestIndex := 0; requestIndex < 25; requestIndex++ {
			wayIndex := (userIndex + requestIndex) % 4
			requests[(userIndex+requestIndex)%2] = append(requests[(userIndex+requestIndex)%2], rateLimitProcessRequest{
				UserHash: testRateHash(rune('A' + userIndex)), WayHash: testRateHash(rune('E' + wayIndex)),
				RequestHash: fmt.Sprintf("%064x", 40000+userIndex*25+requestIndex),
			})
		}
	}
	results = runAliyunRateWorkers(t, dbPath, requests)
	if allowed := countAllowed(results); allowed != 4*AliyunSMSTestUserLimit {
		t.Fatalf("crossed user/Way scopes allowed=%d, want %d", allowed, 4*AliyunSMSTestUserLimit)
	}

	dupUser, dupWay, dupRequest := testRateHash('t'), testRateHash('u'), testRateHash('v')
	requests = [2][]rateLimitProcessRequest{
		{{UserHash: dupUser, WayHash: dupWay, RequestHash: dupRequest}},
		{{UserHash: dupUser, WayHash: dupWay, RequestHash: dupRequest}},
	}
	results = runAliyunRateWorkers(t, dbPath, requests)
	if allowed := countAllowed(results); allowed != 1 {
		t.Fatalf("same request across independent processes: allowed=%d, want 1", allowed)
	}
	if reasons := countReason(results, "duplicate"); reasons != 1 {
		t.Fatalf("same request should be recognized as duplicate once, results=%+v", results)
	}

	// The next worker pair are fresh processes and independent pools. Existing 10
	// events must still block the 11th request after the previous processes exit.
	restartUser, restartWay := testRateHash('w'), testRateHash('x')
	firstBatch := [2][]rateLimitProcessRequest{}
	for i := 0; i < AliyunSMSTestUserLimit; i++ {
		firstBatch[i%2] = append(firstBatch[i%2], rateLimitProcessRequest{
			UserHash: restartUser, WayHash: restartWay, RequestHash: fmt.Sprintf("%064x", 30000+i),
		})
	}
	if allowed := countAllowed(runAliyunRateWorkers(t, dbPath, firstBatch)); allowed != AliyunSMSTestUserLimit {
		t.Fatalf("pre-restart sends allowed=%d", allowed)
	}
	postRestart := [2][]rateLimitProcessRequest{{{UserHash: restartUser, WayHash: restartWay, RequestHash: testRateHash('y')}}, nil}
	if results := runAliyunRateWorkers(t, dbPath, postRestart); countAllowed(results) != 0 || countReason(results, "rate_limited") != 1 {
		t.Fatalf("process restart must not reset the user window: %+v", results)
	}

	repeatUser, repeatWay, repeatRequest := testRateHash('z'), testRateHash('1'), testRateHash('2')
	if allowed := countAllowed(runAliyunRateWorkers(t, dbPath, [2][]rateLimitProcessRequest{{{UserHash: repeatUser, WayHash: repeatWay, RequestHash: repeatRequest}}, nil})); allowed != 1 {
		t.Fatal("pre-restart duplicate seed should be allowed")
	}
	results = runAliyunRateWorkers(t, dbPath, [2][]rateLimitProcessRequest{{{UserHash: repeatUser, WayHash: repeatWay, RequestHash: repeatRequest}}, nil})
	if countAllowed(results) != 0 || countReason(results, "duplicate") != 1 {
		t.Fatalf("process restart must preserve duplicate protection: %+v", results)
	}
}

func openAliyunRateDB(path string) *gorm.DB {
	return openAliyunRateGormDB(sqlite.Open(path+"?_pragma=busy_timeout(5000)"), "message_")
}

func openAliyunRateDBFromWorkerEnv(path string) *gorm.DB {
	prefix := os.Getenv("MN_ALIYUN_RATE_TABLE_PREFIX")
	if prefix == "" {
		prefix = "message_"
	}
	if os.Getenv("MN_ALIYUN_RATE_DIALECT") == "mysql" {
		return openAliyunRateGormDB(mysql.Open(os.Getenv("MN_ALIYUN_RATE_DSN")), prefix)
	}
	return openAliyunRateGormDB(sqlite.Open(path+"?_pragma=busy_timeout(5000)"), prefix)
}

func openAliyunRateGormDB(dialector gorm.Dialector, prefix string) *gorm.DB {
	database, err := gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: prefix, SingularTable: true},
	})
	if err != nil {
		panic(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(20)
	return database
}

func closeAliyunRateDB(t *testing.T, database *gorm.DB) {
	t.Helper()
	sqlDB, err := database.DB()
	if err != nil {
		t.Error(err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		t.Error(err)
	}
}

func runAliyunRateWorkers(t *testing.T, dbPath string, requests [2][]rateLimitProcessRequest) []rateLimitProcessResult {
	return runAliyunRateWorkersWithDB(t, requests, "sqlite", dbPath, "", "message_")
}

func runAliyunRateWorkersWithDB(t *testing.T, requests [2][]rateLimitProcessRequest, dialect, dbPath, dsn, tablePrefix string) []rateLimitProcessResult {
	t.Helper()
	results := make(chan []rateLimitProcessResult, len(requests))
	errCh := make(chan error, len(requests))
	for workerIndex := range requests {
		workerDir := filepath.Join(t.TempDir(), fmt.Sprintf("worker-%d", workerIndex))
		if err := os.MkdirAll(workerDir, 0700); err != nil {
			t.Fatal(err)
		}
		inputPath := filepath.Join(workerDir, "input.json")
		outputPath := filepath.Join(workerDir, "output.json")
		input, err := json.Marshal(requests[workerIndex])
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(inputPath, input, 0600); err != nil {
			t.Fatal(err)
		}
		go func(inputPath, outputPath string) {
			command := exec.Command(os.Args[0], "-test.run=^TestAliyunSMSRateLimitProcessWorker$")
			command.Env = append(os.Environ(),
				"MN_ALIYUN_RATE_WORKER=1",
				"MN_ALIYUN_RATE_DB="+dbPath,
				"MN_ALIYUN_RATE_DIALECT="+dialect,
				"MN_ALIYUN_RATE_DSN="+dsn,
				"MN_ALIYUN_RATE_TABLE_PREFIX="+tablePrefix,
				"MN_ALIYUN_RATE_INPUT="+inputPath,
				"MN_ALIYUN_RATE_OUTPUT="+outputPath,
			)
			command.Stdout = io.Discard
			command.Stderr = io.Discard
			if err := command.Run(); err != nil {
				errCh <- err
				return
			}
			content, err := os.ReadFile(outputPath)
			if err != nil {
				errCh <- err
				return
			}
			var workerResults []rateLimitProcessResult
			if err := json.Unmarshal(content, &workerResults); err != nil {
				errCh <- err
				return
			}
			results <- workerResults
		}(inputPath, outputPath)
	}
	var combined []rateLimitProcessResult
	for i := 0; i < len(requests); i++ {
		select {
		case workerResults := <-results:
			combined = append(combined, workerResults...)
		case err := <-errCh:
			t.Fatalf("rate-limit worker process failed: %v", err)
		case <-time.After(45 * time.Second):
			t.Fatal("rate-limit worker process timed out")
		}
	}
	for _, result := range combined {
		if result.Error != "" {
			t.Fatalf("rate-limit worker returned DB error: %s", result.Error)
		}
	}
	return combined
}

func countAllowed(results []rateLimitProcessResult) int {
	count := 0
	for _, result := range results {
		if result.Allowed {
			count++
		}
	}
	return count
}

func countReason(results []rateLimitProcessResult, reason string) int {
	count := 0
	for _, result := range results {
		if result.Reason == reason {
			count++
		}
	}
	return count
}

func TestAliyunSMSRateLimitExternalDatabaseIntegration(t *testing.T) {
	dialect := os.Getenv("MN_ALIYUN_EXTERNAL_DIALECT")
	dsn := os.Getenv("MN_ALIYUN_EXTERNAL_DSN")
	if dialect == "" || dsn == "" {
		t.Skip("set MN_ALIYUN_EXTERNAL_DIALECT=mysql and MN_ALIYUN_EXTERNAL_DSN to run against a real isolated MySQL-compatible database")
	}
	if dialect != "mysql" {
		t.Fatalf("unsupported external integration dialect %q", dialect)
	}

	prefix := "v" + fmt.Sprintf("%x", time.Now().UnixNano()%0xffffff) + "_"
	database := openAliyunRateGormDB(mysql.Open(dsn), prefix)
	SetDB(database)
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(32)
	t.Cleanup(func() { _ = sqlDB.Close() })

	var server struct {
		Version   string
		Comment   string
		Isolation string
		Port      int
	}
	if err := database.Raw("SELECT VERSION() AS version, @@version_comment AS comment, @@transaction_isolation AS isolation, @@port AS port").Scan(&server).Error; err != nil {
		t.Fatal(err)
	}
	t.Logf("database=%s comment=%s session_isolation=%s port=%d", server.Version, server.Comment, server.Isolation, server.Port)

	if err := MigrateAliyunSMSTestRateLimitTables(database); err != nil {
		t.Fatalf("first limiter migration: %v", err)
	}
	if err := MigrateAliyunSMSTestRateLimitTables(database); err != nil {
		t.Fatalf("repeat limiter migration: %v", err)
	}
	for _, model := range []any{&AliyunSMSTestScopeLock{}, &AliyunSMSTestEvent{}} {
		statement := &gorm.Statement{DB: database}
		if err := statement.Parse(model); err != nil {
			t.Fatal(err)
		}
		var engine string
		if err := database.Raw("SELECT ENGINE FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", statement.Table).Scan(&engine).Error; err != nil {
			t.Fatal(err)
		}
		if !strings.EqualFold(engine, "InnoDB") {
			t.Fatalf("table %s engine=%q, want InnoDB", statement.Table, engine)
		}
	}
	for _, index := range []string{"idx_alisms_evt_user_time", "idx_alisms_evt_way_time", "idx_alisms_evt_request_time", "idx_alisms_evt_created_id"} {
		if !database.Migrator().HasIndex(&AliyunSMSTestEvent{}, index) {
			t.Fatalf("expected event index %s", index)
		}
	}
	dbNow, err := databaseNowMillis(database, "mysql")
	if err != nil {
		t.Fatalf("database time query: %v", err)
	}
	clockDelta := time.Duration(dbNow-time.Now().UnixMilli()) * time.Millisecond
	if clockDelta < -5*time.Second || clockDelta > 5*time.Second {
		t.Fatalf("database epoch time differs from local UTC clock by %s", clockDelta)
	}

	// Hold one real InnoDB scope row lock in a separate connection and prove a
	// reservation waits for it, then proceeds after the blocker commits.
	lockedUser, lockedWay := testRateHash('0'), testRateHash('1')
	if err := database.Create(&[]AliyunSMSTestScopeLock{
		{ScopeHash: lockedUser, ScopeType: "user"},
		{ScopeHash: lockedWay, ScopeType: "way"},
	}).Error; err != nil {
		t.Fatal(err)
	}
	lockTx, err := sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	lockStatement := &gorm.Statement{DB: database}
	if err := lockStatement.Parse(&AliyunSMSTestScopeLock{}); err != nil {
		_ = lockTx.Rollback()
		t.Fatal(err)
	}
	if _, err := lockTx.Exec("UPDATE `"+lockStatement.Table+"` SET touched_at_ms = ? WHERE scope_hash = ?", dbNow, lockedUser); err != nil {
		_ = lockTx.Rollback()
		t.Fatalf("acquire InnoDB scope row lock: %v", err)
	}
	type reserveResult struct {
		decision AliyunSMSTestLimitDecision
		err      error
	}
	reserveDone := make(chan reserveResult, 1)
	reserveCtx, cancelReserve := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancelReserve()
	go func() {
		decision, err := ReserveAliyunSMSTestRate(reserveCtx, lockedUser, lockedWay, testRateHash('2'))
		reserveDone <- reserveResult{decision: decision, err: err}
	}()
	select {
	case result := <-reserveDone:
		_ = lockTx.Rollback()
		t.Fatalf("reservation should wait for the held InnoDB scope row lock, got %+v err=%v", result.decision, result.err)
	case <-time.After(250 * time.Millisecond):
	}
	if err := lockTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	select {
	case result := <-reserveDone:
		if result.err != nil || !result.decision.Allowed {
			t.Fatalf("reservation should continue after row lock release: %+v err=%v", result.decision, result.err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("reservation did not finish within its bounded context after row lock release")
	}
	t.Log("InnoDB scope row lock wait blocked and completed after lock release")

	// Induce a native two-transaction deadlock on disposable scope rows and
	// verify the database driver's error is recognized by the bounded retry path.
	deadlockScopeA, deadlockScopeB := testRateHash('3'), testRateHash('4')
	if err := database.Create(&[]AliyunSMSTestScopeLock{
		{ScopeHash: deadlockScopeA, ScopeType: "user"},
		{ScopeHash: deadlockScopeB, ScopeType: "way"},
	}).Error; err != nil {
		t.Fatal(err)
	}
	deadlockTx1, err := sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	deadlockTx2, err := sqlDB.BeginTx(context.Background(), nil)
	if err != nil {
		_ = deadlockTx1.Rollback()
		t.Fatal(err)
	}
	deadlockUpdate := "UPDATE `" + lockStatement.Table + "` SET touched_at_ms = touched_at_ms + 1 WHERE scope_hash = ?"
	if _, err := deadlockTx1.Exec(deadlockUpdate, deadlockScopeA); err != nil {
		_ = deadlockTx1.Rollback()
		_ = deadlockTx2.Rollback()
		t.Fatal(err)
	}
	if _, err := deadlockTx2.Exec(deadlockUpdate, deadlockScopeB); err != nil {
		_ = deadlockTx1.Rollback()
		_ = deadlockTx2.Rollback()
		t.Fatal(err)
	}
	deadlockWaiting := make(chan error, 1)
	go func() {
		_, err := deadlockTx1.Exec(deadlockUpdate, deadlockScopeB)
		deadlockWaiting <- err
	}()
	time.Sleep(200 * time.Millisecond)
	_, deadlockErr := deadlockTx2.Exec(deadlockUpdate, deadlockScopeA)
	if deadlockErr == nil || !isRetryableAliyunRateLimitError(deadlockErr) {
		_ = deadlockTx1.Rollback()
		_ = deadlockTx2.Rollback()
		t.Fatalf("database deadlock error must be recognized as retryable, err=%v", deadlockErr)
	}
	if err := deadlockTx2.Rollback(); err != nil {
		_ = deadlockTx1.Rollback()
		t.Fatal(err)
	}
	select {
	case err := <-deadlockWaiting:
		if err != nil {
			_ = deadlockTx1.Rollback()
			t.Fatalf("surviving transaction should resume after deadlock victim rolls back: %v", err)
		}
	case <-time.After(8 * time.Second):
		_ = deadlockTx1.Rollback()
		t.Fatal("surviving deadlock transaction did not resume within bounded time")
	}
	if err := deadlockTx1.Commit(); err != nil {
		t.Fatal(err)
	}
	t.Logf("database native deadlock recognized by bounded retry classifier: %T", deadlockErr)

	// Prove effective READ COMMITTED behavior: a transaction's second read sees
	// an event committed after its first read. @@transaction_isolation reports
	// the session default, not the one-shot isolation selected by BeginTx.
	userHashForIsolation := testRateHash('0')
	wayHashForIsolation := testRateHash('1')
	requestHashForIsolation := testRateHash('2')
	directTx, err := sqlDB.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		t.Fatalf("direct READ COMMITTED transaction: %v", err)
	}
	var beforeEvents int64
	if err := directTx.QueryRow("SELECT COUNT(*) FROM `" + prefix + "aliyun_sms_test_event`").Scan(&beforeEvents); err != nil {
		t.Fatal(err)
	}
	isolationEvent := AliyunSMSTestEvent{UserHash: userHashForIsolation, WayHash: wayHashForIsolation, RequestHash: requestHashForIsolation, CreatedAtMS: dbNow}
	if err := database.Create(&isolationEvent).Error; err != nil {
		_ = directTx.Rollback()
		t.Fatal(err)
	}
	var afterEvents int64
	if err := directTx.QueryRow("SELECT COUNT(*) FROM `" + prefix + "aliyun_sms_test_event`").Scan(&afterEvents); err != nil {
		_ = directTx.Rollback()
		t.Fatal(err)
	}
	if err := directTx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if afterEvents != beforeEvents+1 {
		t.Fatalf("READ COMMITTED second read saw %d rows after first saw %d; newly committed event must be visible", afterEvents, beforeEvents)
	}
	if err := database.Delete(&isolationEvent).Error; err != nil {
		t.Fatal(err)
	}
	t.Log("READ COMMITTED verified by committed-row visibility between reads in one transaction")

	userHash, wayHash := testRateHash('a'), testRateHash('b')
	for i := 0; i < AliyunSMSTestUserLimit+1; i++ {
		decision, err := ReserveAliyunSMSTestRate(context.Background(), userHash, wayHash, fmt.Sprintf("%064x", 50000+i))
		if err != nil {
			t.Fatalf("user reservation %d: %v", i+1, err)
		}
		if decision.Allowed != (i < AliyunSMSTestUserLimit) {
			t.Fatalf("user reservation %d decision=%+v", i+1, decision)
		}
	}

	wayLimitHash := testRateHash('c')
	for i := 0; i < AliyunSMSTestWayLimit+1; i++ {
		decision, err := ReserveAliyunSMSTestRate(context.Background(), fmt.Sprintf("%064x", 60000+i), wayLimitHash, fmt.Sprintf("%064x", 70000+i))
		if err != nil {
			t.Fatalf("Way reservation %d: %v", i+1, err)
		}
		if decision.Allowed != (i < AliyunSMSTestWayLimit) {
			t.Fatalf("Way reservation %d decision=%+v", i+1, decision)
		}
	}

	dupUser, dupWay, dupRequest := testRateHash('d'), testRateHash('e'), testRateHash('f')
	first, err := ReserveAliyunSMSTestRate(context.Background(), dupUser, dupWay, dupRequest)
	if err != nil || !first.Allowed {
		t.Fatalf("first duplicate-window request=%+v err=%v", first, err)
	}
	second, err := ReserveAliyunSMSTestRate(context.Background(), dupUser, dupWay, dupRequest)
	if err != nil || second.Allowed || second.Reason != "duplicate" {
		t.Fatalf("repeat request was not rejected: %+v err=%v", second, err)
	}

	// Cover sliding-window expiration without sleeping for a full minute.
	expiredUser, expiredWay, expiredRequest := testRateHash('g'), testRateHash('h'), testRateHash('i')
	if err := database.Create(&AliyunSMSTestEvent{UserHash: expiredUser, WayHash: expiredWay, RequestHash: expiredRequest, CreatedAtMS: dbNow - AliyunSMSTestWindow.Milliseconds() - 1}).Error; err != nil {
		t.Fatal(err)
	}
	expired, err := ReserveAliyunSMSTestRate(context.Background(), expiredUser, expiredWay, testRateHash('j'))
	if err != nil || !expired.Allowed {
		t.Fatalf("expired sliding-window event should not consume quota: %+v err=%v", expired, err)
	}

	// Real independent processes, each with an independent sql.DB pool.
	var batches [2][]rateLimitProcessRequest
	processUser, processWay := testRateHash('k'), testRateHash('l')
	for i := 0; i < 100; i++ {
		batches[i%2] = append(batches[i%2], rateLimitProcessRequest{UserHash: processUser, WayHash: processWay, RequestHash: fmt.Sprintf("%064x", 80000+i)})
	}
	results := runAliyunRateWorkersWithDB(t, batches, "mysql", "", dsn, prefix)
	if countAllowed(results) != AliyunSMSTestUserLimit {
		t.Fatalf("100 cross-process requests for one user allowed=%d, want %d", countAllowed(results), AliyunSMSTestUserLimit)
	}

	batches = [2][]rateLimitProcessRequest{}
	processWay = testRateHash('m')
	for i := 0; i < 80; i++ {
		batches[i%2] = append(batches[i%2], rateLimitProcessRequest{UserHash: fmt.Sprintf("%064x", 90000+i), WayHash: processWay, RequestHash: fmt.Sprintf("%064x", 100000+i)})
	}
	results = runAliyunRateWorkersWithDB(t, batches, "mysql", "", dsn, prefix)
	if countAllowed(results) != AliyunSMSTestWayLimit {
		t.Fatalf("80 cross-process users for one Way allowed=%d, want %d", countAllowed(results), AliyunSMSTestWayLimit)
	}

	batches = [2][]rateLimitProcessRequest{}
	for userIndex := 0; userIndex < 4; userIndex++ {
		for requestIndex := 0; requestIndex < 25; requestIndex++ {
			wayIndex := (userIndex + requestIndex) % 4
			batches[(userIndex+requestIndex)%2] = append(batches[(userIndex+requestIndex)%2], rateLimitProcessRequest{
				// Lowercase hexadecimal avoids case-insensitive CHAR collation aliases.
				UserHash: testRateHash(rune('3' + userIndex)), WayHash: testRateHash(rune('7' + wayIndex)),
				RequestHash: fmt.Sprintf("%064x", 110000+userIndex*25+requestIndex),
			})
		}
	}
	results = runAliyunRateWorkersWithDB(t, batches, "mysql", "", dsn, prefix)
	if countAllowed(results) != 4*AliyunSMSTestUserLimit {
		t.Fatalf("cross-scope requests allowed=%d, want %d", countAllowed(results), 4*AliyunSMSTestUserLimit)
	}

	duplicateUser, duplicateWay, duplicateRequest := testRateHash('n'), testRateHash('o'), testRateHash('p')
	batches = [2][]rateLimitProcessRequest{{{UserHash: duplicateUser, WayHash: duplicateWay, RequestHash: duplicateRequest}}, {{UserHash: duplicateUser, WayHash: duplicateWay, RequestHash: duplicateRequest}}}
	results = runAliyunRateWorkersWithDB(t, batches, "mysql", "", dsn, prefix)
	if countAllowed(results) != 1 || countReason(results, "duplicate") != 1 {
		t.Fatalf("cross-process duplicate protection failed: %+v", results)
	}

	restartUser, restartWay := testRateHash('q'), testRateHash('r')
	batches = [2][]rateLimitProcessRequest{}
	for i := 0; i < AliyunSMSTestUserLimit; i++ {
		batches[i%2] = append(batches[i%2], rateLimitProcessRequest{UserHash: restartUser, WayHash: restartWay, RequestHash: fmt.Sprintf("%064x", 120000+i)})
	}
	if got := countAllowed(runAliyunRateWorkersWithDB(t, batches, "mysql", "", dsn, prefix)); got != AliyunSMSTestUserLimit {
		t.Fatalf("pre-restart reservations=%d, want %d", got, AliyunSMSTestUserLimit)
	}
	postRestart := [2][]rateLimitProcessRequest{{{UserHash: restartUser, WayHash: restartWay, RequestHash: testRateHash('s')}}, nil}
	results = runAliyunRateWorkersWithDB(t, postRestart, "mysql", "", dsn, prefix)
	if countAllowed(results) != 0 || countReason(results, "rate_limited") != 1 {
		t.Fatalf("separate process must observe persisted quota: %+v", results)
	}
	postRestart = [2][]rateLimitProcessRequest{{{UserHash: duplicateUser, WayHash: duplicateWay, RequestHash: duplicateRequest}}, nil}
	results = runAliyunRateWorkersWithDB(t, postRestart, "mysql", "", dsn, prefix)
	if countAllowed(results) != 0 || countReason(results, "duplicate") != 1 {
		t.Fatalf("separate process must observe persisted duplicate key: %+v", results)
	}

	var stored []AliyunSMSTestEvent
	if err := database.Find(&stored).Error; err != nil {
		t.Fatal(err)
	}
	serialized, _ := json.Marshal(stored)
	for _, sensitive := range []string{"155555", "ACCESS_KEY", "TemplateParam", "Config", "verification-code"} {
		if strings.Contains(string(serialized), sensitive) {
			t.Fatalf("rate events contain sensitive payload text %q", sensitive)
		}
	}
	t.Logf("event rows after real DB integration: %d; no payload columns are present", len(stored))
}

func TestAliyunSMSRateLimitRecognizesBoundedRetryErrors(t *testing.T) {
	for _, err := range []error{
		fmt.Errorf("Error 1213 (40001): Deadlock found when trying to get lock"),
		fmt.Errorf("Error 1205 (HY000): Lock wait timeout exceeded"),
		fmt.Errorf("Error 9007: Write conflict"),
		aliyunRateLimitSQLStateTestError{state: "40001", message: "serialization failure"},
		aliyunRateLimitSQLStateTestError{state: "40P01", message: "deadlock detected"},
	} {
		if !isRetryableAliyunRateLimitError(err) {
			t.Errorf("expected retryable classification for %q", err)
		}
	}
	for _, message := range []string{"permission denied", "unknown column", "connection refused"} {
		if isRetryableAliyunRateLimitError(fmt.Errorf("%s", message)) {
			t.Errorf("unexpected retryable classification for %q", message)
		}
	}
}

type aliyunRateLimitSQLStateTestError struct {
	state   string
	message string
}

func (err aliyunRateLimitSQLStateTestError) Error() string    { return err.message }
func (err aliyunRateLimitSQLStateTestError) SQLState() string { return err.state }

func testRateHash(character rune) string {
	return strings.Repeat(string(character), 64)
}
