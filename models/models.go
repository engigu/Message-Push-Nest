package models

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"log"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
)

var db *gorm.DB

type IDModel struct {
	ID uint `gorm:"autoIncrement;type:integer;primaryKey" json:"id"`

	CreatedBy  string    `json:"created_by" gorm:"type:varchar(100) ;default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(100) ;default:'';"`
	CreatedAt  util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime "`
	UpdatedAt  util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime ;"`
}

type UUIDModel struct {
	ID string `gorm:"type:varchar(12) ;primaryKey" json:"id"`

	CreatedBy  string    `json:"created_by" gorm:"type:varchar(100) ;default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(100) ;default:'';"`
	CreatedAt  util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime "`
	UpdatedAt  util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime ;"`
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

	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   setting.DatabaseSetting.TablePrefix,
			SingularTable: true,
		},
	}

	switch setting.DatabaseSetting.Type {
	case "mysql":
		db, err = gorm.Open(mysql.Open(connStr), config)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open("conf/database.db"), config)
	}

	if err != nil {
		log.Fatalf("models.Setup err: %v", err)
	}

	if setting.DatabaseSetting.SqlDebug == "enable" {
		db = db.Debug()
	}
	return db
}

func GetSchema(table any) string {
	stmt := &gorm.Statement{DB: db}
	stmt.Parse(table)
	return stmt.Schema.Table
}
