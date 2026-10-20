package send_task_service

import (
	"message-nest/models"
)

type SendTaskService struct {
	ID         string
	Name       string
	CreatedBy  string
	ModifiedBy string
	CreatedOn  string

	IsAdmin  bool
	PageNum  int
	PageSize int
}

func (st *SendTaskService) Add() error {
	return models.AddSendTask(st.Name, st.CreatedBy)
}

func (st *SendTaskService) AddWithID() error {
	return models.AddSendTaskWithID(st.Name, st.ID, st.CreatedBy)
}

func (st *SendTaskService) Delete() error {
	tenantID := ""
	if !st.IsAdmin && st.ModifiedBy != "" {
		tenantID = st.ModifiedBy
	}
	return models.DeleteMsgTask(st.ID, tenantID)
}

func (st *SendTaskService) Edit(data map[string]interface{}) error {
	data["modified_by"] = st.ModifiedBy
	tenantID := ""
	if !st.IsAdmin && st.ModifiedBy != "" {
		tenantID = st.ModifiedBy
	}
	return models.EditSendTask(st.ID, data, tenantID)
}

func (st *SendTaskService) GetTaskWithIns() (models.TaskIns, error) {
	return models.GetTasksIns(st.ID)
}

func (st *SendTaskService) Count() (int64, error) {
	return models.GetSendTasksTotal(st.Name, models.TenantMap(st.IsAdmin, st.CreatedBy))
}

func (st *SendTaskService) GetAll() ([]models.SendTasks, error) {
	tasks, err := models.GetSendTasks(st.PageNum, st.PageSize, st.Name, models.TenantMap(st.IsAdmin, st.CreatedBy))
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (st *SendTaskService) GetByID() (interface{}, error) {
	return models.GetTaskByID(st.ID)
}
