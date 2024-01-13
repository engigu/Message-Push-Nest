package env_service

import (
	"github.com/sirupsen/logrus"
	"message-nest/models"
	"message-nest/pkg/constant"
)

type EnvService struct {
}

func (es *EnvService) CommonAdd(section string, key string, value string) {
	setting, _ := models.GetSettingByKey(section, key)
	if setting.ID <= 0 {
		err := models.AddOneSetting(models.Settings{
			Section: section,
			Key:     key,
			Value:   value,
		})
		if err != nil {
			logrus.Error("初始化%s:%s失败", section, key)
		} else {
			logrus.Infof("初始化%s:%s成功", section, key)
		}
	}
}

// InitSiteConfig 初始化、重置站点信息设置
func (es *EnvService) InitSiteConfig() {
	section := constant.SiteSettingSectionName
	for key, value := range constant.SiteSiteDefaultValueMap {
		es.CommonAdd(section, key, value)
	}
}

// InitLogConfig 初始化日志清理设置
func (es *EnvService) InitLogConfig() {
	section := constant.LogsCleanSectionName
	for key, value := range constant.LogsCleanDefaultValueMap {
		es.CommonAdd(section, key, value)
	}
}

func Setup() {
	es := EnvService{}
	es.InitSiteConfig()
	es.InitLogConfig()
}
