# Implementation Plan: AI陪伴精灵平台

**Branch**: `001-ai-companion-platform` | **Date**: 2025-11-07 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-ai-companion-platform/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

本项目旨在构建一个企业级的AI陪伴精灵平台,为每个用户打造独一无二的AI伴侣。核心特性包括:

- **用户认证与多设备同步**:支持邮箱/手机注册,准实时多设备同步(3-5秒轮询)
- **AI角色管理**:预设+自定义角色,支持性格、背景、说话风格等全方位定制
- **智能对话**:基于LLM的实时对话,响应时间<3秒,支持10,000并发用户
- **六层记忆系统**:临时记忆、用户画像、情感状态、关系、事件、核心记忆,支持自动整理和优先级管理
- **LLM Agent工具**:天气查询、网络搜索、图像生成等工具集成
- **积分计费系统**:新用户赠送100积分,基于token消耗的计费模式
- **后台管理与销售**:管理员配置、LLM供应商管理、客户订单分析
- **日志监控告警**:集中式日志、性能监控、异常流量检测

**技术方案**:
- 微服务架构(7个Golang服务 + 3个Python服务)
- Kratos框架(gRPC + Protobuf) + FastAPI + LangChain
- PostgreSQL 15 + pgvector(向量检索) + Redis 7
- 多LLM供应商架构(Azure OpenAI + Claude,自动故障切换)
- BGE-M3本地Embedding模型
- Next.js 14前端 + OpenTelemetry可观测性
- Docker Compose(开发) → Kubernetes(生产)

## Technical Context

**Language/Version**:
- Backend Golang: Go 1.21+
- Backend Python: Python 3.11+
- Frontend: TypeScript 5.x + Node.js 20 LTS

**Primary Dependencies**:
- **Golang**: Kratos v2.8+ (微服务框架), gRPC v1.70+, Protobuf v3, Wire (依赖注入), github.com/bwmarrin/snowflake (分布式ID生成)
- **Python**: FastAPI 0.121.0+, LangChain v1.0+, LangGraph v0.0.8+ LiteLLM 1.x, grpcio 1.70+, pysnowflake (分布式ID生成)
- **Frontend**: Next.js 14+, React 18+, TailwindCSS 3+, Shadcn/ui, Zustand 4+
- **AI/ML**: SiliconFlow API (MVP embedding) → HuggingFace Transformers BGE-M3 (生产环境可选), OpenAI SDK, Anthropic SDK
- **Observability**: OpenTelemetry, Prometheus, Jaeger

**Storage**:
- **Primary Database**: PostgreSQL 15+ with pgvector extension (关系数据 + 向量存储)
  - **ID生成**: Snowflake ID算法 (应用层生成,BIGINT主键,替代UUID)
  - **关系维护**: 无外键约束 (应用层维护,支持未来分库分表)
- **Cache/Session**: Redis 7+ (临时记忆、分布式锁、限流)
- **Message Queue**: Kafka 3.5+ (生产环境) / Redis Streams (MVP阶段)
- **Object Storage**: MinIO (自托管) / S3 (云端) - 用于头像、图片等

**Testing**:
- **Golang**: Go标准库 `testing` + `testify` 断言 + `gomock`
- **Python**: `pytest` + `pytest-asyncio` + `pytest-cov` (覆盖率)
- **Integration**: Postman/Newman (API测试)
- **E2E**: Playwright (前端)
- **Load Testing**: k6 (HTTP), ghz (gRPC)
- **Coverage Target**: ≥70% per microservice (宪法要求)

**Target Platform**:
- **Server**: Linux (Ubuntu 22.04 LTS / Alpine for containers)
- **Container**: Docker 24+, Docker Compose 2.x
- **Orchestration**: Kubernetes 1.28+ (生产环境)
- **Client**: Modern browsers (Chrome 90+, Safari 15+, Firefox 90+), 响应式设计

**Project Type**: Web Application (Microservices架构)

**Performance Goals**:
- **QPS**: 单服务器 100 QPS (对话请求)
- **Concurrency**: 支持 10,000 并发在线用户
- **Response Time**: AI回复 <3秒 (95th percentile)
- **Memory Query**: 记忆检索 <100ms
- **Throughput**: 向量检索 >400 QPS (pgvector基准)

