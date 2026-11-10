# 配置说明

## 完整配置文件

```ini
[app]
JwtSecret = message-nest
; Required for the Aliyun SMS test API; use a stable random value of at least 32 bytes.
; Keep identical on every replica. Alternatively set ALIYUN_SMS_TEST_FINGERPRINT_SECRET.
AliyunSMSTestFingerprintSecret =
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
| AliyunSMSTestFingerprintSecret | 阿里云短信测试限流指纹密钥；启用测试接口时必填，至少32字节，所有副本必须一致 | 空（测试接口返回503） |
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
| ALIYUN_SMS_TEST_FINGERPRINT_SECRET | 阿里云短信测试限流指纹密钥；显式配置、至少32字节、所有副本一致；缺失时测试接口返回503 |
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

### 阿里云短信测试限流密钥

阿里云短信测试接口只有在 `AliyunSMSTestFingerprintSecret` 或环境变量
`ALIYUN_SMS_TEST_FINGERPRINT_SECRET` 提供至少 32 字节的稳定随机密钥时才会工作；
缺失或过短时接口 fail closed 并返回 503。所有副本必须使用同一密钥。密钥不写入数据库或日志。

轮换时先在所有副本关闭/阻止短信测试 API，等待至少 60 秒（覆盖 60 秒配额窗口和 20 秒重复窗口），
再同时更新密钥并重启所有副本，确认全部使用新密钥后再开放测试 API。不要在请求仍可进入时逐台滚动换钥，
否则新旧指纹会被视为不同用户/请求并暂时分裂额度。

限流专用表由应用启动时独立、同步地创建/校验，只会触及 `aliyun_sms_test_scope_lock` 和
`aliyun_sms_test_event`（加上配置的 `TablePrefix`）；迁移错误会阻止服务启动。回滚应用时保留这两张表，
应用不会自动删除它们。SQLite 多进程场景必须共享同一台机器上的同一个本地数据库文件。

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
