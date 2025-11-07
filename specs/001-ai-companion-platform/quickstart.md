# Quick Start Guide: AI陪伴精灵平台

**Version**: 1.0
**Date**: 2025-11-07
**Target Audience**: 开发人员

---

## Prerequisites

### Required Tools

- **Docker**: 24.0+ & Docker Compose 2.0+
- **Golang**: 1.21+
- **Python**: 3.11+
- **Node.js**: 20 LTS
- **PostgreSQL Client**: psql 15+
- **Git**: 2.40+

### Optional Tools

- **Kratos CLI**: `go install github.com/go-kratos/kratos/cmd/kratos/v2@latest`
- **protoc**: Protocol Buffers compiler
- **k6**: 压力测试工具
- **Postman**: API测试

---

## Quick Start (Docker Compose)

### 1. Clone Repository

```bash
git clone https://github.com/your-org/myy_chat.git
cd myy_chat
git checkout 001-ai-companion-platform
```

### 2. Environment Setup

```bash
# 复制环境变量模板
cp deployments/docker-compose/.env.example deployments/docker-compose/.env

# 编辑.env文件,填写必要的API密钥
nano deployments/docker-compose/.env
```

**必需配置**:
```env
# PostgreSQL
POSTGRES_USER=myy_chat
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=myy_chat

# Redis
REDIS_PASSWORD=your_redis_password

# LLM API Keys
OPENAI_API_KEY=sk-...
ANTHROPIC_API_KEY=sk-ant-...
AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
AZURE_OPENAI_KEY=...

# JWT Secret
JWT_SECRET=your_jwt_secret_key

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080

# Snowflake ID Configuration (分布式ID生成)
DATACENTER_ID=1        # 数据中心ID (0-31)
WORKER_ID=1            # 机器ID (0-31)
```

**Snowflake ID说明**:
- 每个服务实例需要唯一的 WORKER_ID
- 同一数据中心的服务使用相同的 DATACENTER_ID
- Docker Compose 单机部署: 手动为每个服务实例分配不同的 WORKER_ID
- Kubernetes 部署: 使用 StatefulSet 自动分配 (见下方Kubernetes配置)

### 3. Start Services

```bash
cd deployments/docker-compose

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 检查服务状态
docker-compose ps
```

**服务访问地址**:
- Frontend: http://localhost:3000
- API Gateway: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379
- Grafana: http://localhost:3001
- Prometheus: http://localhost:9090

### 4. Initialize Database

```bash
# 运行数据库迁移
docker-compose exec postgres psql -U myy_chat -d myy_chat -f /scripts/init-db.sql

# 加载种子数据(预设角色)
docker-compose exec postgres psql -U myy_chat -d myy_chat -f /scripts/seed-data.sql

# 验证表创建
docker-compose exec postgres psql -U myy_chat -d myy_chat -c "\dt"
```

### 5. Verify Setup

```bash
# 测试用户服务健康检查
curl http://localhost:8080/health

# 注册测试用户
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "Test1234!",
    "verification_code": "000000"
  }'

# 获取预设角色列表
curl http://localhost:8080/api/v1/characters?is_preset=true
```

---

## Local Development Setup

### Backend (Golang Services)

#### 1. Install Dependencies

```bash
# 进入服务目录
cd backend/golang/user-service

# 安装Go依赖
go mod download

# 安装Kratos CLI(如果未安装)
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest

# 生成Wire依赖注入代码
go generate ./...
```

#### 2. Generate Proto Code

```bash
# 安装protoc插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest

# 生成代码
cd api/user/v1
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       --go-http_out=. --go-http_opt=paths=source_relative \
       user.proto
```

#### 3. Run Service Locally

```bash
# 启动PostgreSQL和Redis(使用Docker)
docker-compose up -d postgres redis

# 配置环境变量
export DB_DSN="postgres://myy_chat:password@localhost:5432/myy_chat?sslmode=disable"
export REDIS_ADDR="localhost:6379"

# 运行服务
go run cmd/main.go

# 服务运行在:
# HTTP: http://localhost:8000
# gRPC: localhost:9000
```

#### 4. Run Tests

```bash
# 单元测试
go test ./internal/...

# 集成测试
go test -tags=integration ./tests/...

# 测试覆盖率
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

---

### Backend (Python Services)

#### 1. Setup Virtual Environment

```bash
cd backend/python/llm-service

# 创建虚拟环境
python3 -m venv venv
source venv/bin/activate  # Linux/Mac
# venv\Scripts\activate  # Windows

# 安装依赖
pip install -r requirements.txt

# 下载BGE-M3模型
python scripts/download_model.py
```

#### 2. Run Service

```bash
# 配置环境变量
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export REDIS_URL="redis://localhost:6379"

# 启动gRPC服务
python app/api/grpc_server.py

