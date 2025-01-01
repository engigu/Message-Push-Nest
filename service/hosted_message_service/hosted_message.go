package hosted_message_service

import (
	"message-nest/models"
)

type HostMessageService struct {
	ID      int
	Text    string
	Title   string
	Type    string
	Content string

	PageNum  int
	PageSize int
}

func (st *HostMessageService) Add() error {
	model := models.HostedMessage{
		Title:   st.Title,
		Content: st.Content,
		Type:    st.Type,
	}
	return model.Add()
}

func (st *HostMessageService) Count() (int64, error) {
	return models.GetHostMessagesTotal(st.Text, st.getMaps())
}

func (st *HostMessageService) GetAll() ([]models.HostMessageResult, error) {
	tasks, err := models.GetHostMessages(st.PageNum, st.PageSize, st.Text, st.getMaps())
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (st *HostMessageService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	//if len(st.Query) > 0 {
	//	decodedString, err := url.QueryUnescape(st.Query)
	//	if err != nil {
	//		logrus.Errorf("queryUrl编码解码失败: %s", err)
	//		return maps
	//	}
	//	err = json.Unmarshal([]byte(decodedString), &maps)
	//	if err != nil {
	//		logrus.Errorf("queryJson反序列化失败: %s", err)
	//		return maps
	//	}
	//}
	return maps
}
