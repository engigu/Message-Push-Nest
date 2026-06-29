# Docker 部署

使用 Docker 部署 Message Nest，支持多种配置方式。

::: tip 推荐指数
🍀🍀🍀🍀🍀 最推荐的部署方式
:::

## 镜像源

Message Nest 使用 GitHub Container Registry 托管镜像：

- **镜像地址**: `ghcr.io/engigu/message-nest:latest`

## 方式一：挂载配置文件

### 1. 准备配置文件

新建 `conf/app.ini` 文件：

```ini
[app]
JwtSecret = message-nest
LogLevel = INFO

[server]
RunMode = release
; docker模式下端口配置文件中只能为8000
HttpPort = 8000
ReadTimeout = 60
WriteTimeout = 60
; 是否使用embed打包的静态资源
; 如果运行release打包后的应用，请注释这个设置。
; 如果取消这个注释，只会单独运行api服务，前端页面需要到web目录手动npm run dev, 运行前端服务
; EmbedHtml = disable
; URL路径前缀，用于子路径部署，如：/message
; UrlPrefix = /message

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

::: warning 端口限制
Docker模式下，配置文件中的端口只能为8000，通过 `-p` 参数映射到宿主机端口。
:::

### 2. 拉取镜像

```bash
docker pull ghcr.io/engigu/message-nest:latest
```

### 3. 启动容器

```bash
# 测试运行
docker run --rm -ti \
  -p 8000:8000 \
  -v /your/path/conf:/app/conf \
  ghcr.io/engigu/message-nest:latest 
  
# 正式运行
docker run -d \
  -p 8000:8000 \
  -v /your/path/conf:/app/conf \
  ghcr.io/engigu/message-nest:latest 
```

## 方式二：环境变量（推荐）

::: tip 推荐
这是最简单的部署方式，无需准备配置文件。
:::

### 环境变量说明

| 变量 | 说明 |
|------|------|
| JWT_SECRET | jwt秘钥，可选，默认为message-nest |
| LOG_LEVEL | 日志等级，可选，默认为INFO，DEBUG/INFO/ERROR |
| RUN_MODE | 运行模式，可选，默认release，为debug将自动添加跨域 |
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

### 使用 MySQL

::: warning 重要
使用 MySQL 或是 PostgreSQL 时必须分别指定 `DB_TYPE=mysql` 或 `DB_TYPE=postgres` 环境变量，否则会默认使用 SQLite。
:::

```bash
docker run -d  \
  -p 8000:8000 \
  -e DB_TYPE=mysql \
  -e MYSQL_HOST=192.168.64.133  \
  -e MYSQL_PORT=3308 \
  -e MYSQL_USER=root \
  -e MYSQL_PASSWORD=Aa123456 \
  -e MYSQL_DB=test_11 \
  -e MYSQL_TABLE_PREFIX=message_ \
  --name message-nest  \
  ghcr.io/engigu/message-nest:latest 
```

### 使用 SQLite

```bash
docker run -d  \
  -p 8000:8000 \
  -v /your/path/database.db:/app/conf/database.db  \
  --name message-nest  \
  ghcr.io/engigu/message-nest:latest 
```

## 访问服务

启动后访问 `http://localhost:8000`

- 默认账号：`admin`
- 默认密码：初始化时随机生成并打印在控制台日志中

## 常见问题

### 容器无法启动

1. 检查端口是否被占用
2. 检查数据库连接配置
3. 查看容器日志：`docker logs message-nest`

### 数据持久化

使用SQLite时，记得挂载数据库文件：
```bash
-v /your/path/database.db:/app/conf/database.db
```

### 查看日志

```bash
# 查看实时日志
docker logs -f message-nest

# 查看最近100行日志
docker logs --tail 100 message-nest
```

### 停止和删除容器

```bash
# 停止容器
docker stop message-nest

# 删除容器
docker rm message-nest
```