# 或启动FastAPI HTTP服务(调试用)
uvicorn app.api.http_api:app --reload --port 8001
```

#### 3. Run Tests

```bash
# 单元测试
pytest tests/unit/

# 集成测试
pytest tests/integration/

# 测试覆盖率
pytest --cov=app --cov-report=html tests/
```

---

### Frontend (Next.js)

#### 1. Install Dependencies

```bash
cd frontend

# 安装npm依赖
npm install

# 或使用pnpm(推荐)
pnpm install
```

#### 2. Run Development Server

```bash
# 启动开发服务器
npm run dev

# 访问: http://localhost:3000
```

#### 3. Build for Production

```bash
# 生产构建
npm run build

# 启动生产服务器
npm run start
```

#### 4. Run Tests

```bash
# 单元测试(如果配置了Jest)
npm run test

# E2E测试(Playwright)
npx playwright test

# 查看测试报告
npx playwright show-report
```

---

## Development Workflow

### 1. Create New Service (Golang)

```bash
# 使用Kratos CLI创建服务
kratos new backend/golang/example-service

cd backend/golang/example-service

# 项目结构
# ├── api/          # Protobuf定义
# ├── cmd/          # 启动入口
# ├── configs/      # 配置文件
# ├── internal/     # 业务逻辑
# │   ├── biz/      # 业务逻辑层
# │   ├── data/     # 数据访问层
# │   ├── service/  # gRPC服务实现
# │   └── server/   # 服务器配置
# └── tests/        # 测试
```

### 2. Database Migration

```bash
# 安装golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 创建迁移文件
migrate create -ext sql -dir migrations -seq add_new_table

# 执行迁移
migrate -path migrations \
        -database "postgres://localhost:5432/myy_chat?sslmode=disable" \
        up

# 回滚迁移
migrate -path migrations \
        -database "..." \
        down 1
```

### 3. Add New Proto API

```bash
# 1. 编辑Proto文件
nano api/user/v1/user.proto

# 2. 生成代码
make api

# 3. 实现服务端逻辑
# internal/service/user.go

# 4. 注册路由
# internal/server/http.go 或 grpc.go
```

### 4. Git Workflow

```bash
# 创建功能分支
git checkout -b feature/add-character-search

# 开发...
git add .
git commit -m "feat: add character search API"

# 推送并创建PR
git push origin feature/add-character-search
gh pr create --title "Add character search" --body "..."
```

---

## Snowflake ID Configuration

### Docker Compose Multi-Instance Setup

当需要水平扩展服务时,为每个实例配置不同的 WORKER_ID:

```yaml
# docker-compose.yml
services:
  user-service-1:
    image: myy_chat/user-service
    environment:
      - DATACENTER_ID=1
      - WORKER_ID=1
      - DB_DSN=postgres://...
    ports:
      - "8001:8000"

  user-service-2:
    image: myy_chat/user-service
    environment:
      - DATACENTER_ID=1
      - WORKER_ID=2  # 不同的 WORKER_ID
      - DB_DSN=postgres://...
    ports:
      - "8002:8000"

  user-service-3:
    image: myy_chat/user-service
    environment:
      - DATACENTER_ID=1
      - WORKER_ID=3  # 不同的 WORKER_ID
      - DB_DSN=postgres://...
    ports:
      - "8003:8000"
```

**启动多实例**:
```bash
docker-compose up -d user-service-1 user-service-2 user-service-3

# 验证每个实例的 WORKER_ID
docker-compose exec user-service-1 env | grep WORKER_ID
docker-compose exec user-service-2 env | grep WORKER_ID
docker-compose exec user-service-3 env | grep WORKER_ID
```

---

### Kubernetes StatefulSet Configuration

Kubernetes 部署使用 StatefulSet 自动分配机器ID:

```yaml
# k8s/user-service-statefulset.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: user-service
  namespace: myy-chat
spec:
  serviceName: user-service
  replicas: 3
  selector:
    matchLabels:
      app: user-service
  template:
    metadata:
      labels:
        app: user-service
    spec:
      containers:
      - name: user-service
        image: myy_chat/user-service:latest
        env:
        - name: DATACENTER_ID
          value: "1"

        # 从 Pod 名称提取序号作为 WORKER_ID
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name

        - name: WORKER_ID
          value: "$(echo $POD_NAME | rev | cut -d'-' -f1 | rev)"

        # 或使用 Init Container 设置
        # (见下方完整示例)

        - name: DB_DSN
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: dsn

        ports:
        - containerPort: 8000
          name: http
        - containerPort: 9000
          name: grpc

        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"

---
apiVersion: v1
kind: Service
metadata:
  name: user-service
  namespace: myy-chat
spec:
  clusterIP: None  # Headless Service for StatefulSet
  selector:
    app: user-service
  ports:
  - name: http
    port: 8000
  - name: grpc
    port: 9000
