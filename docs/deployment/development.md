# 开发调试

本文档介绍如何在开发环境中运行 Message Nest。

## 前置要求

- Go 1.18+
- Node.js 16+
- MySQL 5.7+ 或 SQLite

## 部署步骤

### 1. 克隆项目

```bash
git clone https://github.com/engigu/Message-Push-Nest.git
cd Message-Push-Nest
```

### 2. 配置文件

重命名 `conf/app.example.ini` 为 `conf/app.ini`，关键配置如下：

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

; Type = sqlite
Type = mysql
User = root
Password = Aa123456
Host = vm.server
Port = 3308
Name = yourDbName
TablePrefix = message_
```

::: tip 调试建议
- `SqlDebug` 建议启用，方便排查 SQL 语句
- 前端 `npm run dev` 启动在 5173 端口，Vite 会自动将 API 请求反向代理至后端的 8000 端口，无需额外跨域或开关配置
:::

### 3. 启动后端服务

```bash
go mod tidy
go run main.go
```

服务启动后会运行在8000端口。

### 4. 启动前端服务

```bash
cd web
npm i
npm run dev
```

页面启动后会提示访问URL，一般是 `http://127.0.0.1:5173`。

### 5. 访问应用

访问 `http://127.0.0.1:5173`，进行调试开发。

接口会自动转发到go服务 `http://localhost:8000`。

## 开发说明

### 目录结构

```
Message-Push-Nest/
├── conf/           # 配置文件
├── middleware/     # 中间件
├── migrate/        # 数据库迁移
├── models/         # 数据模型
├── pkg/            # 工具包
├── routers/        # 路由
├── service/        # 业务逻辑
├── web/            # 前端项目
│   ├── src/        # 源代码
│   ├── public/     # 静态资源
│   └── dist/       # 构建输出
└── main.go         # 入口文件
```

### 前端技术栈

- Vue 3
- TypeScript
- Vite
- TailwindCSS
- shadcn-vue

### 后端技术栈

- Go
- Gin
- GORM
- Logrus

## 构建生产版本

### 构建前端

```bash
cd web
npm run build
```

构建产物会输出到 `web/dist` 目录。

### 构建后端

```bash
CGO_ENABLED=0 go build -o Message-Nest
```

## 常见问题

### 前端无法连接后端

检查 `web/vite.config.ts` 中的代理配置是否正确。

### 热更新不生效

1. 检查文件是否保存
2. 重启前端开发服务器
3. 清除浏览器缓存

### 数据库连接失败

1. 检查数据库服务是否启动
2. 检查配置文件中的连接信息
3. 检查数据库用户权限
