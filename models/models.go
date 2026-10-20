package models

import (
	"fmt"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"message-nest/pkg/app"
	"message-nest/pkg/logging"
	"message-nest/pkg/setting"
	"message-nest/pkg/util"
	"os"
	"time"
)

var db *gorm.DB

// AuditModel 统一通用审计字段模型
type AuditModel struct {
	CreatedBy  string    `json:"created_by" gorm:"type:varchar(36);default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(36);default:'';"`
	CreatedAt  util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedAt  util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime"`
}

func handleBeforeCreate(tx *gorm.DB, createdBy, modifiedBy *string) {
	if tx.Statement != nil && tx.Statement.Context != nil {
		if info, ok := app.GetUserFromContext(tx.Statement.Context); ok {
			// 资产归属：如果是管理员穿越代建，归属于目标租户；未穿越归属于自身
			if *createdBy == "" {
				*createdBy = info.GetEffectiveUserID()
			}
			// 审计操作人：永远记录真实操作管理员/用户
			if *modifiedBy == "" {
				*modifiedBy = info.GetAuditUserID()
			}
		}
	}
}

func handleBeforeUpdate(tx *gorm.DB, modifiedBy *string) {
	if tx.Statement != nil && tx.Statement.Context != nil {
		if uid := app.GetAuditUserIDFromContext(tx.Statement.Context); uid != "" {
			*modifiedBy = uid
			tx.Statement.SetColumn("ModifiedBy", uid)
		}
	}
}

func (m *AuditModel) BeforeCreate(tx *gorm.DB) error {
	handleBeforeCreate(tx, &m.CreatedBy, &m.ModifiedBy)
	return nil
}

func (m *AuditModel) BeforeUpdate(tx *gorm.DB) error {
	handleBeforeUpdate(tx, &m.ModifiedBy)
	return nil
}

type IDModel struct {
	ID         uint      `gorm:"autoIncrement;type:integer;primaryKey" json:"id"`
	CreatedBy  string    `json:"created_by" gorm:"type:varchar(36);default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(36);default:'';"`
	CreatedAt  util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedAt  util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime"`
}

func (m *IDModel) BeforeCreate(tx *gorm.DB) error {
	handleBeforeCreate(tx, &m.CreatedBy, &m.ModifiedBy)
	return nil
}

func (m *IDModel) BeforeUpdate(tx *gorm.DB) error {
	handleBeforeUpdate(tx, &m.ModifiedBy)
	return nil
}

type UUIDModel struct {
	ID         string    `gorm:"type:varchar(12);primaryKey" json:"id"`
	CreatedBy  string    `json:"created_by" gorm:"type:varchar(36);default:'';"`
	ModifiedBy string    `json:"modified_by" gorm:"type:varchar(36);default:'';"`
	CreatedAt  util.Time `json:"created_on" gorm:"column:created_on;autoCreateTime"`
	UpdatedAt  util.Time `json:"modified_on" gorm:"column:modified_on;autoUpdateTime"`
}

func (m *UUIDModel) BeforeCreate(tx *gorm.DB) error {
	handleBeforeCreate(tx, &m.CreatedBy, &m.ModifiedBy)
	return nil
}

func (m *UUIDModel) BeforeUpdate(tx *gorm.DB) error {
	handleBeforeUpdate(tx, &m.ModifiedBy)
	return nil
}

// Setup initializes the database instance
func Setup() *gorm.DB {
	var err error

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             500 * time.Millisecond, // 慢 SQL 阈值设为 500ms
			LogLevel:                  logger.Warn,           // 日志级别
			IgnoreRecordNotFoundError: true,                  // 忽略未找到记录错误
			Colorful:                  true,                  // 彩色日志
		},
	)

	config := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   setting.DatabaseSetting.TablePrefix,
			SingularTable: true,
		},
		Logger:      newLogger,
		PrepareStmt: true, // 启用预编译语句缓存，降低数据库解析开销并提升吞吐
	}

	switch setting.DatabaseSetting.Type {
	case "mysql":
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			setting.DatabaseSetting.User,
			setting.DatabaseSetting.Password,
			setting.DatabaseSetting.Host,
			setting.DatabaseSetting.Port,
			setting.DatabaseSetting.Name)
		if setting.DatabaseSetting.Ssl != "false" {
			connStr += fmt.Sprintf("&tls=%s", setting.DatabaseSetting.Ssl)
		}
		db, err = gorm.Open(mysql.Open(connStr), config)
	case "tidb":
		connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true",
			setting.DatabaseSetting.User,
			setting.DatabaseSetting.Password,
			setting.DatabaseSetting.Host,
			setting.DatabaseSetting.Port,
			setting.DatabaseSetting.Name)
		if setting.DatabaseSetting.Ssl != "false" {
			connStr += fmt.Sprintf("&tls=%s", setting.DatabaseSetting.Ssl)
		}
		db, err = gorm.Open(mysql.Open(connStr), config)
	case "sqlite":
		db, err = gorm.Open(sqlite.Open("conf/database.db"), config)
	case "postgres":
		connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Shanghai",
			setting.DatabaseSetting.Host,
			setting.DatabaseSetting.User,
			setting.DatabaseSetting.Password,
			setting.DatabaseSetting.Name,
			setting.DatabaseSetting.Port,
			setting.DatabaseSetting.Ssl)
		if setting.DatabaseSetting.Ssl == "false" || setting.DatabaseSetting.Ssl == "" {
			connStr = fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
				setting.DatabaseSetting.Host,
				setting.DatabaseSetting.User,
				setting.DatabaseSetting.Password,
				setting.DatabaseSetting.Name,
				setting.DatabaseSetting.Port)
		}
		db, err = gorm.Open(postgres.Open(connStr), config)
	}

	if err != nil {
		logging.Database.Fatalf("数据库连接初始化失败: %v", err)
	}

	// 配置生产级底层连接池，防止连接风暴与偶发断连
	sqlDB, err := db.DB()
	if err != nil {
		logging.Database.Errorf("获取底层 sql.DB 失败: %v", err)
	} else {
		sqlDB.SetMaxIdleConns(setting.DatabaseSetting.MaxIdleConns)
		sqlDB.SetMaxOpenConns(setting.DatabaseSetting.MaxOpenConns)
		sqlDB.SetConnMaxLifetime(time.Duration(setting.DatabaseSetting.ConnMaxLifetime) * time.Minute)
		sqlDB.SetConnMaxIdleTime(time.Duration(setting.DatabaseSetting.ConnMaxIdleTime) * time.Minute)
		logging.Database.Infof("数据库连接池配置成功: MaxIdle=%d, MaxOpen=%d, MaxLifetime=%dm, MaxIdleTime=%dm",
			setting.DatabaseSetting.MaxIdleConns,
			setting.DatabaseSetting.MaxOpenConns,
			setting.DatabaseSetting.ConnMaxLifetime,
			setting.DatabaseSetting.ConnMaxIdleTime,
		)
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

func GetDB() *gorm.DB {
	return db
}

func SetDB(d *gorm.DB) {
	db = d
}
