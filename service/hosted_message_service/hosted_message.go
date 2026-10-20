package hosted_message_service

import (
	"message-nest/models"
)

type HostMessageService struct {
	ID        int
	Text      string
	Title     string
	Type      string
	Content   string
	UniqueKey string

	CreatedBy  string
	ModifiedBy string
	IsAdmin    bool

	PageNum  int
	PageSize int
}

func (st *HostMessageService) Add() error {
	model := models.HostedMessage{
		Title:      st.Title,
		Content:    st.Content,
		Type:       st.Type,
		UniqueKey:  st.UniqueKey,
		CreatedBy:  st.CreatedBy,
		ModifiedBy: st.ModifiedBy,
	}
	return model.Add()
}

func (st *HostMessageService) Count() (int64, error) {
	return models.GetHostMessagesTotal(st.Text, models.TenantMap(st.IsAdmin, st.CreatedBy))
}

func (st *HostMessageService) GetAll() ([]models.HostMessageResult, error) {
	tasks, err := models.GetHostMessages(st.PageNum, st.PageSize, st.Text, models.TenantMap(st.IsAdmin, st.CreatedBy))
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
