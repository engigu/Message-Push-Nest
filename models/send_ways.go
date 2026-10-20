package models

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"message-nest/pkg/util"
)

type SendWays struct {
	UUIDModel

	Name string `json:"name" gorm:"type:varchar(100);default:'';index:idx_ways_name"`
	Type string `json:"type" gorm:"type:varchar(100);default:'';index"`
	Auth string `json:"auth" gorm:"type:varchar(2048);default:'';"`
}

func GenerateWayUniqueID() string {
	newUUID := util.GenerateUniqueID()
	return fmt.Sprintf("WY%s", newUUID)
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

func GetSendWaysTotal(name string, type_ string, maps interface{}) (int64, error) {
	var (
		err   error
		total int64
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
	err := db.Where("id = ? ", id).Limit(1).Find(&way).Error
	if err != nil {
		return way, err
	}
	if way.ID == "" {
		return way, gorm.ErrRecordNotFound
	}
	return way, nil
}

func GetWayByName(name string, createdBy ...string) (SendWays, error) {
	var way SendWays
	query := db.Where("name = ?", name)
	if len(createdBy) > 0 && createdBy[0] != "" {
		query = query.Where("created_by = ?", createdBy[0])
	}
	err := query.Limit(1).Find(&way).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return way, err
	}
	return way, nil
}

func DeleteMsgWay(id string, createdBy ...string) error {
	query := db.Where("id = ?", id)
	if len(createdBy) > 0 && createdBy[0] != "" {
		query = query.Where("created_by = ?", createdBy[0])
	}
	result := query.Delete(&SendWays{})
	if result.Error != nil {
		return result.Error
	}
	if len(createdBy) > 0 && createdBy[0] != "" && result.RowsAffected == 0 {
		return ErrNoPermissionOrNotFound
	}
	return nil
}

func EditSendWay(id string, data interface{}, createdBy ...string) error {
	query := db.Model(&SendWays{}).Where("id = ?", id)
	if len(createdBy) > 0 && createdBy[0] != "" {
		query = query.Where("created_by = ?", createdBy[0])
	}
	result := query.Updates(data)
	if result.Error != nil {
		return result.Error
	}
	if len(createdBy) > 0 && createdBy[0] != "" && result.RowsAffected == 0 {
		return ErrNoPermissionOrNotFound
	}

	return nil
}