```

**使用 Init Container 提取 WORKER_ID** (推荐方式):

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: user-service
spec:
  replicas: 3
  template:
    spec:
      # Init Container 提取序号并写入共享卷
      initContainers:
      - name: extract-worker-id
        image: busybox
        command:
        - sh
        - -c
        - |
          # 从 Pod 名称提取序号: user-service-0 -> 0
          WORKER_ID=$(echo $POD_NAME | rev | cut -d'-' -f1 | rev)
          echo $WORKER_ID > /config/worker_id
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        volumeMounts:
        - name: config
          mountPath: /config

      containers:
      - name: user-service
        image: myy_chat/user-service:latest
        env:
        - name: DATACENTER_ID
          value: "1"
        - name: WORKER_ID_FILE
          value: "/config/worker_id"  # 应用从文件读取
        volumeMounts:
        - name: config
          mountPath: /config

      volumes:
      - name: config
        emptyDir: {}
```

**部署和验证**:

```bash
# 部署 StatefulSet
kubectl apply -f k8s/user-service-statefulset.yaml

# 查看 Pods
kubectl get pods -n myy-chat -l app=user-service
# 输出:
# NAME             READY   STATUS    RESTARTS   AGE
# user-service-0   1/1     Running   0          1m
# user-service-1   1/1     Running   0          1m
# user-service-2   1/1     Running   0          1m

# 验证每个 Pod 的 WORKER_ID
kubectl exec -n myy-chat user-service-0 -- env | grep WORKER_ID
kubectl exec -n myy-chat user-service-1 -- env | grep WORKER_ID
kubectl exec -n myy-chat user-service-2 -- env | grep WORKER_ID

# 测试 Snowflake ID 生成
kubectl exec -n myy-chat user-service-0 -- curl localhost:8000/health
```

**扩缩容**:

```bash
# 扩容到5个实例
kubectl scale statefulset user-service -n myy-chat --replicas=5

# 缩容到2个实例
kubectl scale statefulset user-service -n myy-chat --replicas=2
```

**注意事项**:

1. **WORKER_ID 范围**: 0-1023 (10位)
2. **DATACENTER_ID 范围**: 0-31 (5位) - 通常按环境/地域分配:
   - 0: 开发环境
   - 1: 测试环境
   - 2: 预发布环境
   - 3: 生产环境-北京
   - 4: 生产环境-上海
3. **StatefulSet 序号**: 从0开始递增 (user-service-0, user-service-1, ...)
4. **避免冲突**: 同一 DATACENTER_ID 下,每个服务实例的 WORKER_ID 必须唯一
5. **持久化**: 删除 Pod 后重新创建,序号保持不变

---

## Troubleshooting

### Common Issues

#### 1. PostgreSQL Connection Error

```bash
# 检查PostgreSQL是否运行
docker-compose ps postgres

# 查看日志
docker-compose logs postgres

# 测试连接
psql -h localhost -U myy_chat -d myy_chat
```

#### 2. Redis Connection Error

```bash
# 检查Redis
docker-compose ps redis

# 测试连接
redis-cli ping
# 应返回: PONG
```

#### 3. LLM API Error

```bash
# 检查API密钥配置
echo $OPENAI_API_KEY

# 测试OpenAI API
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

#### 4. Proto Generation Error

```bash
# 确保安装了protoc插件
which protoc-gen-go
which protoc-gen-go-grpc

# 重新安装
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

#### 5. Frontend Build Error

```bash
# 清理node_modules
rm -rf node_modules package-lock.json

# 重新安装
npm install

# 清理Next.js缓存
rm -rf .next
npm run build
```

---

## Useful Commands

### Docker Compose

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 重启特定服务
docker-compose restart user-service

# 查看日志
docker-compose logs -f user-service

# 进入容器
docker-compose exec postgres bash
docker-compose exec redis redis-cli
```

### Database

```bash
# 连接数据库
psql -h localhost -U myy_chat -d myy_chat

# 备份数据库
pg_dump -U myy_chat -d myy_chat > backup.sql

# 恢复数据库
psql -U myy_chat -d myy_chat < backup.sql

# 查看表结构
\dt        # 列出所有表
\d users   # 查看users表结构
```

### Monitoring

```bash
# 查看Prometheus指标
curl http://localhost:9090/metrics

# 查看服务健康状态
curl http://localhost:8080/health

# 压力测试
k6 run tests/load/conversation_test.js
```

---

## Next Steps

1. **阅读架构文档**: `docs/architecture.md`
2. **查看API文档**: `docs/api/README.md`
3. **运行测试**: `make test`
4. **开始开发**: 选择一个User Story开始实现
5. **执行任务**: `/speckit.tasks` 生成详细任务清单

---

## Support

- **文档**: `/docs`
- **Issues**: https://github.com/your-org/myy_chat/issues
- **Discussions**: https://github.com/your-org/myy_chat/discussions

---

**Document Version**: 1.0
**Last Updated**: 2025-11-07
