# Docker Compose 部署

使用 Docker Compose 编排部署 Message Nest。

::: tip 推荐指数
🍀🍀🍀🍀🍀 适合生产环境，便于管理
:::

## 方式一：挂载配置文件

### 1. 准备配置文件

创建 `conf/app.ini`，内容参考 [Docker部署](/deployment/docker#_1-准备配置文件)。

### 2. 创建 docker-compose.yml

```yaml
version: "3.7"
services:

  message-nest:
    image: ghcr.io/engigu/message-nest:latest
    container_name: message-nest
    restart: always
    volumes:
      - ./conf:/app/conf
    ports:
      - "8000:8000"
```

### 3. 文件目录结构

```
.
├── conf
│   └── app.ini
└── docker-compose.yml
```

### 4. 启动服务

```bash
# 测试运行
docker-compose up

# 正式运行（后台）
docker-compose up -d
```

## 方式二：环境变量（推荐）

### 使用 MySQL

::: warning 重要
使用 MySQL 或是 PostgreSQL 时必须分别指定 `DB_TYPE=mysql` 或 `DB_TYPE=postgres` 环境变量，否则会默认使用 SQLite。
:::

创建 `docker-compose.yml`：

```yaml
version: "3.7"
services:

  message-nest:
    image: ghcr.io/engigu/message-nest:latest
    container_name: message-nest
    restart: always
    ports:
      - "8000:8000"
    environment:
      - DB_TYPE=mysql
      - MYSQL_HOST=192.168.64.133
      - MYSQL_PORT=3308
      - MYSQL_USER=root
      - MYSQL_PASSWORD=Aa123456
      - MYSQL_DB=test_11
      - MYSQL_TABLE_PREFIX=message_
```

### 使用 SQLite

创建 `docker-compose.yml`：

```yaml
version: "3.7"
services:

  message-nest:
    image: ghcr.io/engigu/message-nest:latest
    container_name: message-nest
    restart: always
    ports:
      - "8000:8000"
    volumes:
      - ./data/database.db:/app/conf/database.db
```

### 启动服务

```bash
# 正式运行
docker-compose up -d
```

## 完整示例：MySQL + Message Nest

如果你还没有MySQL，可以使用以下配置同时部署MySQL和Message Nest：

```yaml
version: "3.7"
services:

  mysql:
    image: mysql:8.0
    container_name: message-nest-mysql
    restart: always
    environment:
      - MYSQL_ROOT_PASSWORD=Aa123456
      - MYSQL_DATABASE=message_nest
    volumes:
      - mysql-data:/var/lib/mysql
    ports:
      - "3306:3306"

  message-nest:
    image: ghcr.io/engigu/message-nest:latest
    container_name: message-nest
    restart: always
    depends_on:
      - mysql
    ports:
      - "8000:8000"
    environment:
      - DB_TYPE=mysql
      - MYSQL_HOST=mysql
      - MYSQL_PORT=3306
      - MYSQL_USER=root
      - MYSQL_PASSWORD=Aa123456
      - MYSQL_DB=message_nest
      - MYSQL_TABLE_PREFIX=message_

volumes:
  mysql-data:
```

## 常用命令

### 启动服务

```bash
# 前台启动（查看日志）
docker-compose up

# 后台启动
docker-compose up -d
```

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs message-nest

# 实时查看日志
docker-compose logs -f message-nest
```

### 停止服务

```bash
# 停止服务
docker-compose stop

# 停止并删除容器
docker-compose down

# 停止并删除容器和数据卷
docker-compose down -v
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart message-nest
```

### 更新镜像

```bash
# 拉取最新镜像
docker-compose pull

# 重新创建容器
docker-compose up -d
```

## 访问服务

启动后访问 `http://localhost:8000`

- 默认账号：`admin`
- 默认密码：初始化时随机生成并打印在控制台日志中

## 常见问题

### 服务无法启动

1. 检查端口是否被占用
2. 检查配置是否正确
3. 查看日志：`docker-compose logs`

### MySQL连接失败

1. 确保MySQL服务已启动
2. 检查 `MYSQL_HOST` 是否正确（使用服务名）
3. 等待MySQL完全启动（约10-30秒）

### 数据持久化

使用volumes确保数据持久化：
```yaml
volumes:
  - ./data:/app/conf  # 配置文件
  - mysql-data:/var/lib/mysql  # MySQL数据
```

### 修改配置后重启

```bash
# 修改配置文件或环境变量后
docker-compose down
docker-compose up -d
```