**Constraints**:
- **Response Time**: 对话接口p95 <3秒, p99 <5秒
- **Memory Usage**: 单个Golang服务 <512MB, Python服务 <1GB
- **Database**: 记忆查询 <100ms, 事务性操作 <50ms
- **Availability**: 系统可用性 ≥99.5% (月度)
- **Data Retention**: 对话历史永久保留, 临时记忆24小时TTL
- **Compliance**: 符合《个人信息保护法》,账号删除30天冷静期

**Scale/Scope**:
- **Users**: MVP目标 1,000-10,000 用户
- **Services**: 10个微服务 (7 Golang + 3 Python)
- **Data Volume**:
  - 用户数: 10,000
  - 每用户对话: 1,000条/年
  - 记忆向量: ~1,000,000条
  - 存储估算: PostgreSQL ~50GB/年, Redis ~10GB
- **Code Size**: 预估 50,000-100,000 LOC
- **API Endpoints**: ~100个 REST/gRPC 接口

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### MyY Chat Constitution Gates

**代码质量原则 Gates:**
- [x] Testing strategy ensures ≥70% test coverage per microservice
  - Go: `testing` + `testify` + `gomock`
  - Python: `pytest` + `pytest-cov`
  - Integration: Newman, Playwright
  - Coverage tracking: `go test -cover`, `pytest --cov`
- [x] Security risks, concurrency, and idempotency addressed in design
  - 幂等性: Redis分布式锁 + `idempotency_key`字段
  - 并发安全: PostgreSQL事务隔离, Redis原子操作
  - 安全风险: API鉴权(JWT), 敏感数据加密, SQL注入防护
- [x] Static code analysis tools configured
  - Go: `golangci-lint`, `go vet`
  - Python: `ruff`, `mypy`, `bandit`(安全扫描)
  - 集成到CI/CD pipeline

**API规范原则 Gates:**
- [x] Service communication uses gRPC + Protobuf (Kratos framework)
  - Kratos v2.7+原生支持gRPC
  - Proto文件统一管理在`api/`目录
  - Python服务使用`grpcio`实现gRPC服务端
- [x] External APIs follow RESTful standards
  - HTTP接口遵循REST规范(GET/POST/PUT/DELETE)
  - 标准HTTP状态码(200, 201, 400, 401, 404, 500)
  - 统一响应格式: `{"code": 0, "message": "success", "data": {...}}`
- [x] API documentation plan included
  - gRPC: Protobuf注释自动生成文档
  - REST: OpenAPI 3.0规范, Swagger UI
  - 文档部署到`/docs`路径

**部署原则 Gates:**
- [x] Docker Compose configuration for development/single-machine
  - `docker-compose.yml`包含所有服务
  - 本地开发一键启动: `docker-compose up -d`
  - 环境变量配置: `.env`文件
- [x] Kubernetes manifests for cloud-native deployment
  - Helm Charts组织K8s配置
  - Deployments, Services, ConfigMaps, Secrets
  - HPA(水平扩容): CPU/Memory阈值触发
- [x] Configuration switching mechanism designed
  - 配置文件分层: `config/dev.yaml`, `config/prod.yaml`
  - 环境变量覆盖: `CONFIG_ENV=prod`
  - Kubernetes ConfigMap动态挂载

**设计原则 Gates:**
- [x] Object-oriented design principles followed (LSP, DIP, etc.)
  - 依赖倒置(DIP): 接口层`biz/`依赖抽象,数据层`data/`实现接口
  - 里氏替换(LSP): LLM供应商可互相替换(统一接口)
  - 接口隔离(ISP): 服务接口职责单一
- [x] Design patterns justified, not over-engineered
  - **Factory**: LLM客户端工厂(根据配置创建OpenAI/Claude客户端)
  - **Strategy**: 记忆检索策略(向量检索/关键词匹配)
  - **Circuit Breaker**: LLM调用熔断
  - 避免过度设计,仅在必要时使用模式
- [x] Single responsibility principle applied
  - 每个微服务单一职责(用户服务只管理用户,对话服务只管理会话)
  - 每个函数/方法职责明确
  - 代码审查强制SRP

**项目结构原则 Gates:**
- [x] backend/golang and backend/python separation maintained
  ```
  backend/
  ├── golang/
  │   ├── user-service/
  │   ├── character-service/
  │   ├── conversation-service/
  │   ├── memory-service/
  │   └── billing-service/
  └── python/
      ├── llm-service/
      ├── memory-processor/
      └── compression-service/
  ```
