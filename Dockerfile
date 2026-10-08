# ==========================================
# Stage 1: Build Web Frontend (Full Local Build)
# ==========================================
FROM --platform=$BUILDPLATFORM node:20-alpine AS web-builder

WORKDIR /app/web

COPY web/package*.json ./
RUN --mount=type=cache,target=/root/.npm npm ci

COPY web/ ./
ENV NODE_ENV=prod
RUN npm run build

# ==========================================
# Stage 2: Build Go Backend (Full Local Build)
# ==========================================
FROM --platform=$BUILDPLATFORM golang:alpine AS go-builder

ENV GOTOOLCHAIN=auto
ARG TARGETOS
ARG TARGETARCH
ARG APP_VERSION=""

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata git

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# 如果传入了 APP_VERSION，则写入 .release_version
RUN if [ -n "$APP_VERSION" ]; then \
        echo "$APP_VERSION" > .release_version; \
    fi

# 拷贝前端构建产物用于 embed
COPY --from=web-builder /app/web/dist ./web/dist

# 编译单一二进制
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -ldflags="-s -w" -o Message-Nest .

# ==========================================
# Stage 3: Runtime Base (Alpine 极简底包)
# ==========================================
FROM alpine:3.21 AS runtime-base

ENV TZ=Asia/Shanghai

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

EXPOSE 8000

CMD ["/app/Message-Nest"]

# ==========================================
# Stage 4: CI Assembly (Target: ci)
# ==========================================
FROM runtime-base AS ci

COPY Message-Nest /app/Message-Nest
COPY conf /app/conf
COPY LICENSE NOTICE /app/

RUN chmod +x /app/Message-Nest

# ==========================================
# Stage 5: Full Release (Target: full, Default)
# ==========================================
FROM runtime-base AS full

COPY --from=go-builder /app/Message-Nest /app/Message-Nest
COPY --from=go-builder /app/conf /app/conf
COPY --from=go-builder /app/LICENSE /app/LICENSE
COPY --from=go-builder /app/NOTICE /app/NOTICE

RUN chmod +x /app/Message-Nest



