package settings_service

import (
	"errors"
	"fmt"
	"message-nest/models"
	"message-nest/pkg/app"
	"message-nest/pkg/constant"
	"message-nest/pkg/util"
	"message-nest/service/cron_service"
	"strconv"

	"github.com/robfig/cron/v3"
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
		// 更新日志清理配置
		if section == constant.LogsCleanSectionName {
			if key == constant.LogsCleanCronKeyName || key == constant.LogsCleanEnabledKeyName {
				cronService := cron_service.CronService{}
				// 获取最新的cron和enabled值
				cronSetting, _ := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanCronKeyName)
				enabledSetting, _ := models.GetSettingByKey(constant.LogsCleanSectionName, constant.LogsCleanEnabledKeyName)
				cronService.UpdateLogsCronRun(cronSetting.Value, enabledSetting.Value == "true")
			}
		}
		// 更新托管消息清理配置
		if section == constant.HostedMsgCleanSectionName {
			if key == constant.HostedMsgCleanCronKeyName || key == constant.HostedMsgCleanEnabledKeyName {
				cronService := cron_service.CronService{}
				// 获取最新的cron和enabled值
				cronSetting, _ := models.GetSettingByKey(constant.HostedMsgCleanSectionName, constant.HostedMsgCleanCronKeyName)
				enabledSetting, _ := models.GetSettingByKey(constant.HostedMsgCleanSectionName, constant.HostedMsgCleanEnabledKeyName)
				cronService.UpdateHostedMsgCronRun(cronSetting.Value, enabledSetting.Value == "true")
			}
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
	Title         string `json:"title" validate:"omitempty,min=1,max=50" label:"网站标题"`
	Slogan        string `json:"slogan" validate:"omitempty,min=1,max=50" label:"网站slogan"`
	Logo          string `json:"logo" validate:"omitempty,min=1" label:"logo"`
	CookieExpDays string `json:"cookie_exp_days" validate:"omitempty,numeric,min=1,max=365" label:"cookie过期天数"`
	ThemeColor    string `json:"theme_color" validate:"omitempty,max=20" label:"主题颜色"`
}

type LogConfig struct {
	Cron    string `json:"cron" validate:"required,cron" label:"日志定时表达式"`
	KeepNum string `json:"keep_num" validate:"required,min=1,max=50" label:"日志保留数"`
	Enabled string `json:"enabled" validate:"required,oneof=true false" label:"是否启用"`
}

type HostedMsgConfig struct {
	Cron    string `json:"cron" validate:"required,cron" label:"托管消息定时表达式"`
	KeepNum string `json:"keep_num" validate:"required,min=1,max=50" label:"托管消息保留数"`
	Enabled string `json:"enabled" validate:"required,oneof=true false" label:"是否启用"`
}

// GetCookieExpDays 获取 cookie 过期天数，若无配置则返回默认值 1
func GetCookieExpDays() int {
	// 优先从缓存获取
	if IsSiteConfigCacheValid() {
		cache := GetSiteConfigCache()
		if expDays, ok := cache["cookie_exp_days"]; ok && expDays != "" {
			if days, err := strconv.Atoi(expDays); err == nil && days > 0 {
				return days
			}
		}
	}

	// 从数据库获取
	setting, _ := models.GetSettingByKey(constant.SiteSettingSectionName, "cookie_exp_days")
	if setting.ID > 0 && setting.Value != "" {
		if days, err := strconv.Atoi(setting.Value); err == nil && days > 0 {
			return days
		}
	}

	// 返回默认值
	return 1
}

// ValidateDiffSetting 校验不同的设置
func (us *UserSettings) ValidateDiffSetting(section string, data map[string]string) string {
	if section == constant.SiteSettingSectionName {
		var config SiteConfig
		config.Title = data["title"]
		config.Slogan = data["slogan"]
		config.Logo = data["logo"]
		config.CookieExpDays = data["cookie_exp_days"]
		config.ThemeColor = data["theme_color"]
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	if section == constant.LogsCleanSectionName {
		var config LogConfig
		config.Cron = data["cron"]
		config.KeepNum = data["keep_num"]
		config.Enabled = data["enabled"]
		_, err := cron.ParseStandard(config.Cron)
		if err != nil {
			return fmt.Sprintf("%s 不是合法的corn表达式", config.Cron)
		}
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	if section == constant.HostedMsgCleanSectionName {
		var config HostedMsgConfig
		config.Cron = data["cron"]
		config.KeepNum = data["keep_num"]
		config.Enabled = data["enabled"]
		_, err := cron.ParseStandard(config.Cron)
		if err != nil {
			return fmt.Sprintf("%s 不是合法的corn表达式", config.Cron)
		}
		_, errStr := app.CommonPlaygroundValid(config)
		return errStr
	}
	return fmt.Sprintf("未知的section：%s", section)
}
