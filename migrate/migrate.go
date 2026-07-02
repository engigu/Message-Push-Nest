package migrate

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/util"
	"message-nest/service/settings_service"
	"reflect"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// 初始化admin账户
func InitAuthTableData() {
	initSection := "init"
	initAuthKey := "account"
	initAccount := "admin"

	settingO, err := models.GetSettingByKey(initSection, initAuthKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Error(fmt.Sprintf("查询账号初始化失败！"))
		return
	}
	if settingO.Value == "1" {
		// 已经初始化过
		return
	}
	
	initAccountPasswd := util.GenerateRandomString(10)
	
	err = models.AddUser(initAccount, initAccountPasswd)
	if err != nil {
		logrus.Error(fmt.Sprintf("添加初始化admin账号失败！"))
		return
	} else {
		logrus.Info(fmt.Sprintf("初始化admin账号成功！您的账号：%s 密码：%s", initAccount, initAccountPasswd))
	}

	err = models.AddOneSetting(models.Settings{Section: initSection, Key: initAuthKey, Value: "1"})
	if err != nil {
		logrus.Error(fmt.Sprintf("标记admin账号初始化状态失败！err: %s", err.Error()))
		return
	}
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
func checkAndMigrateTables(db *gorm.DB, tables []interface{}, entry *logrus.Entry) {
	// 1. 优先迁移 Settings 表（如果是新数据库），保证可以安全读取/存储迁移签名
	err := db.AutoMigrate(&models.Settings{})
	if err != nil {
		entry.Errorf("Migrate settings table error: %s", err.Error())
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
		entry.Infof("检测到模型特征发生变更，开始更新数据库结构... Current Sig: %s", currentSig)
		for _, table := range tables {
			tableName := models.GetSchema(table)
			entry.Infof("Migrate table: %s", tableName)
			err := db.AutoMigrate(table)
			if err != nil {
				entry.Infof("Migrate table erorr: %s", err.Error())
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
			entry.Errorf("保存迁移签名错误: %s", err.Error())
		}
	} else {
		entry.Infof("模型特征与数据库一致，跳过数据库自动迁移流程。Sig: %s", currentSig)
	}
}

func Setup() {
	db := models.Setup()
	//defer func(db *gorm.DB) {
	//	err := db.Close()
	//	if err != nil {
	//
	//	}
	//}(db)

	//if setting.AppSetting.InitData != "enable" {
	//	return
	//}

	entry := logrus.WithFields(logrus.Fields{
		"prefix": "[Init Data]",
	})

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
	}

	checkAndMigrateTables(db, tables, entry)

	entry.Infof("Init Account data...")
	InitAuthTableData()

	entry.Infof("Init Custom Site data...")
	ss := settings_service.InitSettingService{}
	ss.InitSiteConfig()

	entry.Infof("Init Cron data...")
	ss.InitLogConfig()
	ss.InitHostedMsgConfig()

	entry.Infof("All table data init done.")

	// 补全历史数据的 UniqueKey
	entry.Infof("Backfilling UniqueKey for historical hosted messages...")
	go BackfillHostedMessagesUniqueKey()
}

// BackfillHostedMessagesUniqueKey 为历史托管消息生成 UniqueKey
func BackfillHostedMessagesUniqueKey() {
	var messages []models.HostedMessage
	// 仅选择需要的 id 字段减少内存与I/O开销
	err := models.GetDB().Model(&models.HostedMessage{}).Select("id").Where("unique_key = ? OR unique_key IS NULL", "").Find(&messages).Error
	if err != nil {
		logrus.Errorf("查找未设置 unique_key 的托管消息失败: %s", err.Error())
		return
	}
	total := len(messages)
	if total == 0 {
		return
	}

	logrus.Infof("发现有 %d 条托管消息缺少 unique_key，正在进行高效分批批量更新...", total)

	// 分批进行批量更新，每批处理 500 条，防止生成的 SQL 语句过长
	batchSize := 500
	tableName := models.GetSchema(models.HostedMessage{})

	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		batch := messages[i:end]

		// 拼接标准的 SQL CASE WHEN 语句实现兼容各种数据库（MySQL, TiDB, SQLite, PostgreSQL）的单条批量更新
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
		
		// GORM 执行时会自动展开 IN (?) 绑定变量切片为 IN (1, 2, 3...)，这里直接传入整数切片
		args = append(args, ids)

		err = models.GetDB().Exec(sqlRaw, args...).Error
		if err != nil {
			logrus.Errorf("批量补全第 %d 到 %d 条记录的 unique_key 失败: %s", i+1, end, err.Error())
			return
		}
	}

	logrus.Infof("成功补全 %d 条历史托管消息的 unique_key", total)
}
