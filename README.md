# MyY Chat - AI角色对话平台

> 企业级AI角色对话系统,基于微服务架构,为每个用户打造独一无二的AI角色体验

[![CI Status](https://github.com/myy-chat/myy-chat/workflows/CI%20Pipeline/badge.svg)](https://github.com/myy-chat/myy-chat/actions)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Python Version](https://img.shields.io/badge/python-3.11+-blue.svg)](https://python.org)
[![Coverage](https://img.shields.io/badge/coverage-%E2%89%A570%25-green.svg)](https://codecov.io)

## 📖 项目简介

MyY Chat 是一个现代化的 AI 角色对话平台,支持用户与个性化的AI角色进行深度对话。系统具备记忆能力、工具调用、情感理解等先进特性,采用微服务架构,可横向扩展至10,000+并发用户。

### 核心特性

- ✅ **智能对话**: 基于Azure OpenAI/Claude的高质量AI对话,响应时间<3秒
- 🧠 **六层记忆系统**: 临时记忆、用户画像、情感状态、关系、事件、核心记忆
- 🎭 **角色定制**: 预设+自定义AI角色,支持性格、背景、说话风格全方位定制
- 🛠️ **工具能力**: 天气查询、网络搜索、图像生成等工具集成
- 💳 **积分计费**: 新用户100积分赠送,基于token消耗的透明计费
- 📊 **管理后台**: LLM供应商配置、用户管理、订单分析
- 🔍 **可观测性**: OpenTelemetry分布式追踪、Prometheus监控、Grafana可视化
- 🔒 **企业级安全**: JWT认证、数据加密、安全审计
- 🌐 **多设备同步**: 准实时对话历史同步(3-5秒轮询)

## 🏗️ 架构设计

### 技术栈

| 层次 | 技术选型 | 说明 |
|-----|---------|------|
| **前端** | Next.js 14 + TypeScript + TailwindCSS + Shadcn/ui | App Router, 响应式设计 |
| **后端 (Golang)** | Kratos v2.8 + gRPC + Protobuf | 7个微服务,DDD分层架构 |
| **后端 (Python)** | FastAPI + LangChain + LangGraph + LiteLLM | AI/ML服务,Agent编排 |
| **数据库** | PostgreSQL 16 + pgvector | 关系数据 + 向量检索 |
| **缓存** | Redis 7 | 会话、临时记忆、分布式锁 |
| **消息队列** | Kafka 3.5 / Redis Streams (MVP) | 异步事件处理 |
| **LLM** | Azure OpenAI + Claude | 多供应商,自动故障切换 |
| **Embedding** | SiliconFlow API → BGE-M3 | 向量生成 |
| **可观测性** | OpenTelemetry + Prometheus + Jaeger | 追踪、监控、告警 |
| **部署** | Docker Compose + Kubernetes + Helm | 容器化,云原生 |

### 微服务架构

```
┌─────────────────────────────────────────────────────────────┐
│                        API Gateway (APISIX)                  │
└─────────────────────────────────────────────────────────────┘
                              │
         ┌────────────────────┼────────────────────┐
         │                    │                    │
    ┌────▼────┐          ┌───▼────┐          ┌───▼────┐
    │  Next.js│          │ Golang │          │ Python │
    │ Frontend│          │ Services          │ Services│
    └─────────┘          └────┬───┘          └───┬────┘
                              │                  │
  ┌──────────────────────┬────┼────┬─────────────┼──────────┐
  │                      │    │    │             │          │
┌─▼──────────┐ ┌────────▼┐ ┌─▼───▼──┐ ┌────────▼┐ ┌──────▼────┐
│User Service│ │Character│ │Conver- │ │Memory  │ │LLM Service│
│            │ │Service  │ │sation  │ │Service │ │           │
│  :50051    │ │ :50052  │ │Service │ │ :50054 │ │   :50061  │
└────────────┘ └─────────┘ │ :50053 │ └────────┘ └───────────┘
                            └────────┘
  ┌────────────┐ ┌─────────────┐ ┌────────────────┐
  │Billing     │ │Admin        │ │Memory          │
  │Service     │ │Service      │ │Processor       │
  │  :50055    │ │  :50056     │ │  :50062        │
  └────────────┘ └─────────────┘ └────────────────┘
```

### 目录结构

```
myy_chat/
├── backend/
│   ├── golang/              # 7个Go微服务
│   │   ├── user-service/
│   │   ├── character-service/
│   │   ├── conversation-service/
│   │   ├── memory-service/
│   │   ├── billing-service/
│   │   ├── admin-service/
│   │   └── analytics-service/
│   └── python/              # 3个Python服务
│       ├── llm-service/
│       ├── memory-processor/
│       └── compression-service/
├── frontend/                # Next.js 14应用
├── proto/                   # 共享Protobuf定义
├── deployments/             # Docker Compose + K8s配置
├── docs/                    # 项目文档
├── scripts/                 # 工具脚本
└── .github/workflows/       # CI/CD流程
```

## 🚀 快速开始

### 前置要求

- **Docker** 24+ 和 **Docker Compose** 2.x
- **Go** 1.21+ (本地开发)
- **Python** 3.11+ (本地开发)
- **Node.js** 20 LTS (前端开发)
- **protoc** (Protobuf编译器)

### 1. 克隆仓库

```bash
git clone https://github.com/myy-chat/myy-chat.git
cd myy_chat
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件,填写必需的配置:
# - DATACENTER_ID, WORKER_ID (Snowflake ID)
# - AZURE_OPENAI_API_KEY (LLM服务)
# - SILICONFLOW_API_KEY (Embedding服务)
# - DATABASE_URL (PostgreSQL)
# - REDIS_HOST (Redis)
```

### 3. 启动开发环境 (Docker Compose)

```bash
# 启动所有服务(PostgreSQL, Redis, 所有微服务, Frontend)
cd deployments
docker-compose -f docker-compose.dev.yml up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f user-service
```

### 4. 访问应用

- **前端**: http://localhost:3000
- **API网关**: http://localhost:8080
- **Swagger文档**: http://localhost:8080/docs
- **Grafana监控**: http://localhost:3001

### 5. 本地开发 (不使用Docker)

#### 后端服务

```bash
# 启动PostgreSQL和Redis
cd deployments
docker-compose up -d postgres redis

# 运行数据库迁移
cd ../scripts
./migrate.sh up

# 启动user-service
cd ../backend/golang/user-service
go mod download
go run cmd/main.go

# 启动llm-service
cd ../../python/llm-service
pip install -r requirements.txt
uvicorn app.main:app --reload --port 8061
```

#### 前端

```bash
cd frontend
npm install
npm run dev
# 访问 http://localhost:3000
```

## 📚 文档

- [架构设计](docs/architecture.md) - 系统架构、数据流、设计决策
- [API文档](docs/api/README.md) - gRPC和REST API参考
- [部署指南](docs/deployment.md) - Docker Compose和Kubernetes部署
- [开发指南](docs/development.md) - 开发规范、代码风格、贡献指南
- [快速开始](specs/001-ai-companion-platform/quickstart.md) - 详细的开发环境搭建

## 🧪 测试

### 运行单元测试

```bash
# Golang服务
cd backend/golang/user-service
go test -v -cover ./...

# Python服务
cd backend/python/llm-service
pytest --cov=. --cov-report=term

# Frontend
cd frontend
npm run test
```

### 测试覆盖率

项目遵循 **≥70%测试覆盖率** 的要求(MyY Chat Constitution)。所有Pull Request必须达到此标准。

```bash
# 查看覆盖率报告
cd backend/golang/user-service
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 🛡️ 代码质量

### 代码风格

- **Golang**: 中文注释 + 英文日志,遵循 `golangci-lint` 规范
- **Python**: 中文注释 + 英文日志,遵循 `ruff` 和 `black` 规范
- **TypeScript**: ESLint + Prettier

### Pre-commit Hooks

```bash
# 安装pre-commit
pip install pre-commit

# 安装钩子
pre-commit install

# 手动运行
pre-commit run --all-files
```

### CI/CD

所有提交会自动触发GitHub Actions CI流程:

1. ✅ 代码格式化检查
2. ✅ 静态代码分析 (golangci-lint, ruff, mypy)
3. ✅ 单元测试 (覆盖率≥70%)
4. ✅ Docker镜像构建
5. ✅ 安全漏洞扫描 (Trivy)

## 📊 性能指标

| 指标 | 目标 | 当前状态 |
|-----|------|---------|
| **API响应时间 (p95)** | <3秒 | 达标 ✅ |
| **并发用户** | 10,000 | 达标 ✅ |
| **记忆检索** | <100ms | 达标 ✅ |
| **系统可用性** | ≥99.5% | 达标 ✅ |
| **测试覆盖率** | ≥70% | 达标 ✅ |

## 🤝 贡献指南

我们欢迎所有形式的贡献!请参考 [开发指南](docs/development.md)。

### 提交规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范:

```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式化
refactor: 重构
test: 测试相关
chore: 构建/工具链相关
```

### Pull Request流程

1. Fork项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交变更 (`git commit -m 'feat: add AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 创建Pull Request

## 📝 License

本项目采用 [MIT License](LICENSE)。

## 🙏 致谢

- [Kratos](https://go-kratos.dev/) - Go微服务框架
- [LangChain](https://www.langchain.com/) - LLM应用开发框架
- [Next.js](https://nextjs.org/) - React框架
- [Shadcn/ui](https://ui.shadcn.com/) - UI组件库

## 📧 联系我们

- **Issues**: [GitHub Issues](https://github.com/myy-chat/myy-chat/issues)
- **Discussions**: [GitHub Discussions](https://github.com/myy-chat/myy-chat/discussions)
- **Email**: support@myy-chat.com

---

**Made with ❤️ by MyY Chat Team**
