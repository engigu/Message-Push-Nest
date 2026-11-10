package models

import (
	"context"
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
	database := openAliyunRateDB(dbPath)
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
	database, err := gorm.Open(sqlite.Open(path+"?_pragma=busy_timeout(5000)"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: "message_", SingularTable: true},
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

func testRateHash(character rune) string {
	return strings.Repeat(string(character), 64)
}
