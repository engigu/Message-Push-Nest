package cron_msg_service

import (
	"message-nest/models"
)

type CronMsgService struct {
	ID string

	Name    string
	TaskID  string
	Cron    string
	Title   string
	Content string
	Url     string

	CreatedBy  string
	ModifiedBy string
	CreatedOn  string

	PageNum  int
	PageSize int
}

func (st *CronMsgService) Add() (string, error) {
	return models.AddSendCronMsg(st.Name, st.TaskID, st.Cron, st.Title, st.Content, st.Url, st.CreatedBy)
}

func (st *CronMsgService) Edit(data map[string]interface{}) error {
	return models.EditCronMsg(st.ID, data)
}

func (st *CronMsgService) GetByID() (models.CronMessages, error) {
	return models.GetCronMsgByID(st.ID)
}

func (st *CronMsgService) Count() (int, error) {
	return models.GetCronMessagesTotal(st.Name, st.getMaps())
}

func (st *CronMsgService) GetAll() ([]models.CronMessages, error) {
	msgs, err := models.GetCronMessages(st.PageNum, st.PageSize, st.Name, st.getMaps())
	if err != nil {
		return nil, err
	}
	return msgs, nil
}

func (st *CronMsgService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	return maps
}

func (st *CronMsgService) Delete() error {
	return models.DeleteCronMsg(st.ID)
}
