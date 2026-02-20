FROM golang:1.24-alpine AS builder

# 安装必要工具
RUN apk add --no-cache make git gcc musl-dev

WORKDIR /src

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 构建服务
RUN mkdir -p /app && \
    go build -ldflags "-s -w" -o /app/character ./app/character/cmd/character

# 运行阶段
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/character .
COPY app/character/configs /data/conf

RUN chown -R appuser:appuser /app /data

USER appuser

EXPOSE 8000 9000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

CMD ["./character", "-conf", "/data/conf"]
