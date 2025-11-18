# Message Nest 文档站点

这是 Message Nest 项目的 VitePress 文档站点。

## 本地开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm run docs:dev

# 构建文档
npm run docs:build

# 预览构建结果
npm run docs:preview
```

## 文档结构

```
docs/
├── .vitepress/          # VitePress 配置
│   └── config.mts       # 站点配置文件
├── guide/               # 指南
│   ├── introduction.md  # 介绍
│   ├── features.md      # 特色功能
│   ├── changelog.md     # 更新日志
│   ├── configuration.md # 配置说明
│   └── embed-html.md    # EmbedHtml说明
├── deployment/          # 部署
│   ├── overview.md      # 部署概览
│   ├── direct-run.md    # 直接运行
│   ├── development.md   # 开发调试
│   ├── docker.md        # Docker部署
│   └── docker-compose.md # Docker Compose
├── api/                 # API文档
│   ├── usage.md         # 使用说明
│   └── examples.md      # 调用示例
├── public/              # 静态资源
│   └── logo.svg         # Logo
└── index.md             # 首页
```

## 访问地址

开发服务器启动后，默认访问地址为：`http://localhost:5173`
