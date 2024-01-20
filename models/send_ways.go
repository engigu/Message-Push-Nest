package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"message-nest/pkg/util"
)

type SendWays struct {
	UUIDModel

	Name string `json:"name" gorm:"type:varchar(100) comment '渠道名称';default:'';"`
	Type string `json:"type" gorm:"type:varchar(100) comment '渠道类型';default:'';index:type"`
	Auth string `json:"auth" gorm:"type:varchar(2048) comment '认证信息';default:'';"`
}

func GenerateWayUniqueID() string {
	newUUID := util.GenerateUniqueID()
	return fmt.Sprintf("W-%s", newUUID)
}

func AddSendWay(name string, auth string, wayType string, createdBy string, modifiedBy string) error {
	newUUID := GenerateWayUniqueID()
	way := SendWays{
		UUIDModel: UUIDModel{
			ID:         newUUID,
			CreatedBy:  createdBy,
			ModifiedBy: modifiedBy,
		},
		Name: name,
		Type: wayType,
		Auth: auth,
	}
	if err := db.Create(&way).Error; err != nil {
		return err
	}
	return nil
}

func GetSendWays(pageNum int, pageSize int, name string, type_ string, maps interface{}) ([]SendWays, error) {
	var (
		ways []SendWays
		err  error
	)
	query := db.Where(maps)
	if name != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", name))
	}
	if type_ != "" {
		query = query.Where("type = ?", type_)
	}

	query = query.Order("created_on DESC")
	if pageSize > 0 || pageNum > 0 {
		query = query.Offset(pageNum).Limit(pageSize)
	}

	err = query.Find(&ways).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return ways, nil
}

func GetSendWaysTotal(name string, type_ string, maps interface{}) (int, error) {
	var (
		err   error
		total int
	)
	query := db.Model(&SendWays{}).Where(maps)
	if name != "" {
		query = query.Where("name like ?", fmt.Sprintf("%%%s%%", name))
	}
	if type_ != "" {
		query = query.Where("type = ?", type_)
	}

	err = query.Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}

func GetWayByID(id string) (SendWays, error) {
	var way SendWays
	err := db.Where("id = ? ", id).Find(&way).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return way, err
	}
	return way, nil
}

func DeleteMsgWay(id string) error {
	if err := db.Where("id = ?", id).Delete(&SendWays{}).Error; err != nil {
		return err
	}

	return nil
}

func EditSendWay(id string, data interface{}) error {
	if err := db.Model(&SendWays{}).Where("id = ? ", id).Updates(data).Error; err != nil {
		return err
	}

	return nil
}
