package send_logs_service

import (
	"message-nest/models"
)

type SendTaskLogsService struct {
	ID     int
	TaskId string
	Name   string

	PageNum  int
	PageSize int
}

func (st *SendTaskLogsService) Count() (int, error) {
	return models.GetSendLogsTotal(st.Name, st.TaskId, st.getMaps())
}

func (st *SendTaskLogsService) GetAll() ([]models.LogsResult, error) {
	tasks, err := models.GetSendLogs(st.PageNum, st.PageSize, st.Name, st.TaskId, st.getMaps())
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (st *SendTaskLogsService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	return maps
}
