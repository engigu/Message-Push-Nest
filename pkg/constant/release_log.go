package constant

import "strings"

var LatestVersion = map[string]string{}

var V1Version = "v0.0.1"
var V1VersionDesc = `
1. 单应用打包
2. 支持邮件发送
3. 用户密码设置
4. 支持用户定时任务清理，更新定时时间
5. 查看定时清理日志
6. 单应用的html浏览器自动缓存
7. gin的日志使用logrus
8. 支持异步发送
`

var V2Version = "v0.0.2"
var V2VersionDesc = `
1. 单应用打包
2. 支持邮件发送
3. 用户密码设置
4. 支持用户定时任务清理，更新定时时间
5. 查看定时清理日志
6. 单应用的html浏览器自动缓存
7. gin的日志使用logrus
8. 支持异步发送
9. 支持钉钉消息推送
`

func init() {
	LatestVersion["version"] = V2Version
	LatestVersion["desc"] = strings.TrimSpace(V2VersionDesc)
}
