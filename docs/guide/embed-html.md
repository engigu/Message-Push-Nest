# EmbedHtml 配置说明

## 什么是 EmbedHtml？

这个配置可以理解为单应用模式（或者前后端分离）的开关。

## 配置方式

### 方式一：单应用模式（推荐）

**注释 EmbedHtml 配置**

```ini
[server]
RunMode = release
HttpPort = 8000
ReadTimeout = 60
WriteTimeout = 60
; 注释EmbedHtml，启用单应用模式
; EmbedHtml = disable
```

**说明：**
- 启动go服务，会把web/dist目录下文件作为前端静态资源
- 如果目录下没有静态资源文件，需要到web目录下，执行 `npm run build` 构建生成
- 只需要运行一个服务即可访问完整功能
- 适合生产环境部署

**适用场景：**
- 使用 Release 打包的可执行文件
- Docker 部署
- 生产环境

### 方式二：前后端分离模式

**启用 EmbedHtml 配置**

```ini
[server]
RunMode = debug
HttpPort = 8000
ReadTimeout = 60
WriteTimeout = 60
; 取消注释，启用前后端分离
EmbedHtml = disable
```

**说明：**
- go服务启动时只会有API服务
- 需要到web目录下，执行 `npm run dev` 启动前端项目
- 访问前端项目提示的端口服务，一般是 `http://127.0.0.1:5173`
- 或者使用 `npm run build`，用Nginx部署前端

**适用场景：**
- 开发调试
- 需要修改前端代码
- 前后端独立部署

## 两种方式对比

| 特性 | 单应用模式 | 前后端分离模式 |
|------|-----------|--------------|
| 部署复杂度 | 简单，只需一个服务 | 复杂，需要两个服务 |
| 开发调试 | 不便，需要重新构建 | 方便，支持热更新 |
| 资源占用 | 较少 | 较多 |
| 适用环境 | 生产环境 | 开发环境 |
| 推荐指数 | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ |

## 推荐方案

综合考虑下来，推荐直接使用以下方式：

1. **Release 打包执行文件** - 已内置页面静态资源，开箱即用
2. **Docker 环境变量部署** - 配置简单，一键启动

这两种方式都是单应用模式，只需运行一个服务即可。

## 常见问题

### Q: 为什么访问页面显示404？

A: 检查以下几点：
1. 如果是单应用模式，确保 `EmbedHtml` 已注释
2. 确保 `web/dist` 目录下有静态资源文件
3. 如果没有，执行 `cd web && npm run build` 构建

### Q: 开发时如何快速调试？

A: 
1. 设置 `RunMode = debug`
2. 启用 `EmbedHtml = disable`
3. 启动后端：`go run main.go`
4. 启动前端：`cd web && npm run dev`
5. 访问前端提示的URL（通常是 http://127.0.0.1:5173）

### Q: Docker 部署需要配置 EmbedHtml 吗？

A: 不需要。Docker 镜像已经内置了前端静态资源，默认就是单应用模式。
