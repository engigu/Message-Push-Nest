package migrate

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"message-nest/models"
	"message-nest/pkg/setting"
)

func InitAuthTableData() {
	initSection := "init"
	initAuthKey := "account"
	initAccount := "admin"
	initAccountPasswd := "123456"

	settingO, err := models.GetSettingByKey(initSection, initAuthKey)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logrus.Error(fmt.Sprintf("查询账号初始化失败！"))
		return
	}
	if settingO.Value == "1" {
		// 已经初始化过
		return
	}
	err = models.AddUser(initAccount, initAccountPasswd)
	if err != nil {
		logrus.Error(fmt.Sprintf("添加初始化admin账号失败！"))
		return
	} else {
		logrus.Error(fmt.Sprintf("初始化admin账号成功！您的账号：%s 密码：%s", initAccount, initAccountPasswd))
	}

	err = models.AddOneSetting(models.Settings{Section: initSection, Key: initAuthKey, Value: "1"})
	if err != nil {
		logrus.Error(fmt.Sprintf("标记admin账号初始化状态失败！"))
		return
	}
}

func Setup() {
	setting.Setup()
	db := models.Setup()
	defer func(db *gorm.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	db.AutoMigrate(
		&models.Auth{},
		&models.SendTasks{},
		&models.SendWays{},
		&models.SendTasksLogs{},
		&models.SendTasksIns{},
		&models.Settings{},
	)

	InitAuthTableData()

}
