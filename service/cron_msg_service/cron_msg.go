package cron_msg_service

import (
	"github.com/robfig/cron/v3"
	"message-nest/models"
	"time"
)

type CronMsgResult struct {
	models.CronMessages

	NextTime string `json:"next_time"`
}

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

	IsAdmin bool

	PageNum  int
	PageSize int
}

func (st *CronMsgService) Add() (string, error) {
	return models.AddSendCronMsg(st.Name, st.TaskID, st.Cron, st.Title, st.Content, st.Url, st.CreatedBy)
}

func (st *CronMsgService) Edit(data map[string]interface{}) error {
	tenantID := ""
	if !st.IsAdmin {
		tenantID = st.CreatedBy
		if st.ModifiedBy != "" {
			tenantID = st.ModifiedBy
		}
	}
	return models.EditCronMsg(st.ID, data, tenantID)
}

func (st *CronMsgService) GetByID() (models.CronMessages, error) {
	return models.GetCronMsgByID(st.ID)
}

func (st *CronMsgService) Count() (int64, error) {
	return models.GetCronMessagesTotal(st.Name, models.TenantMap(st.IsAdmin, st.CreatedBy))
}

func (st *CronMsgService) GetAll() ([]CronMsgResult, error) {
	msgs, err := models.GetCronMessages(st.PageNum, st.PageSize, st.Name, models.TenantMap(st.IsAdmin, st.CreatedBy))
	if err != nil {
		return nil, err
	}
	return st.FillNextExecTime(msgs), nil
}

func (st *CronMsgService) FillNextExecTime(msgs []models.CronMessages) []CronMsgResult {
	result := make([]CronMsgResult, 0)
	for _, msg := range msgs {
		r := CronMsgResult{
			CronMessages: msg,
			NextTime:     GetCronNextTime(msg.Cron),
		}
		result = append(result, r)
	}
	return result
}

func (st *CronMsgService) Delete() error {
	tenantID := ""
	if !st.IsAdmin {
		tenantID = st.CreatedBy
		if st.ModifiedBy != "" {
			tenantID = st.ModifiedBy
		}
	}
	return models.DeleteCronMsg(st.ID, tenantID)
}

// GetCronNextTime 获取下次的执行时间
func GetCronNextTime(cronExpr string) string {
	specParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor)
	schedule, err := specParser.Parse(cronExpr)
	if err != nil {
		return ""
	}
	nextTime := schedule.Next(time.Now()).Format("2006-01-02 15:04:05")
	return nextTime
}

// SendNow 立即发送定时消息（根据定时消息ID）
func (st *CronMsgService) SendNow(callerIP string) error {
	// 获取定时消息详情
	msg, err := st.GetByID()
	if err != nil {
		return err
	}

	// 调用发送服务
	return SendCronMessage(msg, callerIP)
}

// SendNowByParams 立即发送定时消息（根据传入的参数）
func (st *CronMsgService) SendNowByParams(callerIP string) error {
	// 直接使用传入的参数构造消息对象
	msg := models.CronMessages{
		TaskID:  st.TaskID,
		Title:   st.Title,
		Content: st.Content,
		Url:     st.Url,
	}

	// 调用发送服务
	return SendCronMessage(msg, callerIP)
}
