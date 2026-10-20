package send_way_service

import (
	"fmt"
	"message-nest/models"
	"strings"
)

type SendWay struct {
	ID         string
	Name       string
	Type       string
	CreatedBy  string
	ModifiedBy string
	Auth       string
	CreatedOn  string

	IsAdmin  bool
	PageNum  int
	PageSize int
}

func (sw *SendWay) GetByID() (interface{}, error) {
	return models.GetWayByID(sw.ID)
}

func (sw *SendWay) NameIsExist(method string) error {
	tenantID := ""
	if !sw.IsAdmin {
		tenantID = sw.CreatedBy
		if method == "edit" && sw.ModifiedBy != "" {
			tenantID = sw.ModifiedBy
		}
	}
	way, err := models.GetWayByName(sw.Name, tenantID)
	if err != nil {
		return err
	}
	if method == "add" {
		if len(way.ID) > 0 {
			return fmt.Errorf("已经存在同名的渠道名：%s", way.Name)
		}
	} else {
		// 只允许当前的重名
		if len(way.ID) > 0 && way.ID != sw.ID {
			return fmt.Errorf("已经存在同名的渠道名：%s", way.Name)
		}
	}
	return nil
}

func (sw *SendWay) Add() error {
	err := sw.NameIsExist("add")
	if err != nil {
		return err
	}
	return models.AddSendWay(sw.Name, sw.Auth, sw.Type, sw.CreatedBy, sw.ModifiedBy)
}

func (sw *SendWay) Edit() error {
	tenantID := ""
	if !sw.IsAdmin && sw.ModifiedBy != "" {
		tenantID = sw.ModifiedBy
	}
	err := sw.NameIsExist("edit")
	if err != nil {
		return err
	}
	data := make(map[string]interface{})
	data["modified_by"] = sw.ModifiedBy
	data["name"] = sw.Name
	data["auth"] = sw.Auth
	return models.EditSendWay(sw.ID, data, tenantID)
}

func (sw *SendWay) Delete() error {
	tenantID := ""
	if !sw.IsAdmin && sw.ModifiedBy != "" {
		tenantID = sw.ModifiedBy
	}
	tasks := models.FindTaskByWayId(sw.ID)
	if len(tasks) > 0 {
		var names []string
		for _, task := range tasks {
			names = append(names, task.Name)
		}
		return fmt.Errorf("已经存在使用的任务，删除失败！任务名：%s", strings.Join(names, ", "))
	}
	return models.DeleteMsgWay(sw.ID, tenantID)
}

func (sw *SendWay) Count() (int64, error) {
	return models.GetSendWaysTotal(sw.Name, sw.Type, models.TenantMap(sw.IsAdmin, sw.CreatedBy))
}

func (sw *SendWay) GetAll() ([]models.SendWays, error) {
	tags, err := models.GetSendWays(sw.PageNum, sw.PageSize, sw.Name, sw.Type, models.TenantMap(sw.IsAdmin, sw.CreatedBy))
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// getValidator 根据渠道类型获取对应的验证器
func (sw *SendWay) getValidator() WayValidator {
	factory, exists := validatorRegistry[sw.Type]
	if !exists {
		return nil
	}
	return factory()
}

// ValidateDiffWay 各种发信渠道具体字段校验
func (sw *SendWay) ValidateDiffWay() (string, interface{}) {
	var empty interface{}
	validator := sw.getValidator()
	if validator == nil {
		return fmt.Sprintf("未知的发信渠道校验: %s", sw.Type), empty
	}
	return validator.Validate(sw.Auth)
}

// TestSendWay 尝试带发信测试连通性
func (sw *SendWay) TestSendWay(msgObj interface{}) (string, string) {
	factory, exists := testerRegistry[sw.Type]
	if !exists {
		return fmt.Sprintf("未知的发信渠道测试: %s", sw.Type), ""
	}
	tester := factory(msgObj)
	return tester.Test()
}
