---
layout: home

hero:
  name: "Message Nest"
  text: "消息推送整合平台"
  tagline: 灵活而强大的消息推送整合平台，简化并自定义多种消息通知方式
  image:
    src: /logo.svg
    alt: Message Nest
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/introduction
    - theme: alt
      text: 部署指南
      link: /deployment/overview
    - theme: alt
      text: GitHub
      link: https://github.com/engigu/Message-Push-Nest

features:
  - icon: 🔄
    title: 整合性
    details: 提供了多种消息推送方式，包括邮件、钉钉、企业微信等，方便你集中管理和定制通知。
  - icon: 🎨
    title: 自定义性
    details: 可以根据需求定制消息推送策略，满足不同场景的个性化需求。
  - icon: 🛠
    title: 开放性
    details: 易于扩展和集成新的消息通知服务，以适应未来的变化。
  - icon: 📧
    title: 多渠道支持
    details: 支持邮件、钉钉、企业微信、微信测试公众号、自定义Webhook等多种推送方式。
  - icon: ⏰
    title: 定时任务
    details: 支持自定义的定时消息发送，满足定时推送需求。
  - icon: 🎯
    title: 自托管消息
    details: 可以将站点作为消息的接收方，登录站点查看消息。
  - icon: 🐳
    title: Docker支持
    details: 支持Docker和Docker Compose部署，快速启动服务。
  - icon: 💾
    title: 多数据库支持
    details: 支持SQLite、MySQL 5.x/8.x、TiDB等多种数据库。
  - icon: 📊
    title: 数据统计
    details: 支持数据统计展示，查看消息发送情况。
---

<!-- ## 快速开始

### Docker 部署（推荐）

使用环境变量快速启动：

```bash
docker run -d \
  -p 8000:8000 \
  -e MYSQL_HOST=192.168.64.133 \
  -e MYSQL_PORT=3308 \
  -e MYSQL_USER=root \
  -e MYSQL_PASSWORD=Aa123456 \
  -e MYSQL_DB=test_11 \
  -e MYSQL_TABLE_PREFIX=message_ \
  --name message-nest \
  engigu/message-nest:latest
```

### 直接运行

1. 下载最新的 [Release](https://github.com/engigu/Message-Push-Nest/releases)
2. 配置 `conf/app.ini`
3. 启动服务，访问 `http://localhost:8000`
4. 默认账号：`admin`，密码：`123456`

## 项目来由

自己常常写一些脚本需要消息推送，经常需要接入不同的消息发送，很不方便，于是就有了这个项目。

如果你有很多消息推送方式，每次都需要调用各种接口去发送消息到各个渠道，或者不同的项目你都需要复制同样的发消息代码，这个项目可以帮你管理各种消息方式，并提供统一的发送API接入。你可以自由组合各种消息渠道，一个API推送到各种渠道，帮你省去接入的繁琐步骤。

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=engigu/Message-Push-Nest&type=Date)](https://star-history.com/#engigu/Message-Push-Nest&Date) -->
