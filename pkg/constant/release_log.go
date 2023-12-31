package constant

var LatestVersion = map[string]string{}

//var info map[string]string

var V1Version = "v1.0.0"
var V1VersionDesc = `1. 支持邮件消息发送
2. 支持日志定时删除
3. 支持账号密码重置`

func init() {
	LatestVersion["version"] = V1Version
	LatestVersion["desc"] = V1VersionDesc
}
