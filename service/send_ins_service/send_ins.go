package send_ins_service

import (
	"encoding/json"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/app"
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

// ValidateDiffWay 各种发信渠道具体字段校验
func (sw *SendTaskInsService) ValidateDiffIns(ins models.SendTasksIns) (string, interface{}) {
	var empty interface{}
	if ins.WayType == "Email" {
		var emailConfig models.InsEmailConfig
		err := json.Unmarshal([]byte(ins.Config), &emailConfig)
		if err != nil {
			return "邮箱auth反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(emailConfig)
		return Msg, emailConfig
	}
	if ins.WayType == "Dtalk" {
		var Config models.InsDtalkConfig
		return "", Config
	}
	if ins.WayType == "QyWeiXin" {
		var Config models.InsQyWeiXinConfig
		return "", Config
	}
	if ins.WayType == "MessageNest" {
		var Config models.InsQyWeiXinConfig
		return "", Config
	}
	if ins.WayType == "Custom" {
		var Config models.InsCustomConfig
		err := json.Unmarshal([]byte(ins.Config), &Config)
		if err != nil {
			return "自定义webhook反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(Config)
		return Msg, Config
	}
	if ins.WayType == "WeChatOFAccount" {
		var Config models.InsWeChatAccountConfig
		err := json.Unmarshal([]byte(ins.Config), &Config)
		if err != nil {
			return "微信公众号发送配置反序列化失败！", empty
		}
		_, Msg := app.CommonPlaygroundValid(Config)
		return Msg, Config
	}
	return "未知的渠道的config校验", empty
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