- [x] frontend directory properly structured
  ```
  frontend/
  ├── src/
  │   ├── app/           # Next.js 14 App Router
  │   ├── components/    # 可复用组件
  │   ├── lib/          # 工具函数
  │   └── hooks/        # 自定义Hooks
  ├── public/           # 静态资源
  └── tests/            # E2E测试
  ```
- [x] Each service has independent config, tests, deployment files
  - 每个服务包含: `Dockerfile`, `config/`, `tests/`, `README.md`
  - 独立部署单元,可单独构建和发布

**架构原则 Gates:**
- [x] High readability, scalability, maintainability demonstrated
  - **可读性**: 遵循Go/Python官方编码规范,中文注释解释业务逻辑
  - **可扩展性**: 微服务架构,服务可独立扩容,HPA自动扩展
  - **可维护性**: DDD分层架构,清晰的服务边界,完善的文档
- [x] Horizontal expansion capability designed
  - 无状态服务设计(会话存Redis,数据存DB)
  - Kubernetes HPA根据CPU/Memory自动扩展Pod数量
  - 数据库读写分离(主从复制),后期可引入Citus分片
- [x] Service coupling minimized
  - 服务间仅通过gRPC通信,无直接数据库访问
  - 事件驱动架构(Kafka),异步解耦
  - API网关统一入口,服务不直接暴露

**开发规范原则 Gates:**
- [x] Chinese comments planned for code
  ```go
  // 用户服务 - 负责用户认证、授权和画像管理
  type UserService struct {
      repo UserRepository
  }

  // CreateUser 创建新用户,包含邮箱验证和密码哈希
  func (s *UserService) CreateUser(ctx context.Context, req *CreateUserRequest) error {
      // 验证邮箱格式
      if !isValidEmail(req.Email) {
          return errors.New("邮箱格式不正确")
      }
      // ...
  }
  ```
- [x] English logging configured
  ```go
  log.Info("user created successfully",
      "user_id", user.ID,
      "email", user.Email,
      "timestamp", time.Now(),
  )
  ```
