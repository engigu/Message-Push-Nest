# 直接运行

使用最新的Release打包的可执行文件部署，无需部署前端页面。

::: tip 推荐指数
🍀🍀🍀🍀 适合没有Docker环境的生产部署
:::

## 部署步骤

### 1. 下载Release

访问 [GitHub Releases](https://github.com/engigu/Message-Push-Nest/releases) 下载最新的系统版本对应的release，然后解压。

### 2. 创建数据库

新建一个MySQL数据库（或使用SQLite）。

### 3. 配置文件

新建conf文件夹，或者重命名项目中 `conf/app.example.ini` 为 `conf/app.ini`，然后修改配置：

```ini
[app]
JwtSecret = message-nest
LogLevel = INFO

[server]
HttpPort = 8000
ReadTimeout = 60
WriteTimeout = 60

[database]
; 关闭SQL打印
; SqlDebug = enable

; Type = sqlite
Type = mysql
User = root
Password = Aa123456
Host = vm.server
Port = 3308
Name = yourDbName
TablePrefix = message_
```

### 4. 启动项目

直接运行可执行文件，项目会自动创建表和账号。

```bash
# Windows
./Message-Nest.exe

# Linux/Mac
./Message-Nest
```

### 5. 查看日志

INFO日志级别启动会出现如下日志：

```log
[2024-01-13 13:40:09.075]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_auth
[2024-01-13 13:40:11.778]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_send_tasks
[2024-01-13 13:40:16.518]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_send_ways
[2024-01-13 13:40:23.300]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_send_tasks_logs
[2024-01-13 13:40:28.715]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_send_tasks_ins
[2024-01-13 13:40:39.538]  INFO [migrate.go:70 Setup] [Init Data]: Migrate table: message_settings
[2024-01-13 13:40:46.299]  INFO [migrate.go:74 Setup] [Init Data]: Init Account data...
[2024-01-13 13:40:46.751]  INFO [migrate.go:77 Setup] [Init Data]: All table data init done.
```

### 6. 访问服务

访问 `http://localhost:8000`

- 默认账号：`admin`
- 默认密码：初始化时随机生成并打印在控制台日志中

## 使用SQLite

如果不想安装MySQL，可以使用SQLite：

```ini
[database]
Type = sqlite
TablePrefix = message_
```

SQLite数据库文件会自动创建在 `conf/database.db`。

## 常见问题

### 启动失败

1. 检查端口8000是否被占用
2. 检查数据库连接配置是否正确
3. 查看日志输出的错误信息

### 无法访问页面

1. 确认服务已正常启动
2. 检查防火墙及网络监听端口设置
