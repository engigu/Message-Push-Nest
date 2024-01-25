package send_logs_service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"message-nest/models"
	"net/url"
)

type SendTaskLogsService struct {
	ID     int
	TaskId string
	Name   string
	Query  string

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
	if len(st.Query) > 0 {
		decodedString, err := url.QueryUnescape(st.Query)
		if err != nil {
			logrus.Errorf("queryUrl编码解码失败: %s", err)
			return maps
		}
		err = json.Unmarshal([]byte(decodedString), &maps)
		if err != nil {
			logrus.Errorf("queryJson反序列化失败: %s", err)
			return maps
		}
	}
	return maps
}
