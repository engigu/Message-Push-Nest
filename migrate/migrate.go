package migrate

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/logging"
	"message-nest/pkg/util"
	"message-nest/service/settings_service"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// PreMigrateAuthTable 预先平滑处理 Auth 表的主键结构升级（整型自增 -> varchar(36) 随机字符串）
func PreMigrateAuthTable(db *gorm.DB) {
	authTableName := models.GetSchema(&models.Auth{})
	if !db.Migrator().HasTable(&models.Auth{}) {
		return
	}

	// 检查当前 id 字段类型
	if db.Dialector.Name() == "mysql" || db.Dialector.Name() == "tidb" {
		var colInfo struct {
			Field string `gorm:"column:Field"`
			Type  string `gorm:"column:Type"`
			Extra string `gorm:"column:Extra"`
		}
		err := db.Raw(fmt.Sprintf("SHOW COLUMNS FROM %s WHERE Field = 'id'", authTableName)).Scan(&colInfo).Error
		if err == nil && (strings.Contains(strings.ToLower(colInfo.Type), "int") || strings.Contains(strings.ToLower(colInfo.Extra), "auto_increment")) {
			logging.Database.Infof("检测到 %s 表的主键 id 为整型/自增，正在执行主键类型平滑升级...", authTableName)

			// 1. 读取现存用户信息 (用户名列表)
			var usernames []string
			_ = db.Table(authTableName).Pluck("username", &usernames).Error

			// 2. 解除 AUTO_INCREMENT 并更改为 varchar(36)
			_ = db.Exec(fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN id int NOT NULL", authTableName)).Error
			err = db.Exec(fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN id varchar(36) NOT NULL", authTableName)).Error
			if err != nil {
				logging.Database.Errorf("修改 %s 表主键为 varchar(36) 失败: %v", authTableName, err)
			} else {
				logging.Database.Infof("成功将 %s 表主键升级为 varchar(36)", authTableName)
			}

			// 3. 为已存在的用户分配全新的随机字符串 ID
			for _, uname := range usernames {
				newUUID := models.GenerateUserUniqueID()
				db.Table(authTableName).Where("username = ?", uname).Update("id", newUUID)
				logging.Database.Infof("已为已有用户 [%s] 生成随机字符串 ID: %s", uname, newUUID)
			}
		}
	} else if db.Dialector.Name() == "postgres" {
		var udtName string
		db.Raw("SELECT udt_name FROM information_schema.columns WHERE table_name = ? AND column_name = 'id'", authTableName).Scan(&udtName)
		if strings.Contains(strings.ToLower(udtName), "int") {
			logging.Database.Infof("检测到 Postgres %s 表的主键 id 为整型，正在升级为 varchar(36)...", authTableName)
			_ = db.Exec(fmt.Sprintf("ALTER TABLE %s ALTER COLUMN id TYPE varchar(36) USING id::varchar(36)", authTableName)).Error
		}
	}
}

// 初始化admin账户
func InitAuthTableData() {
	initSection := "init"
	initAuthKey := "account"
	initAccount := "admin"

	adminUser, err := models.GetUserByUsername(initAccount)
	if err == nil && adminUser != nil && adminUser.ID != "" {
		// 已经存在 admin 用户，保证其角色为 admin
		if adminUser.Role != "admin" {
			_ = models.EditUser(initAccount, map[string]interface{}{"role": "admin"})
		}
		return
	}

	settingO, err := models.GetSettingByKey(initSection, initAuthKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logging.Init.Errorf("查询账号初始化失败！")
		return
	}
	if settingO.Value == "1" {
		return
	}

	initAccountPasswd := util.GenerateRandomString(10)

	createdUser, err := models.AddUser(initAccount, initAccountPasswd, "admin", "system")
	if err != nil {
		logging.Init.Errorf("添加初始化admin账号失败！")
		return
	} else {
		logging.Init.Infof("初始化admin账号成功！您的账号：%s 密码：%s ID: %s", initAccount, initAccountPasswd, createdUser.ID)
	}

	err = models.AddOneSetting(models.Settings{Section: initSection, Key: initAuthKey, Value: "1"})
	if err != nil {
		logging.Init.Errorf("标记admin账号初始化状态失败！err: %s", err.Error())
		return
	}
}

// BackfillInitialDataToAdmin 将系统现存的初始/历史数据全部统一标记归属为管理员
func BackfillInitialDataToAdmin() {
	backfillSetting, _ := models.GetSettingByKey("init", "all_data_to_admin_backfill_v2")
	if backfillSetting.Value == "1" {
		return
	}

	adminUser, err := models.GetUserByUsername("admin")
	if err != nil || adminUser == nil || adminUser.ID == "" {
		logging.Init.Warnf("未找到管理员账号，暂跳过历史数据归属标记")
		return
	}

	adminID := adminUser.ID
	logging.Init.Infof("正在将系统内所有初始存量数据批量归属标记为管理员 (ID: %s)...", adminID)

	db := models.GetDB()

	// 1. 发信渠道
	waysTable := models.GetSchema(models.SendWays{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", waysTable), adminID, adminID)

	// 2. 发信任务
	tasksTable := models.GetSchema(models.SendTasks{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", tasksTable), adminID, adminID)

	// 3. 任务实例
	insTable := models.GetSchema(models.SendTasksIns{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", insTable), adminID, adminID)

	// 4. 消息模板
	templateTable := models.GetSchema(models.Template{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", templateTable), adminID, adminID)

	// 5. 定时消息
	cronTable := models.GetSchema(models.CronMessages{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", cronTable), adminID, adminID)

	// 6. 发信日志
	logsTable := models.GetSchema(models.SendTasksLogs{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", logsTable), adminID, adminID)

	// 7. 托管消息
	hostedTable := models.GetSchema(models.HostedMessage{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", hostedTable), adminID, adminID)

	// 8. 每日统计
	statsTable := models.GetSchema(models.SendStats{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", statsTable), adminID)

	// 9. 用户自身
	authTable := models.GetSchema(models.Auth{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE (created_by = '' OR created_by IS NULL) AND username = 'admin'", authTable), "system", "system")
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE (created_by = '' OR created_by IS NULL) AND username != 'admin'", authTable), adminID, adminID)

	// 10. 系统配置表 (message_settings)
	settingsTable := models.GetSchema(models.Settings{})
	db.Exec(fmt.Sprintf("UPDATE %s SET created_by = ?, modified_by = ? WHERE created_by = '' OR created_by = 'admin' OR created_by IS NULL", settingsTable), adminID, adminID)

	_ = models.AddOneSetting(models.Settings{Section: "init", Key: "all_data_to_admin_backfill_v2", Value: "1", CreatedBy: adminID, ModifiedBy: adminID})
	logging.Init.Infof("初始存量数据已全部归属于管理员 (ID: %s)", adminID)
}

// calculateModelsSignature 计算所有数据模型结构特征的 MD5 签名
func calculateModelsSignature(tables []interface{}) string {
	var sigData string
	for _, table := range tables {
		t := reflect.TypeOf(table)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		sigData += t.Name() + ":"
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath == "" { // 仅处理公开可见的字段
				sigData += fmt.Sprintf("%s/%s/%s;", field.Name, field.Type.String(), field.Tag.Get("gorm"))
			}
		}
		sigData += "|"
	}
	hasher := md5.New()
	hasher.Write([]byte(sigData))
	return hex.EncodeToString(hasher.Sum(nil))
}

// checkAndMigrateTables 检测数据模型特征并选择性执行数据库迁移
func checkAndMigrateTables(db *gorm.DB, tables []interface{}) {
	// 1. 优先迁移 Settings 表（如果是新数据库），保证可以安全读取/存储迁移签名
	err := db.AutoMigrate(&models.Settings{})
	if err != nil {
		logging.Database.Errorf("迁移配置表失败: %s", err.Error())
	}

	// 2. 计算当前模型特征签名
	currentSig := calculateModelsSignature(tables)

	// 3. 从数据库读取旧特征签名，比对是否需要迁移
	storedSigSetting, err := models.GetSettingByKey("migration", "sig")
	needMigrate := false
	if err != nil {
		needMigrate = true // 获取设置报错（可能不存在该记录），需要迁移
	} else if storedSigSetting.ID == 0 || storedSigSetting.Value != currentSig {
		needMigrate = true // 记录为空，或者签名不一致，需要迁移
	}

	if needMigrate {
		logging.Database.Infof("检测到模型特征发生变更，开始更新数据库结构... 当前特征签名: %s", currentSig)
		for _, table := range tables {
			tableName := models.GetSchema(table)
			logging.Database.Infof("正在迁移数据表: %s", tableName)
			err := db.AutoMigrate(table)
			if err != nil {
				logging.Database.Errorf("迁移数据表 %s 发生错误: %s", tableName, err.Error())
			}
		}

		// 保存/更新迁移签名
		if storedSigSetting.ID > 0 {
			err = models.EditSetting(storedSigSetting.ID, map[string]interface{}{"value": currentSig})
		} else {
			err = models.AddOneSetting(models.Settings{
				Section: "migration",
				Key:     "sig",
				Value:   currentSig,
			})
		}
		if err != nil {
			logging.Database.Errorf("保存迁移签名错误: %s", err.Error())
		}
	} else {
		logging.Database.Infof("模型特征与数据库一致，跳过数据库自动迁移流程。Sig: %s", currentSig)
	}
}

func Setup() {
	db := models.Setup()

	// 预迁移处理主键升级
	PreMigrateAuthTable(db)

	tables := []interface{}{
		&models.Auth{},
		&models.SendTasks{},
		&models.SendWays{},
		&models.SendTasksLogs{},
		&models.SendTasksIns{},
		&models.Settings{},
		&models.CronMessages{},
		&models.HostedMessage{},
		&models.LoginLog{},
		&models.Template{},
		&models.SendStats{},
		&models.UserSSOIdentity{},
	}

	checkAndMigrateTables(db, tables)

	logging.Init.Infof("正在初始化管理员账号数据...")
	InitAuthTableData()

	logging.Init.Infof("正在检查初始存量数据归属标记...")
	BackfillInitialDataToAdmin()

	adminUser, _ := models.GetUserByUsername("admin")
	adminID := ""
	if adminUser != nil {
		adminID = adminUser.ID
	}

	logging.Init.Infof("正在初始化自定义站点配置数据...")
	ss := settings_service.InitSettingService{AdminID: adminID}
	ss.InitSiteConfig()

	logging.Init.Infof("正在初始化定时任务配置数据...")
	ss.InitLogConfig()
	ss.InitHostedMsgConfig()

	logging.Init.Infof("所有数据表基础数据初始化完成。")

	// 补全历史数据的 UniqueKey（仅在未标记完成时执行）
	backfillSetting, _ := models.GetSettingByKey("init", "hosted_msg_uniquekey_backfill")
	if backfillSetting.Value != "1" {
		go BackfillHostedMessagesUniqueKey()
	}
}

// BackfillHostedMessagesUniqueKey 为历史托管消息生成 UniqueKey，并持久化完成标记到数据库
func BackfillHostedMessagesUniqueKey() {
	// 双重校验，如果已经标记完成直接退出
	backfillSetting, _ := models.GetSettingByKey("init", "hosted_msg_uniquekey_backfill")
	if backfillSetting.Value == "1" {
		return
	}

	var messages []models.HostedMessage
	err := models.GetDB().Model(&models.HostedMessage{}).Select("id").Where("unique_key = ? OR unique_key IS NULL", "").Find(&messages).Error
	if err != nil {
		logging.Init.Errorf("查找未设置 unique_key 的托管消息失败: %s", err.Error())
		return
	}
	total := len(messages)
	if total == 0 {
		models.AddOneSetting(models.Settings{Section: "init", Key: "hosted_msg_uniquekey_backfill", Value: "1"})
		return
	}

	logging.Init.Infof("发现有 %d 条托管消息缺少 unique_key，正在为历史托管消息补充唯一标识 (UniqueKey)...", total)

	batchSize := 500
	tableName := models.GetSchema(models.HostedMessage{})

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		batch := messages[i:end]

		sqlRaw := "UPDATE " + tableName + " SET unique_key = CASE id "
		var args []interface{}
		var ids []int

		for _, msg := range batch {
			uniqueKey := util.GenerateRandomString(16)
			sqlRaw += "WHEN ? THEN ? "
			args = append(args, msg.ID, uniqueKey)
			ids = append(ids, msg.ID)
		}
		sqlRaw += "END WHERE id IN (?)"
		
		args = append(args, ids)

		err = models.GetDB().Exec(sqlRaw, args...).Error
		if err != nil {
			logging.Init.Errorf("批量补全第 %d 到 %d 条记录的 unique_key 失败: %s", i+1, end, err.Error())
			return
		}
	}

	models.AddOneSetting(models.Settings{Section: "init", Key: "hosted_msg_uniquekey_backfill", Value: "1"})
	logging.Init.Infof("成功补全 %d 条历史托管消息的 unique_key 并已记录完成标记", total)
}
