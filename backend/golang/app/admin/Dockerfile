# Admin Service Dockerfile (大仓模式)
# 构建: docker build -t myy-chat-admin -f app/admin/Dockerfile .
# 注意: 需要在 backend/golang/ 目录下执行构建

# ============================================
# 构建阶段
# ============================================
FROM golang:1.23-alpine AS builder

# 安装必要的构建工具
RUN apk add --no-cache make git

# 设置工作目录
WORKDIR /src

# 先复制依赖文件，利用Docker缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建 admin 服务
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=$(git describe --tags --always 2>/dev/null || echo 'v0.0.1')" \
    -o /app/admin \
    ./app/admin/cmd/admin

# ============================================
# 运行阶段
# ============================================
FROM alpine:3.19

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN adduser -D -g '' appuser

# 复制二进制文件
COPY --from=builder /app/admin /app/admin

# 创建配置目录
RUN mkdir -p /data/conf && chown -R appuser:appuser /app /data

# 切换到非root用户
USER appuser

WORKDIR /app

# 暴露端口
EXPOSE 8000 9000

# 挂载配置卷
VOLUME /data/conf

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# 启动服务
CMD ["./admin", "-conf", "/data/conf"]
