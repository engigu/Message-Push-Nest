# URL 路径前缀配置

Message-Nest 支持配置 URL 路径前缀，允许您在子路径下部署应用。

## 配置方式

### 方式 1: 配置文件

在 `conf/app.ini` 中添加：

```ini
[server]
UrlPrefix = /message
```

### 方式 2: 环境变量

设置环境变量：

```bash
export URL_PREFIX=/message
```

## 使用说明

1. **配置前缀**：在配置文件或环境变量中设置 `UrlPrefix`
2. **重启服务**：修改配置后需要重启 Message-Nest 服务
3. **访问应用**：使用新的 URL 访问，例如：`http://your-domain.com/message`

## 示例

### 默认访问（无前缀）
```
http://localhost:8000/
http://localhost:8000/api/v1/sendways/list
```

### 配置前缀后
```ini
[server]
UrlPrefix = /message
```

访问地址变为：
```
http://localhost:8000/message/
http://localhost:8000/message/api/v1/sendways/list
```

## 注意事项

- 前缀会自动添加 `/` 前缀（如果没有的话）
- 前端静态资源会自动使用相对路径，无需额外配置
- API 请求会自动添加路径前缀
- 修改配置后必须重启服务才能生效

## Nginx 反向代理示例

如果使用 Nginx 反向代理，配置示例：

```nginx
location /message/ {
    proxy_pass http://localhost:8000/message/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## 技术实现

- 后端使用 Gin 的路由组（RouterGroup）实现路径前缀
- 前端使用 Vite 的 `base: './'` 配置生成相对路径
- 后端在 HTML 中注入 `<base>` 标签和配置脚本
- 前端 API 请求自动添加路径前缀
