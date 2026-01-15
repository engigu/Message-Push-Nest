# URL 前缀配置

Message Nest 支持配置 URL 前缀，允许您将应用部署在子路径下（例如 `/message-nest`），而不是根路径。

## 工作原理

配置 URL 前缀后：
1. 后端会在所有路由前添加指定的前缀
2. 前端页面加载时，后端会在 HTML 中自动注入前缀配置
3. 前端的所有 API 请求会自动使用该前缀

**无需手动配置前端**，前后端会自动同步！

## 使用场景

- 在同一域名下部署多个应用
- 通过反向代理（如 Nginx）将应用映射到子路径
- 在企业内网中统一管理多个服务

## 配置方式

### 方式一：配置文件

编辑 `conf/app.ini` 文件，在 `[server]` 部分添加：

```ini
[server]
RunMode = release
HttpPort = 8000
UrlPrefix = /message-nest
```

### 方式二：环境变量

设置环境变量：

```bash
export URL_PREFIX=/message-nest
```

或在 Docker 中：

```bash
docker run -e URL_PREFIX=/message-nest ...
```

## 配置说明

- **路径格式**：可以带或不带前导斜杠 `/`，系统会自动处理
- **示例值**：
  - `/message-nest`
  - `/api/message-nest`
  - `message-nest`（会自动转换为 `/message-nest`）
- **默认值**：空字符串（部署在根路径）

## 访问方式

配置 URL 前缀后，访问地址会变为：

- **原地址**：`http://your-domain.com/`
- **新地址**：`http://your-domain.com/message-nest/`

所有 API 端点也会自动添加前缀：

- **原 API**：`http://your-domain.com/api/v1/sendways/list`
- **新 API**：`http://your-domain.com/message-nest/api/v1/sendways/list`

## Nginx 反向代理配置示例

如果使用 Nginx 反向代理，配置示例：

```nginx
location /message-nest/ {
    proxy_pass http://localhost:8000/message-nest/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## 注意事项

::: warning 注意
1. **修改配置后需要重启服务**才能生效
2. **前后端会自动同步**前缀配置，后端会在 HTML 中注入配置
3. **已有的 API 调用**会自动适配新的前缀
4. **静态资源**（CSS、JS、图片等）也会自动使用新的前缀
5. **开发模式**下，如果前端单独运行，需要确保后端 API 的前缀配置正确
:::

## 常见问题

### Q: 修改前缀后无法访问？

**A:** 请检查：
1. 是否已重启服务
2. 前缀格式是否正确
3. 如果使用反向代理，检查代理配置是否正确

### Q: 可以使用多级路径吗？

**A:** 可以，例如：`/api/message-nest` 或 `/services/messaging/nest`

### Q: URL 前缀会影响性能吗？

**A:** 不会，URL 前缀只是路径的一部分，不会影响应用性能。

## 示例

### 示例 1：部署在子路径

```ini
# conf/app.ini
[server]
UrlPrefix = /message-nest
```

访问地址：`http://your-domain.com/message-nest/`

### 示例 2：多级子路径

```ini
# conf/app.ini
[server]
UrlPrefix = /api/services/message-nest
```

访问地址：`http://your-domain.com/api/services/message-nest/`

### 示例 3：Docker 环境变量

```bash
docker run -d \
  -p 8000:8000 \
  -e URL_PREFIX=/message-nest \
  -e DB_TYPE=sqlite \
  message-nest:latest
```

访问地址：`http://localhost:8000/message-nest/`
