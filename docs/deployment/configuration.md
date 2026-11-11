# 配置说明

## 完整配置文件

```ini
[app]
JwtSecret = message-nest
; 阿里云短信测试限流指纹密钥；至少 32 字节高熵随机字符串，多副本部署必须一致
; 亦可通过环境变量 ALIYUN_SMS_TEST_FINGERPRINT_SECRET 注入
AliyunSMSTestFingerprintSecret =
LogLevel = INFO

[server]
HttpPort = 8000
ReadTimeout = 60
WriteTimeout = 60
; URL路径前缀，用于子路径部署，如：/message
; UrlPrefix = /message

[database]
; 项目数据库类型：mysql、sqlite、postgres、tidb
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
| Type | 数据库类型：mysql/sqlite/postgres/tidb | mysql |
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

服务启动引导期会同步执行数据库结构检查与迁移，限流模块包含通用排他锁表 `rate_limit_scope_lock` 与
滑动窗口事件表 `rate_limit_event`（加上配置的 `TablePrefix`），由系统迁移签名机制统一管理。回滚应用时保留这两张表，
应用不会自动删除它们。SQLite 多进程场景必须共享同一台机器上的同一个本地数据库文件。

发送测试会同步调用阿里云短信接口，可能产生短信费用；接口返回受理成功只表示供应商已受理，不表示最终送达。
请先确认测试手机号、签名、模板和变量内容，再发起测试。

### 共享限流数据库验证状态

项目的通用数据库支持说明不等于本功能已经通过每种数据库的真实集成验收。当前共享限流验证状态：

| 数据库 | 验证状态 | 验证版本/范围 |
|--------|----------|---------------|
| SQLite | PASS | 本地文件数据库回归及跨进程测试 |
| MySQL | PASS | MySQL Community Server 8.4.13，真实驱动、InnoDB、并发和完整路由 Mock 集成 |
| MariaDB | PASS | MariaDB 11.4.13，真实驱动、InnoDB、并发和完整路由 Mock 集成 |
| PostgreSQL | NOT VERIFIED | 尚未完成该功能的真实数据库集成验收 |
| TiDB | NOT VERIFIED | 尚未完成该功能的真实数据库集成验收 |

不要将 PostgreSQL/TiDB 的通用项目支持表述为共享限流功能的已验证兼容性，也不要据此推断生产级验收已完成。

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
