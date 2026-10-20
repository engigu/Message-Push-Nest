package models

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// InitTestDB 初始化内存 SQLite 数据库用于自动化单元测试
func InitTestDB(name ...string) *gorm.DB {
	dbName := "test"
	if len(name) > 0 && name[0] != "" {
		dbName = name[0]
	}
	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "message_",
			SingularTable: true,
		},
	}
	testDb, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), config)
	if err != nil {
		panic(err)
	}
	db = testDb
	_ = db.AutoMigrate(
		&Auth{},
		&SendWays{},
		&SendTasks{},
		&SendTasksIns{},
		&Template{},
		&CronMessages{},
		&SendTasksLogs{},
		&HostedMessage{},
		&SendStats{},
		&LoginLog{},
		&Settings{},
		&UserSSOIdentity{},
	)
	return db
}
