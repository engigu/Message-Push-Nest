# Message Nest 🕊️ 

[![Hits](https://hits.sh/github.com/engigu/Message-Push-Nest.svg?view=today-total)](https://hits.sh/github.com/engigu/Message-Push-Nest/)
![Latest Version](https://ghcr-badge.egpl.dev/engigu/message-nest/latest_tag?color=%2344cc11&ignore=latest%2Cmain*&label=docker+version&trim=)
![Image Size](https://ghcr-badge.egpl.dev/engigu/message-nest/size?color=%2344cc11&tag=latest&label=docker+image&trim=)
![Image pulls](https://img.shields.io/badge/dynamic/json?url=https://ghcr-badge.elias.eu.org/api/engigu/Message-Push-Nest/message-nest&query=downloadCount&style=flat&label=ghcr%20pulls&color=44cc11)
![Dockerhub Pulls](https://img.shields.io/docker/pulls/engigu/message-nest?style=flat&color=44cc11&label=dockerhub%20pulls)


Message Nest 是一个灵活而强大的消息推送整合平台，旨在简化并自定义多种消息通知方式。

项目名叫信息巢，意思是一个拥有各种渠道信息方式的集合站点。

如果你有很多消息推送方式，每次都需要调用各种接口去发送消息到各个渠道，或者不同的项目你都需要复制同样的发消息代码，这个项目可以帮你管理各种消息方式，并提供统一的发送api接入。你可以自由组合各种消息渠道，一个api推送到各种渠道，帮你省去接入的繁琐步骤。

演示站点(演示站点的服务器比较烂，见谅) [demo](https://message-nest-demo-site.qwapi.eu.org/)

如果项目有用，帮忙点个star。


## 特色 ✨

- 🔄 **整合性：** 提供了多种消息推送方式，包括邮件、钉钉、企业微信、飞书、Telegram 等，方便你集中管理和定制通知。
- 🎨 **自定义性：** 可以根据需求定制消息推送策略，满足不同场景的个性化需求。
- 📝 **模板化（⭐推荐）：** 支持消息模板功能，通过占位符实现动态内容替换，一次定义多处复用，大幅提高开发效率和维护便利性。
- 🛠 **开放性：** 易于扩展和集成新的消息通知服务，以适应未来的变化。

## 进度 🔨

项目还在不断更新中，欢迎大家提出各种建议。

关于运行日志，考虑到目前多数服务以收集控制台输出为主，暂时不支持写出日志文件。

这里自荐下自己开发，性能和占用都很能打的，定时任务执行面板 [baihu](https://github.com/engigu/baihu-panel/)

## 更新日志 ☕

### 最近更新

**2026.01.14** - 新增 Telegram 机器人渠道、支持 HTTP/SOCKS5 代理  
**2025.12.17** - 新增飞书机器人渠道、阿里云短信、代码架构优化  
**2025.12.06** - 新增消息模板功能、V2 API  
**2025.10.12** - 增加 cookies 过期天数设置  
**2025.09.30** - 支持明暗主题切换、登录日志  
**2025.08.10** - 重构 web 页面，UI 升级（shadcn-vue + tailwindcss）  
**2025.04.28** - 支持 TiDB、数据库 SSL 配置  

[查看完整更新日志](https://engigu.github.io/Message-Push-Nest/guide/changelog.html)
  
## 项目来由 💡

自己常常写一些脚本需要消息推送，经常需要接入不同的消息发送，很不方便，于是就有了这个项目。

## 效果图 📺

![CPT2501011740-1492x733](https://raw.githubusercontent.com/engigu/resources/refs/heads/images/CPT2508101507-1460x745.gif)


## 文档 📝

[点我跳转](https://engigu.github.io/Message-Push-Nest/)

## 贡献 🤝

欢迎通过提交问题和提出改进建议。

<img src="https://f.pz.al/pzal/2026/01/07/d0fad512d0346.jpg" style="width:200px;">

## Star History ⭐

[![Star History Chart](https://api.star-history.com/svg?repos=engigu/Message-Push-Nest&type=Date)](https://star-history.com/#engigu/Message-Push-Nest&Date)

## 许可证 📝

[LICENSE](LICENSE)

