# Message Nest 🕊️

Message Nest 是一个灵活而强大的消息推送整合平台，旨在简化并自定义多种消息通知方式。

项目名叫信息巢，意思是一个拥有各种渠道信息方式的集合站点。

如果你有很多消息推送方式，每次都需要调用各种接口去发送消息，这个项目可以帮你管理各种消息方式，并提供统一的api接入。你可以自由组合各种方式，一个api推送到各种渠道，帮你省去接入的繁琐步骤。

## 特色 ✨

- 🔄 **整合性：** 提供了多种消息推送方式，包括邮件、钉钉、企业微信等，方便你集中管理和定制通知。
- 🎨 **自定义性：** 可以根据需求定制消息推送策略，满足不同场景的个性化需求。
- 🛠 **开放性：** 易于扩展和集成新的消息通知服务，以适应未来的变化。


## 进度 🔨
项目还在不断更新中，欢迎大家提出各种建议。

关于日志，考虑到目前多数服务以收集控制台输出为主，暂时不支持写出日志文件。

2024.01.07
- [x] 支持站点信息自定义

----
2024.01.03
- [x] 支持企业微信

----

- [x] 单应用打包
- [x] 支持邮件发送
- [x] 用户密码设置
- [x] 支持用户定时任务清理，更新定时时间
- [x] 查看定时清理日志
- [x] 单应用的html浏览器自动缓存
- [x] gin的日志使用logrus
- [x] 支持异步发送
- [x] 支持邮件发送
- [x] 支持钉钉
- [x] 支持自定义的webhook消息发送
- [x] 企业微信
- [ ] ....

## 项目来由 💡
自己常常写一些脚本需要消息推送，经常需要接入不同的消息发送，很不方便，于是就有了这个项目。

## 效果图 📺
![image](https://raw.githubusercontent.com/engigu/resources/images/2024/01/18/d1cdd1be348351b49ae9f1d90e1098c1.gif)

## 使用方法 🚀

1. 下载最新的系统版本对应的release， 解压
2. 新建一个数据库
3. 重命名conf/app.example.ini为conf/app.ini
4. 修改app.ini对应的配置
5. 启动项目会自动创建表和账号 
```shell
# 第一次运行将app.ini中的app.InitData设置为enable，会自动进行表数据的初始化
# 后续不需要开启这个配置
# INFO登录启动回出现如下日志

[2024-01-13 13:40:09.075]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_auth
[2024-01-13 13:40:11.778]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_send_tasks
[2024-01-13 13:40:16.518]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_send_ways
[2024-01-13 13:40:23.300]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_send_tasks_logs
[2024-01-13 13:40:28.715]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_send_tasks_ins
[2024-01-13 13:40:39.538]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_settings
[2024-01-13 13:40:46.299]  INFO [migrate.go:74 Setup] [Init Data]: Init Account data...
[2024-01-13 13:40:46.751]  INFO [migrate.go:77 Setup] [Init Data]: All table data init done.

```
6. 启动项目，访问8000端口，初始账号为admin，密码为123456


## 贡献 🤝
欢迎通过提交问题和提出改进建议。

## 致谢 🙏
该项目汲取了[go-gin-example](https://github.com/eddycjy/go-gin-example)项目的灵感，展示了 Go 和 Gin 在实际应用中的强大和多才多艺。

## 许可证 📝
[LICENSE](LICENSE)

