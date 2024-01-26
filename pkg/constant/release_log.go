package constant

import "strings"

var LatestVersion = map[string]string{}

//var V1Version = "v0.0.1"
//var V1VersionDesc = `
//1. 单应用打包
//2. 支持邮件发送
//3. 用户密码设置
//4. 支持用户定时任务清理，更新定时时间
//5. 查看定时清理日志
//6. 单应用的html浏览器自动缓存
//7. gin的日志使用logrus
//8. 支持异步发送
//`
//
//var V2Version = "v0.0.2"
//var V2VersionDesc = `
//1. 单应用打包
//2. 支持邮件发送
//3. 用户密码设置
//4. 支持用户定时任务清理，更新定时时间
//5. 查看定时清理日志
//6. 单应用的html浏览器自动缓存
//7. gin的日志使用logrus
//8. 支持异步发送
//9. 支持钉钉消息推送
//`
//
//var V3Version = "v0.0.3"
//var V3VersionDesc = `
//1. 单应用打包
//2. 支持邮件发送
//3. 用户密码设置
//4. 支持用户定时任务清理，更新定时时间
//5. 查看定时清理日志
//6. 单应用的html浏览器自动缓存
//7. gin的日志使用logrus
//8. 支持异步发送
//9. 支持钉钉消息推送
//10. 支持自定义的webhook推送
//`
//
//var V4Version = "v0.0.4"
//var V4VersionDesc = `
//1. 单应用打包
//2. 支持邮件发送
//3. 用户密码设置
//4. 支持用户定时任务清理，更新定时时间
//5. 查看定时清理日志
//6. 单应用的html浏览器自动缓存
//7. gin的日志使用logrus
//8. 支持异步发送
//9. 支持钉钉消息推送
//10. 支持自定义的webhook推送
//11. 支持自动初始化表，以及初始化账号
//`

//var V5Version = "v0.0.5"
//var V5VersionDesc = `
//1. 单应用打包
//2. 支持邮件发送
//3. 用户密码设置
//4. 支持用户定时任务清理，更新定时时间
//5. 查看定时清理日志
//6. 单应用的html浏览器自动缓存
//7. gin的日志使用logrus
//8. 支持异步发送
//9. 支持钉钉消息推送
//10. 支持自定义的webhook推送
//11. 支持自动初始化表，以及初始化账号
//12. 调整日志格式
//13. 支持更多的api接入示例
//`

var LatestVersionS = "v0.1.0"
var LatestVersionDesc = `
1. 单应用打包
2. 支持邮件发送
3. 用户密码设置
4. 支持用户定时任务清理，更新定时时间
5. 查看定时清理日志
6. 单应用的html浏览器自动缓存
7. gin的日志使用logrus
8. 支持异步发送
9. 支持钉钉消息推送
10. 支持自定义的webhook推送
11. 支持自动初始化表，以及初始化账号
12. 调整日志格式
13. 支持更多的api接入示例
14. 支持发送实例的暂停与开启
15. 支持企业微信消息发送
16. 支持站点信息自定义
17. 支持数据统计展示
`

func init() {
	LatestVersion["version"] = LatestVersionS
	LatestVersion["desc"] = strings.TrimSpace(LatestVersionDesc)
}