- [x] Documentation plan aligned with requirements
  - **API文档**: Protobuf注释 + Swagger
  - **架构文档**: `docs/architecture.md`
  - **部署文档**: `docs/deployment.md`
  - **开发指南**: `docs/development.md`

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
myy_chat/
├── backend/
│   ├── golang/
│   │   ├── user-service/           # 用户认证与管理服务
│   │   │   ├── api/                # Protobuf API定义
│   │   │   │   └── user/v1/
│   │   │   │       ├── user.proto
│   │   │   │       └── user_http.proto
│   │   │   ├── cmd/                # 启动入口
│   │   │   │   └── main.go
│   │   │   ├── internal/           # 业务逻辑(私有)
│   │   │   │   ├── biz/            # 业务逻辑层
│   │   │   │   ├── data/           # 数据访问层
│   │   │   │   ├── service/        # gRPC服务实现
│   │   │   │   └── server/         # 服务器配置
│   │   │   ├── configs/            # 配置文件
│   │   │   ├── tests/              # 单元测试+集成测试
│   │   │   ├── Dockerfile
│   │   │   ├── go.mod
│   │   │   └── README.md
│   │   │
│   │   ├── character-service/      # 角色管理服务
│   │   ├── conversation-service/   # 对话编排服务
│   │   ├── memory-service/         # 记忆存储与检索服务
│   │   ├── billing-service/        # 积分计费服务
│   │   ├── admin-service/          # 后台管理服务
│   │   └── analytics-service/      # 统计分析服务
│   │
│   └── python/
│       ├── llm-service/            # LLM调用与Agent服务
│       │   ├── app/
│       │   │   ├── api/
│       │   │   │   ├── grpc_server.py    # gRPC服务端
│       │   │   │   └── http_api.py       # FastAPI(可选)
│       │   │   ├── core/
│       │   │   │   ├── llm_client.py     # LLM客户端封装
│       │   │   │   ├── circuit_breaker.py
│       │   │   │   └── config.py
│       │   │   ├── services/
│       │   │   │   ├── chat.py           # 对话编排
│       │   │   │   ├── agent.py          # LangGraph Agent
│       │   │   │   └── tools.py          # 工具函数
│       │   │   └── prompts/              # Prompt模板
│       │   ├── proto/                    # Proto文件
│       │   ├── tests/
│       │   ├── Dockerfile
│       │   ├── requirements.txt
│       │   └── README.md
│       │
│       ├── memory-processor/       # 记忆提取与Embedding生成
│       │   ├── app/
│       │   │   ├── services/
│       │   │   │   ├── embedding.py      # BGE-M3 Embedding
│       │   │   │   └── extraction.py     # 记忆提取
│       │   │   └── models/
│       │   └── tests/
│       │
│       └── compression-service/    # 上下文压缩服务
│           └── app/
│               └── services/
│                   └── compression.py
│
├── frontend/
│   ├── src/
│   │   ├── app/                    # Next.js 14 App Router
│   │   │   ├── (auth)/             # 认证路由组
│   │   │   │   ├── login/
│   │   │   │   └── register/
│   │   │   ├── (main)/             # 主应用路由组
│   │   │   │   ├── chat/
│   │   │   │   ├── characters/
│   │   │   │   ├── credits/
│   │   │   │   └── profile/
│   │   │   ├── admin/              # 后台管理
│   │   │   ├── layout.tsx
│   │   │   └── page.tsx
│   │   ├── components/
│   │   │   ├── ui/                 # Shadcn/ui组件
│   │   │   ├── chat/               # 对话相关组件
│   │   │   ├── character/          # 角色相关组件
│   │   │   └── common/             # 通用组件
│   │   ├── lib/
│   │   │   ├── api/                # API客户端
│   │   │   ├── hooks/              # 自定义Hooks
│   │   │   └── utils/              # 工具函数
│   │   └── styles/                 # 全局样式
│   ├── public/                     # 静态资源
│   ├── tests/                      # E2E测试
│   ├── Dockerfile
│   ├── package.json
│   └── README.md
│
├── api-gateway/                    # API网关(APISIX配置)
│   ├── routes/
│   └── plugins/
│
├── deployments/
│   ├── docker-compose/
│   │   ├── docker-compose.yml      # 开发环境
│   │   ├── docker-compose.prod.yml # 单机生产环境
│   │   └── .env.example
│   └── kubernetes/
│       ├── helm/                   # Helm Charts
│       │   ├── user-service/
│       │   ├── llm-service/
│       │   └── frontend/
│       └── base/                   # 基础设施
│           ├── postgres/
│           ├── redis/
│           └── kafka/
│
├── proto/                          # 共享Proto文件
│   ├── user/v1/
│   ├── character/v1/
│   ├── conversation/v1/
│   └── memory/v1/
│
├── docs/
│   ├── architecture.md             # 架构设计
│   ├── api/                        # API文档
│   ├── deployment.md               # 部署指南
│   └── development.md              # 开发指南
│
├── scripts/
│   ├── init-db.sql                 # 数据库初始化
│   ├── seed-data.sql               # 种子数据
│   └── migrate.sh                  # 数据库迁移
│
└── .github/
    └── workflows/
        ├── ci.yml                  # CI流程
        └── cd.yml                  # CD流程
```

**Structure Decision**:

本项目采用**微服务Web应用架构**,严格遵循项目宪法的目录结构要求:

1. **backend/golang**: 包含7个Go微服务,每个服务遵循Kratos DDD分层架构(api/biz/data/service)
2. **backend/python**: 包含3个Python服务,用于AI/ML相关功能(LLM调用、Embedding生成、上下文压缩)
3. **frontend**: Next.js 14应用,使用App Router架构
4. **proto/**: 共享的Protobuf定义,供所有服务使用
5. **deployments/**: 部署配置,支持Docker Compose和Kubernetes两种模式
6. **docs/**: 集中式文档管理

每个微服务都是独立的部署单元,包含独立的配置、测试和Dockerfile。

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

无违反宪法的设计决策。所有架构选择均符合项目宪法原则。

**复杂度合理性说明**:
- **10个微服务**: 基于DDD bounded context划分,每个服务职责单一明确
- **双语言栈**(Golang+Python): Golang处理高并发业务逻辑,Python处理AI/ML计算,各展所长
- **多数据库**(PostgreSQL+Redis+Kafka): 分别用于持久化、缓存、消息队列,职责清晰
- **双部署模式**(Docker Compose+K8s): 渐进式复杂度,适应不同阶段需求

所有复杂度引入均有明确的业务或技术理由,不存在过度设计。
