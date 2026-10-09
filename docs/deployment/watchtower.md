# 自动更新 (Watchtower)

[Watchtower](https://containrrr.dev/watchtower/) 是一个开源的 Docker 镜像自动更新工具。它可以监控运行中的容器，当发现远程仓库（如 `ghcr.io/engigu/message-nest:latest`）发布了新版本时，会自动拉取最新镜像，并以原有配置平滑重启容器。

::: tip 推荐配合使用
对于使用 `latest` 标签部署的 Docker / Docker Compose 用户，配置 Watchtower 可以实现 Message Nest 的免运维自动更新。
:::

---

## 方式一：Docker 独立运行

如果已经通过 `docker run` 启动了 Message Nest，可以直接运行一个 Watchtower 容器进行监控。

### 1. 监控特定容器并自动更新

以下命令将每小时检查一次 `message-nest` 容器的更新，并在更新后清理旧镜像：

```bash
docker run -d \
  --name watchtower \
  --restart always \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower \
  --interval 3600 \
  --cleanup \
  message-nest
```

参数说明：
- `-v /var/run/docker.sock:/var/run/docker.sock`：挂载 Docker 守护进程通信接口。
- `--interval 3600`：检查更新的时间间隔（单位：秒），此处为 1 小时。
- `--cleanup`：更新后自动删除旧镜像，释放磁盘空间。
- `message-nest`：指定仅监控并更新名为 `message-nest` 的容器（若不指定则监控全部容器）。

### 2. 单次手动触发更新（运行后销毁）

如果不想让 Watchtower 在后台常驻，仅想手动一键更新：

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  containrrr/watchtower \
  --run-once \
  --cleanup \
  message-nest
```

---

## 方式二：Docker Compose 集成（推荐）

在使用 Docker Compose 部署时，可以将 Watchtower 作为一个独立服务写入 `docker-compose.yml` 中。

推荐使用**标签模式（Label）**，仅对明确标记的容器生效，防止意外升级其它基础服务（如 MySQL）。

### 示例配置

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
    # 启用 Watchtower 监控更新
    labels:
      - "com.centurylinklabs.watchtower.enable=true"

  watchtower:
    image: containrrr/watchtower
    container_name: watchtower
    restart: always
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      # 仅更新带有 com.centurylinklabs.watchtower.enable=true 标签的容器
      - WATCHTOWER_LABEL_ENABLE=true
      # 检查周期（秒），如 86400 为每天检查一次
      - WATCHTOWER_POLL_INTERVAL=86400
      # 自动清理旧镜像
      - WATCHTOWER_CLEANUP=true
      # 时区配置
      - TZ=Asia/Shanghai
```

启动命令：

```bash
docker-compose up -d
```

---

## 常用参数与环境变量

| 参数选项 | 环境变量 | 说明 | 示例 |
|---|---|---|---|
| `--interval` | `WATCHTOWER_POLL_INTERVAL` | 轮询检查镜像的时间间隔（单位：秒） | `3600`（1小时） |
| `--schedule` | `WATCHTOWER_SCHEDULE` | 使用 6 段式 Cron 表达式指定检查时间 | `0 0 4 * * *`（每天凌晨 4 点） |
| `--cleanup` | `WATCHTOWER_CLEANUP` | 更新后自动删除遗留的旧镜像 | `true` |
| `--label-enable` | `WATCHTOWER_LABEL_ENABLE` | 仅监控包含 enable 标签的容器 | `true` |
| `--run-once` | `WATCHTOWER_RUN_ONCE` | 仅执行一次检查并更新，随后退出容器 | 命令行测试常用 |

---

## 注意事项

::: warning 数据持久化提示
由于自动更新会重新创建容器，请**务必确保数据持久化**已正确配置（如挂载 SQLite 数据库文件 `./data/database.db:/app/conf/database.db` 或使用外部独立数据库 MySQL/PostgreSQL），避免因容器重建造成数据丢失。
:::

::: tip 私有镜像源说明
Message Nest 镜像公开托管在 GitHub Container Registry，拉取公开镜像无需配置凭据。如果服务器拉取 GitHub 镜像网络较慢，可考虑配置 Docker 镜像加速。
:::
