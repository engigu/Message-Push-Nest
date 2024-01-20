package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"time"
)

var db *gorm.DB

type IDModel struct {
	ID int `gorm:"type:int(11) AUTO_INCREMENT comment 'id';primary_key" json:"id"`

	CreatedBy  string    `json:"created_by" gorm:"type:varchar(100) comment '创建人';default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(100) comment '修改人';default:'';"`
	CreatedOn  util.Time `json:"created_on" gorm:"type:timestamp comment '创建时间';default:current_timestamp;"`
	ModifiedOn util.Time `json:"modified_on" gorm:"type:timestamp comment '更新时间';"`
}

type UUIDModel struct {
	ID string `gorm:"type:varchar(12) comment 'id';primary_key" json:"id"`

	CreatedBy  string    `json:"created_by" gorm:"type:varchar(100) comment '创建人';default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(100) comment '修改人';default:'';"`
	CreatedOn  util.Time `json:"created_on" gorm:"type:timestamp comment '创建时间';default:current_timestamp;"`
	ModifiedOn util.Time `json:"modified_on" gorm:"type:timestamp comment '更新时间';"`
}

// Setup initializes the database instance
func Setup() *gorm.DB {
	var err error
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		setting.DatabaseSetting.User,
		setting.DatabaseSetting.Password,
		setting.DatabaseSetting.Host,
		setting.DatabaseSetting.Port,
		setting.DatabaseSetting.Name)
	db, err = gorm.Open(setting.DatabaseSetting.Type, connStr)

	if err != nil {
		log.Fatalf("models.Setup err: %v", err)
	}
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return setting.DatabaseSetting.TablePrefix + defaultTableName
	}

	if setting.DatabaseSetting.SqlDebug == "enable" {
		db.LogMode(true)
	}
	db.SingularTable(true)
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	return db
}

// CloseDB closes database connection (unnecessary)
func CloseDB() {
	defer db.Close()
}

// updateTimeStampForCreateCallback will set `CreatedOn`, `ModifiedOn` when creating
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now()

		if createTimeField, ok := scope.FieldByName("CreatedOn"); ok {
			if createTimeField.IsBlank {
				createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("ModifiedOn"); ok {
			if modifyTimeField.IsBlank {
				modifyTimeField.Set(nowTime)
			}
		}
	}
}

// updateTimeStampForUpdateCallback will set `ModifiedOn` when updating
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		scope.SetColumn("ModifiedOn", time.Now())
	}
}

// deleteCallback will set `DeletedOn` where deleting
func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedOnField, hasDeletedOnField := scope.FieldByName("DeletedOn")

		if !scope.Search.Unscoped && hasDeletedOnField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v%v%v",
				scope.QuotedTableName(),
				scope.Quote(deletedOnField.DBName),
				scope.AddToVars(time.Now().Unix()),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

// addExtraSpaceIfExist adds a separator
func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}
