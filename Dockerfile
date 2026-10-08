# ==========================================
# Stage 1: Build Web Frontend
# ==========================================
FROM node:20-alpine AS web-builder

WORKDIR /app/web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
ENV NODE_ENV=prod
RUN npm run build

# ==========================================
# Stage 2: Build Go Backend
# ==========================================
FROM golang:alpine AS go-builder

ENV GOTOOLCHAIN=auto
WORKDIR /app

ARG APP_VERSION=""

RUN apk add --no-cache ca-certificates tzdata git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 如果传入了 APP_VERSION 参数，则写入 .release_version 供系统信息展示
RUN if [ -n "$APP_VERSION" ]; then \
        echo "$APP_VERSION" > .release_version; \
    fi

# 将前端构建产物复制到 web/dist（供 //go:embed 使用）
COPY --from=web-builder /app/web/dist ./web/dist

# 编译生成单一静态二进制
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o Message-Nest .

# ==========================================
# Stage 3: Minimal Production Runtime
# ==========================================
FROM debian:bookworm-slim AS runner

ENV TZ=Asia/Shanghai

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=go-builder /app/Message-Nest /app/Message-Nest
COPY --from=go-builder /app/conf /app/conf
COPY --from=go-builder /app/LICENSE /app/LICENSE
COPY --from=go-builder /app/NOTICE /app/NOTICE

RUN chmod +x /app/Message-Nest

EXPOSE 8000

CMD ["/app/Message-Nest"]


