# 部署概览

Message Nest 提供多种部署方式，您可以根据自己的需求选择合适的部署方案。

## 部署方式对比

| 部署方式 | 难度 | 推荐指数 | 适用场景 |
|---------|------|---------|---------|
| Docker环境变量 | ⭐ | ⭐⭐⭐⭐⭐ | 生产环境，快速部署 |
| Docker Compose | ⭐ | ⭐⭐⭐⭐⭐ | 生产环境，编排部署 |
| 直接运行Release | ⭐⭐ | ⭐⭐⭐⭐ | 生产环境，无Docker环境 |
| 开发调试 | ⭐⭐⭐ | ⭐⭐⭐ | 开发环境 |

## 快速选择

### 我想快速体验
推荐使用 **Docker环境变量部署**，一条命令即可启动。

### 我要用于生产环境
推荐使用 **Docker Compose部署**，便于管理和维护。

### 我没有Docker环境
推荐使用 **直接运行Release**，下载可执行文件即可。

### 我要进行开发
推荐使用 **开发调试模式**，支持热更新。

## 默认账号

所有部署方式启动后，默认账号信息：
- 用户名：`admin`
- 密码：`123456`

::: warning 安全提示
首次登录后请立即修改默认密码！
:::

## 数据库选择

Message Nest 支持多种数据库：

- **SQLite** - 轻量级，无需额外配置，适合小规模使用
- **MySQL 5.x / 8.x** - 成熟稳定，适合中大规模使用
- **TiDB** - 分布式数据库，适合大规模使用

## 下一步

选择适合您的部署方式：

- [直接运行](/deployment/direct-run)
- [开发调试](/deployment/development)
- [Docker](/deployment/docker)
- [Docker Compose](/deployment/docker-compose)
- [Nginx](/deployment/nginx)
