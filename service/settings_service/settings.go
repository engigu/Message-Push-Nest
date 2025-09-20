package settings_service

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/constant"
	"message-nest/pkg/util"
	"message-nest/service/cron_service"
)

type UserSettings struct {
	UserName    string
	OldPassword string
	NewPassword string
}

// EditUserPasswd 用户设置密码
func (us *UserSettings) EditUserPasswd() error {
	ok, _ := models.CheckAuth(us.UserName, us.OldPassword)
	if !ok {
		return errors.New("旧密码校验失败！")
	}
	var user = map[string]interface{}{
		"password": us.NewPassword,
	}
	return models.EditUser(us.UserName, user)
}

// GetUserSetting 获取用户设置
func (us *UserSettings) GetUserSetting(section string) (map[string]string, error) {
	// 如果是site_config，优先从缓存获取
	if section == constant.SiteSettingSectionName {
		if IsSiteConfigCacheValid() {
			return GetSiteConfigCache(), nil
		}
	}
	
	result := make(map[string]string)
	settings, err := models.GetSettingBySection(section)
	if err != nil {
		return result, err
	}
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}
	
	// 如果是site_config，更新缓存
	if section == constant.SiteSettingSectionName {
		SetSiteConfigCache(result)
	}
	
	// 版本信息单独获取
	if section == constant.AboutSectionName {
		result = constant.LatestVersion
		// 添加内存使用信息
		memoryInfo := util.GetMemoryUsage()
		for key, value := range memoryInfo {
			result[key] = value
		}
	}
	return result, nil
}

// EditSettings 便捷自定义设置
func (us *UserSettings) EditSettings(section string, key string, value string, currentUser string) error {
	setting, _ := models.GetSettingByKey(section, key)
	if setting.ID <= 0 {
		err := models.AddOneSetting(models.Settings{
			IDModel: models.IDModel{
				CreatedBy:  currentUser,
				ModifiedBy: currentUser,
			},
			Section: section,
			Key:     key,
			Value:   value,
		})
		// 如果是site_config，清除缓存
		if section == constant.SiteSettingSectionName {
			ClearSiteConfigCache()
		}
		return err
	} else {
		if value == "" {
			return nil
		}
		data := make(map[string]interface{})
		data["section"] = section
		data["key"] = key
		data["value"] = value
		data["modified_by"] = currentUser
		err := models.EditSetting(setting.ID, data)
		if key == constant.LogsCleanCronKeyName {
			cronService := cron_service.CronService{}
			cronService.UpdateLogsCronRun(value)
		}
		// 如果是site_config，清除缓存
		if section == constant.SiteSettingSectionName {
			ClearSiteConfigCache()
		}
		return err
	}
}

// 站点自定义的结构
type SiteConfig struct {
	Title  string `json:"title" validate:"omitempty,min=1,max=50" label:"网站标题"`
	Slogan string `json:"slogan" validate:"omitempty,min=1,max=50" label:"网站slogan"`
	Logo   string `json:"logo" validate:"omitempty,min=1" label:"logo"`
}

type LogConfig struct {
	Cron    string `json:"cron" validate:"required,cron" label:"日志定时表达式"`
	KeepNum string `json:"keep_num" validate:"required,min=1,max=50" label:"日志保留数"`
}

// ValidateDiffSetting 校验不同的设置
func (us *UserSettings) ValidateDiffSetting(section string, data map[string]string) string {
	if section == constant.SiteSettingSectionName {
		var config SiteConfig
		config.Title = data["title"]
		config.Slogan = data["slogan"]
		config.Logo = data["logo"]
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	if section == constant.LogsCleanSectionName {
		var config LogConfig
		config.Cron = data["cron"]
		config.KeepNum = data["keep_num"]
		_, err := cron.ParseStandard(config.Cron)
		if err != nil {
			return fmt.Sprintf("%s 不是合法的corn表达式", config.Cron)
		}
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	return fmt.Sprintf("未知的section：%s", section)
}
