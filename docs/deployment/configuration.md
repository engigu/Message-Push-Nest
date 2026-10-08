# 配置说明

## 完整配置文件

```ini
[app]
JwtSecret = message-nest
LogLevel = INFO

[server]
HttpPort = 8000
ReadTimeout = 60
WriteTimeout = 60
; URL路径前缀，用于子路径部署，如：/message
; UrlPrefix = /message

[database]
; 配置使用什么数据库，支持：mysql、sqlite、tidb
Type = mysql
User = root
Password = password
Host = 123.1.1.1
Name = db_name
Port = 3306
; -- 其他配置
; 表前缀
TablePrefix = message_
; -- 是否打开sql打印
; SqlDebug = enable
; 数据库连接是否开启ssl, value: [false | true]
Ssl = true
```

## 配置项说明

### [app] 应用配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| JwtSecret | JWT密钥，用于token生成 | message-nest |
| LogLevel | 日志级别：DEBUG/INFO/ERROR | INFO |

### [server] 服务器配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| HttpPort | HTTP服务端口 | 8000 |
| ReadTimeout | 读取超时时间（秒） | 60 |
| WriteTimeout | 写入超时时间（秒） | 60 |
| UrlPrefix | URL路径前缀，用于子路径部署，如：/message | - |

### [database] 数据库配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Type | 数据库类型：mysql/sqlite/tidb | mysql |
| User | 数据库用户名 | - |
| Password | 数据库密码 | - |
| Host | 数据库主机地址 | - |
| Port | 数据库端口 | 3306 |
| Name | 数据库名称 | - |
| TablePrefix | 数据表前缀 | message_ |
| SqlDebug | 是否打印SQL，设置enable开启 | - |
| Ssl | 数据库连接是否开启SSL | false |

## Docker 环境变量

使用Docker部署时，可以通过环境变量进行配置：

| 变量 | 说明 |
|------|------|
| JWT_SECRET | jwt秘钥，可选，默认为message-nest |
| LOG_LEVEL | 日志等级，可选，默认为INFO，DEBUG/INFO/ERROR |
| URL_PREFIX | URL路径前缀，用于子路径部署，如：/message |
| DB_TYPE | 数据库类型，sqlite/mysql/postgres。默认为sqlite,存储路径为conf/database.db |
| MYSQL_HOST | mysql-host，DB_TYPE=mysql必填 |
| MYSQL_PORT | mysql端口，DB_TYPE=mysql必填 |
| MYSQL_USER | mysql用户名，DB_TYPE=mysql必填 |
| MYSQL_PASSWORD | mysql数据库密码，DB_TYPE=mysql必填 |
| MYSQL_DB | mysql数据库名字，DB_TYPE=mysql必填 |
| MYSQL_TABLE_PREFIX | mysql数据表前缀，可选 |
| POSTGRES_HOST | postgres-host，DB_TYPE=postgres必填 |
| POSTGRES_PORT | postgres端口，DB_TYPE=postgres必填 |
| POSTGRES_USER | postgres用户名，DB_TYPE=postgres必填 |
| POSTGRES_PASSWORD | postgres数据库密码，DB_TYPE=postgres必填 |
| POSTGRES_DB | postgres数据库名字，DB_TYPE=postgres必填 |
| SSL | 是否开启SSL |
| SQL_DEBUG | 是否打印SQL，可选，默认关，设置enable为开启 |

## 配置示例

### 生产标准配置

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

Type = mysql
User = root
Password = Aa123456
Host = vm.server
Port = 3308
Name = yourDbName
TablePrefix = message_
```

### 开发调试配置（开启 SQL 打印）

```ini
[app]
JwtSecret = message-nest
LogLevel = INFO

[server]
HttpPort = 8000
ReadTimeout = 60
WriteTimeout = 60

[database]
; 开启SQL打印
SqlDebug = enable

Type = mysql
User = root
Password = Aa123456
Host = vm.server
Port = 3308
Name = yourDbName
TablePrefix = message_
```

### SQLite 配置

```ini
[app]
JwtSecret = message-nest
LogLevel = INFO

[server]
HttpPort = 8000
ReadTimeout = 60
WriteTimeout = 60

[database]
Type = sqlite
TablePrefix = message_
```
