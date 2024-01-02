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
	return models.DeleteMsgTask(st.ID)
}

func (st *SendTaskService) Edit(data map[string]string) error {
	return models.EditSendTask(st.ID, data)
}

func (st *SendTaskService) GetTaskWithIns() (models.TaskIns, error) {
	return models.GetTasksIns(st.ID)
}

func (st *SendTaskService) Count() (int, error) {
	return models.GetSendTasksTotal(st.Name, st.getMaps())
}

func (st *SendTaskService) GetAll() ([]models.SendTasks, error) {
	tasks, err := models.GetSendTasks(st.PageNum, st.PageSize, st.Name, st.getMaps())
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (st *SendTaskService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	return maps
}
