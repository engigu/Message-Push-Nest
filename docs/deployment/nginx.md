# Nginx 反向代理配置

本文档介绍如何使用 Nginx 作为 Message-Nest 的反向代理服务器。

## 基础配置（无 URL 前缀）

如果您希望直接通过域名访问 Message-Nest（如 `http://your-domain.com/`），使用以下配置：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 子路径配置（带 URL 前缀）

如果您希望在子路径下部署 Message-Nest（如 `http://your-domain.com/message/`），需要：

### 1. 配置 Message-Nest

在 `conf/app.ini` 中添加：

```ini
[server]
UrlPrefix = /message
```

或使用环境变量：

```bash
export URL_PREFIX=/message
```

### 2. 配置 Nginx

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location /message/ {
        proxy_pass http://localhost:8000/message/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

::: warning 注意
- Nginx 的 `location` 路径必须与 Message-Nest 的 `UrlPrefix` 配置一致
- `proxy_pass` 的 URL 末尾必须包含 `/`
:::

## HTTPS 配置

推荐使用 HTTPS 保护您的应用：

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate /path/to/your/certificate.crt;
    ssl_certificate_key /path/to/your/private.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://localhost:8000/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 完整配置示例

包含日志、缓存、超时等优化配置：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 访问日志
    access_log /var/log/nginx/message-nest-access.log;
    error_log /var/log/nginx/message-nest-error.log;

    # 客户端请求体大小限制
    client_max_body_size 10M;

    location / {
        proxy_pass http://localhost:8000/;
        
        # 代理头设置
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # 缓冲设置
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }

    # 静态资源缓存（可选）
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://localhost:8000;
        proxy_set_header Host $host;
        expires 7d;
        add_header Cache-Control "public, immutable";
    }
}
```

## 负载均衡配置

如果您有多个 Message-Nest 实例，可以配置负载均衡：

```nginx
upstream message_nest_backend {
    server localhost:8000;
    server localhost:8001;
    server localhost:8002;
}

server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://message_nest_backend/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 常见问题

### 1. 404 错误

确保：
- Message-Nest 的 `UrlPrefix` 配置与 Nginx 的 `location` 路径一致
- `proxy_pass` URL 末尾包含 `/`

### 2. 静态资源加载失败

检查：
- 浏览器开发者工具中的网络请求路径是否正确
- Nginx 配置中的 `proxy_pass` 是否正确

### 3. API 请求失败

确认：
- `proxy_set_header` 配置是否完整
- Message-Nest 服务是否正常运行

## 测试配置

修改配置后，测试并重载 Nginx：

```bash
# 测试配置文件语法
sudo nginx -t

# 重载配置
sudo nginx -s reload
```

## 相关文档

- [URL 路径前缀配置](/guide/url-prefix)
- [配置说明](/deployment/configuration)
