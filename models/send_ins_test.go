package models

import (
	"fmt"
	"testing"
)

// TestManyAddTaskInsBatchAndDedup 测试批量插入的排重与批量写入正确性
func TestManyAddTaskInsBatchAndDedup(t *testing.T) {
	InitTestDB("many_add_task_ins_test")

	// 1. 准备首批实例数据 (20条)
	var firstBatch []SendTasksIns
	for i := 1; i <= 20; i++ {
		firstBatch = append(firstBatch, SendTasksIns{
			UUIDModel: UUIDModel{
				ID:        fmt.Sprintf("IN_BATCH_%03d", i),
				CreatedBy: "test_user",
			},
			TaskID:  "TK_BATCH_01",
			WayID:   "WY_TEST_01",
			WayType: "webhook",
		})
	}

	err := ManyAddTaskIns(firstBatch)
	if err != nil {
		t.Fatalf("ManyAddTaskIns first batch failed: %v", err)
	}

	// 验证全部 20 条已写入
	var count int64
	db.Model(&SendTasksIns{}).Where("task_id = ?", "TK_BATCH_01").Count(&count)
	if count != 20 {
		t.Fatalf("Expected 20 records, got %d", count)
	}

	// 2. 准备第二批实例数据（包含 10 条已存在的老数据 + 15 条全新数据，总共 25 条）
	var mixedBatch []SendTasksIns
	for i := 11; i <= 35; i++ {
		mixedBatch = append(mixedBatch, SendTasksIns{
			UUIDModel: UUIDModel{
				ID:        fmt.Sprintf("IN_BATCH_%03d", i),
				CreatedBy: "test_user",
			},
			TaskID:  "TK_BATCH_01",
			WayID:   "WY_TEST_01",
			WayType: "webhook",
		})
	}

	err = ManyAddTaskIns(mixedBatch)
	if err != nil {
		t.Fatalf("ManyAddTaskIns mixed batch failed: %v", err)
	}

	// 验证总数应精确为 35 条（原 20 条 + 新 15 条，已存在的 10 条被精准排重忽略）
	db.Model(&SendTasksIns{}).Where("task_id = ?", "TK_BATCH_01").Count(&count)
	if count != 35 {
		t.Fatalf("Expected 35 records after deduplicated batch insert, got %d", count)
	}
}
