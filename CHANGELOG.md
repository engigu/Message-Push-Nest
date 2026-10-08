# Release Notes (v2.2.6)

## 🚀 重大更新与改进

### 📦 CI/CD 与全平台发布
- **构建架构重构**：统一收敛所有构建逻辑至脚本驱动，支持灵活分模块构建与多平台编译。
- **全平台二进制矩阵**：补齐 `Darwin (amd64/arm64)`、`Linux (amd64/arm64/armv7)`、`FreeBSD (amd64/arm64)`、`OpenBSD (amd64/arm64)` 以及 `Windows (amd64/arm64)` 全套 11 架构发布资产，并自动导出 `checksums.txt` 校验和。
- **双仓库同步发布**：实现 GitHub Container Registry (`ghcr.io`) 与 Docker Hub (`docker.io`) 的多架构镜像原子化同步发布。

### 🐳 容器镜像极简优化
- **Alpine 底包改造**：运行底包由 `debian:bookworm-slim` 切换为 `alpine:3.21`，底包体积从 ~100MB 骤降至 ~10MB，最终镜像总体积缩减 70%+（仅约 ~35MB）。
- **运行环境完备性**：保留完整 `ca-certificates`（确保 HTTPS 消息推送）与 `tzdata` 时区配置，并内置 `sh` 命令行工具便于调试与运维。

### ⚡ 前端性能与 UI 优化
- **打包策略优化**：配置 Vite/Rolldown 动态拆包（提取 `vendor-core` 与 `vendor-ui`），提升长效缓存利用率，消除了静态资源 Chunk 体积警告。
- **安全与排版**：全面升级依赖库消除潜在漏洞，加固 JWT 密钥防护与敏感配置脱敏，美化设置页暗色模式与 Inter 字体排版。
