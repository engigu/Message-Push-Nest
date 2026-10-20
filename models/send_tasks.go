package models

import (
	"fmt"
	"gorm.io/gorm"
	"message-nest/pkg/util"
)

type SendTasks struct {
	UUIDModel

	Name string `json:"name" gorm:"type:varchar(100);default:'';index:idx_tasks_name"`
}

func GenerateTaskUniqueID() string {
	newUUID := util.GenerateUniqueID()
	return fmt.Sprintf("TK%s", newUUID)
}

// AddSendTaskWithID 添加实例的时候添加任务
func AddSendTaskWithID(name string, id string, createdBy string) error {
	var etask SendTasks
	db.Where("id = ?", id).Limit(1).Find(&etask)
	if len(etask.ID) > 0 {
		return nil
	}
	task := SendTasks{
		UUIDModel: UUIDModel{
			ID:         id,
			CreatedBy:  createdBy,
			ModifiedBy: createdBy,
		},
		Name: name,
	}
	if err := db.Create(&task).Error; err != nil {
		return err
	}
	return nil
}

// AddSendTask 添加任务
func AddSendTask(name string, createdBy string) error {
	newUUID := GenerateTaskUniqueID()
	task := SendTasks{
		UUIDModel: UUIDModel{
			ID:         newUUID,
			CreatedBy:  createdBy,
			ModifiedBy: createdBy,
		},
		Name: name,
	}
	if err := db.Create(&task).Error; err != nil {
		return err
	}
	return nil
}

// GetSendTasks 获取所有任务
func GetSendTasks(pageNum int, pageSize int, name string, maps interface{}) ([]SendTasks, error) {
	var (
		tasks []SendTasks
		err   error
	)
	query := db.Where(maps)
	if name != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", name))
	}
	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}
	err = query.Find(&tasks).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return tasks, nil
}

// GetSendTasksTotal 获取所有任务总数
func GetSendTasksTotal(name string, maps interface{}) (int64, error) {
	var (
		err   error
		total int64
	)
	query := db.Model(&SendTasks{}).Where(maps)
	if name != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", name))
	}

	err = query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

type SendTasksInsRes struct {
	SendTasksIns

	WayName string `json:"way_name"`
}

type TaskIns struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	CreatedBy string            `json:"created_by"`
	InsData   []SendTasksInsRes `json:"ins_data"`
}

// GetTasksIns 获取所有任务下所有的实例
func GetTasksIns(id string) (TaskIns, error) {
	//insTable := GetSchema(SendTasksIns{})
	insTable := GetSchema(SendTasksIns{})
	waysTable := GetSchema(SendWays{})
	var (
		task       SendTasks
		taskIns    []SendTasksInsRes
		taskResult TaskIns
	)
	err := db.Where("id = ?", id).Limit(1).Find(&task).Error
	if err != nil {
		return taskResult, err
	}

	db.
		Table(insTable).
		Select(fmt.Sprintf("%s.*, %s.name as way_name", insTable, waysTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.way_id = %s.id", waysTable, insTable, waysTable)).
		Where(fmt.Sprintf("%s.task_id = ?", insTable), id).
		Order(fmt.Sprintf("%s.created_on DESC", insTable)).
		Scan(&taskIns)

	taskResult.ID = task.ID
	taskResult.Name = task.Name
	taskResult.CreatedBy = task.CreatedBy
	taskResult.InsData = taskIns
	return taskResult, nil
}

// FindTaskByWayId 通过way_id找到关联的任务
func FindTaskByWayId(wayId string) []SendTasks {
	insTable := GetSchema(SendTasksIns{})
	taskTable := GetSchema(SendTasks{})
	var (
		tasks []SendTasks
	)

	db.
		Table(taskTable).
		Select(fmt.Sprintf("%s.*", taskTable)).
		Joins(fmt.Sprintf("JOIN %s ON %s.task_id = %s.id", insTable, insTable, taskTable)).
		Where(fmt.Sprintf("%s.way_id = ?", insTable), wayId).
		Scan(&tasks)

	return tasks
}

// DeleteMsgTask 删除任务并级联删除所有关联的实例（行级原子校验租户防越权）
func DeleteMsgTask(id string, createdBy ...string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		delQuery := tx.Where("id = ?", id)
		if len(createdBy) > 0 && createdBy[0] != "" {
			delQuery = delQuery.Where("created_by = ?", createdBy[0])
		}
		result := delQuery.Delete(&SendTasks{})
		if result.Error != nil {
			return result.Error
		}
		if len(createdBy) > 0 && createdBy[0] != "" && result.RowsAffected == 0 {
			return ErrNoPermissionOrNotFound
		}
		return tx.Where("task_id = ?", id).Delete(&SendTasksIns{}).Error
	})
}

// EditSendTask 原子更新任务（行级原子校验租户防越权）
func EditSendTask(id string, data map[string]interface{}, createdBy ...string) error {
	query := db.Model(&SendTasks{}).Where("id = ?", id)
	if len(createdBy) > 0 && createdBy[0] != "" {
		query = query.Where("created_by = ?", createdBy[0])
	}
	result := query.Updates(data)
	if result.Error != nil {
		return result.Error
	}
	if len(createdBy) > 0 && createdBy[0] != "" && result.RowsAffected == 0 {
		return ErrNoPermissionOrNotFound
	}

	return nil
}

func GetTaskByID(id string) (SendTasks, error) {
	var task SendTasks
	err := db.Where("id = ? ", id).Limit(1).Find(&task).Error
	if err != nil {
		return task, err
	}
	if task.ID == "" {
		return task, gorm.ErrRecordNotFound
	}
	return task, nil
}
