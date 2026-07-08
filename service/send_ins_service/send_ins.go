package send_ins_service

import (
	"fmt"
	"message-nest/models"
)

type SendTaskInsService struct {
	ID         string
	Name       string
	CreatedBy  string
	ModifiedBy string
	CreatedOn  string

	PageNum  int
	PageSize int
	Enable   int
}

// ValidateDiffIns 各种发信渠道具体字段校验
func (sw *SendTaskInsService) ValidateDiffIns(ins models.SendTasksIns) (string, interface{}) {
	var empty interface{}
	factory, exists := insValidatorRegistry[ins.WayType]
	if !exists {
		return "未知的渠道的config校验", empty
	}
	validator := factory()
	return validator.Validate(ins.Config)
}

func (st *SendTaskInsService) ManyAdd(taskIns []models.SendTasksIns) string {

	for _, ins := range taskIns {
		errStr, _ := st.ValidateDiffIns(ins)
		if errStr != "" {
			return errStr
		}
	}
	err := models.ManyAddTaskIns(taskIns)
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return ""
}

func (st *SendTaskInsService) AddOne(ins models.SendTasksIns) string {
	errStr, _ := st.ValidateDiffIns(ins)
	if errStr != "" {
		return errStr
	}
	err := models.AddTaskInsOne(ins)
	if err != nil {
		return fmt.Sprintf("%s", err)
	}
	return ""
}

func (st *SendTaskInsService) Delete() error {
	return models.DeleteMsgTaskIns(st.ID)
}

func (st *SendTaskInsService) Update(data map[string]interface{}) error {
	return models.UpdateMsgTaskIns(st.ID, data)
}

func (st *SendTaskInsService) Count() (int64, error) {
	return models.GetSendTasksTotal(st.Name, st.getMaps())
}

func (st *SendTaskInsService) GetAll() ([]models.SendTasks, error) {
	tasks, err := models.GetSendTasks(st.PageNum, st.PageSize, st.Name, st.getMaps())
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (st *SendTaskInsService) getMaps() map[string]interface{} {
	maps := make(map[string]interface{})
	return maps
}
