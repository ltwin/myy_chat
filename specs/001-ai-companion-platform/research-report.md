# AI 情感伴侣平台 - 技术研究报告

**生成时间**: 2026-02-01
**分支**: `001-ai-companion-platform`
**研究方法**: 多模型协作探索 (Codex 后端 + Gemini 前端)

---

## 1. 执行摘要

本报告基于对现有代码库的深度探索和 `specs/001-ai-companion-platform/spec.md` 需求分析，产出可执行的约束集合，用于指导后续规划和实施阶段。

### 用户决策确认

| 决策项 | 用户选择 |
|--------|----------|
| 后端语言策略 | **保持混合架构** (Go 核心业务 + Python AI/LLM) |
| Proto API 源 | **Go api/ 目录为准**，Python 生成客户端代码 |
| 数据库迁移 | **统一到 migrations/ 目录**，使用 golang-migrate |
| 前端 API 切换 | **环境变量切换** (VITE_API_MODE) |

---

## 2. 约束集合

### 2.1 硬约束 (Hard Constraints)

这些约束**不可违反**，违反将导致系统无法运行或严重架构问题。

| ID | 约束 | 来源 |
|----|------|------|
| HC-001 | Go 服务必须使用 Kratos v2 框架 | 现有代码锁定 |
| HC-002 | Go module 要求 `go 1.24.0`，构建镜像需匹配 | go.mod |
| HC-003 | 微服务间通信必须使用 gRPC + Protobuf | spec.md QR-002 |
| HC-004 | 对外 HTTP API 必须遵循 RESTful 标准 | spec.md QR-003 |
| HC-005 | Proto 定义以 `backend/golang/api/` 为权威源 | 用户决策 |
| HC-006 | 数据库主键使用 Snowflake BIGINT | 现有迁移脚本 |
| HC-007 | 必须支持 Docker Compose 和 Kubernetes 部署 | spec.md QR-004 |
| HC-008 | 测试覆盖率 ≥70% | spec.md QR-001 |
| HC-009 | 代码中文注释，日志英文 | spec.md QR-005 |
| HC-010 | AI 回复必须保持角色一致性 | spec.md QR-015 |

### 2.2 软约束 (Soft Constraints)

这些约束是**强烈建议**遵循的最佳实践。

| ID | 约束 | 来源 |
|----|------|------|
| SC-001 | Go 服务使用 Wire 依赖注入 | 现有约定 |
| SC-002 | 配置使用 YAML + Protobuf 格式 | Kratos 约定 |
| SC-003 | HTTP 端口 8000，gRPC 端口 9000 | 现有约定 |
| SC-004 | API 版本路径 `/api/v1/...` | 现有约定 |
| SC-005 | 前端使用 `cn()` 工具合并 Tailwind 类 | 现有约定 |
| SC-006 | 前端数据获取通过 service 层 | 现有约定 |
| SC-007 | 新类型定义添加到 `src/types/index.ts` | 现有约定 |

### 2.3 依赖关系

```
┌─────────────────────────────────────────────────────────────────┐
│                         APISIX Gateway                          │
│                    (JWT 认证 + 路由 + 限流)                      │
└─────────────────────────────────────────────────────────────────┘
                                │
        ┌───────────────────────┼───────────────────────┐
        ▼                       ▼                       ▼
┌───────────────┐     ┌───────────────┐     ┌───────────────┐
│  user-service │     │ conversation  │     │  character    │
│     (Go)      │     │   -service    │     │   -service    │
│               │     │     (Go)      │     │     (Go)      │
└───────┬───────┘     └───────┬───────┘     └───────────────┘
        │                     │
        │                     ▼
        │             ┌───────────────┐
        │             │  llm-agent    │
        │             │   -service    │
        │             │   (Python)    │
        │             └───────┬───────┘
        │                     │
        │                     ▼
        │             ┌───────────────┐
        │             │ LiteLLM Proxy │
        │             │  (多模型路由)  │
        │             └───────────────┘
        │
        ▼
┌───────────────────────────────────────────────────────────────┐
│                    PostgreSQL + Redis                          │
│              (pgvector 向量存储 + 缓存/会话)                    │
└───────────────────────────────────────────────────────────────┘
```

### 2.4 风险清单

| 优先级 | 风险 | 影响 | 缓解措施 |
|--------|------|------|----------|
| P0 | Go Dockerfile 使用 `golang:1.23-alpine` 但 go.mod 要求 `1.24.0` | 构建失败 | 更新 Dockerfile 基础镜像 |
| P0 | 健康检查路径不一致 (`/health` vs `/api/v1/health` vs `/healthz`) | 容器无法启动 | 统一为 `/api/v1/health` |
| P1 | Python gRPC 服务实现不完整 | 核心链路不可用 | 完成 proto 生成和 servicer 注册 |
| P1 | Go 跨服务客户端为占位实现 | 服务间调用失败 | 实现 gRPC 客户端 |
| P1 | 迁移脚本存在重复且不幂等 | 数据库初始化失败 | 整理并添加幂等性检查 |
| P2 | APISIX JWT 配置与 user-service 不对齐 | 认证失败 | 统一 JWT secret/issuer |
| P2 | 前端自定义 i18n 维护成本高 | 扩展困难 | 考虑迁移到 i18next |
| P3 | OpenAPI 产物为空 | 文档缺失 | 检查 proto 注解并重新生成 |

---

## 3. 技术栈确认

### 3.1 后端技术栈

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| **框架** | Kratos v2.8 (Go) | 微服务框架 |
| **AI 服务** | FastAPI + gRPC (Python) | LLM 集成 |
| **通信** | gRPC + Protobuf | 服务间通信 |
| **依赖注入** | Google Wire | Go 服务 |
| **数据库** | PostgreSQL 16 + pgvector | 主库 + 向量存储 |
| **缓存** | Redis 7 | 会话/缓存/限流 |
| **网关** | APISIX | API 网关 |
| **LLM 代理** | LiteLLM Proxy | 多模型路由 |
| **消息队列** | Kafka/NATS (计划) | 异步处理 |

### 3.2 前端技术栈

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| **框架** | React 19.2.0 | UI 框架 |
| **语言** | TypeScript 5.9.3 (strict) | 类型安全 |
| **构建** | Vite 7.2.4 | 构建工具 |
| **路由** | React Router DOM 7.11.0 | 客户端路由 |
| **状态** | React Context API | 全局状态 |
| **样式** | Tailwind CSS 3.4.17 | 原子化 CSS |
| **动画** | Framer Motion 12.23.26 | 动画库 |
| **图标** | Lucide React 0.562.0 | 图标库 |

### 3.3 部署技术栈

| 层级 | 技术选型 | 说明 |
|------|----------|------|
| **容器** | Docker | 容器化 |
| **编排** | Kubernetes + Helm | 云原生部署 |
| **CI/CD** | ArgoCD (计划) | GitOps |
| **可观测** | OpenTelemetry + Prometheus + Grafana | 监控 |

---

## 4. 成功标准提示

### 4.1 构建验证

- [ ] `backend/golang` 执行 `make setup` 后编译通过
- [ ] 所有 `*_grpc.pb.go`、`*_http.pb.go`、`*_pb.validate.go` 生成完整
- [ ] Python proto 代码从 Go api/ 目录生成

### 4.2 部署验证

- [ ] `docker-compose.dev.yml` 启动后所有容器 healthcheck 为 healthy
- [ ] APISIX upstream active check 通过
- [ ] 数据库迁移脚本可重复执行（幂等）

### 4.3 功能验证

- [ ] 通过 APISIX 访问 `/api/v1/...` 路由正常
- [ ] JWT 认证流程完整（登录 → 获取 token → 访问受保护资源）
- [ ] `conversation` 服务能调用 `llm-agent-service` 获得 AI 回复
- [ ] 积分扣费流程正常

### 4.4 前端验证

- [ ] `VITE_API_MODE=mock` 时使用 Mock API
- [ ] `VITE_API_MODE=real` 时连接真实后端
- [ ] 深色/浅色模式切换正常
- [ ] 中英文切换正常

---

## 5. 下一步行动

### 立即修复 (P0)

1. **更新 Go Dockerfile 基础镜像**
   ```dockerfile
   FROM golang:1.24-alpine AS builder
   ```

2. **统一健康检查路径**
   - Go 服务添加 `/api/v1/health` 端点
   - 更新 docker-compose 和 Helm 探针配置

### 短期任务 (P1)

3. **完成 Python gRPC 服务实现**
   - 运行 `generate_proto.sh` 生成代码
   - 完成 servicer 注册

4. **实现 Go 跨服务客户端**
   - `conversation` → `llm-agent-service` gRPC 客户端
   - `conversation` → `user-service` 积分客户端

5. **整理数据库迁移脚本**
   - 删除重复脚本
   - 添加幂等性检查 (`IF NOT EXISTS`)

### 中期任务 (P2)

6. **对齐 APISIX JWT 配置**
7. **实现前端 API 切换机制**
8. **生成 OpenAPI 文档**

---

## 6. 附录

### A. 服务清单

| 服务名 | 语言 | 端口 (HTTP/gRPC) | 状态 |
|--------|------|------------------|------|
| user-service | Go | 8000/9000 | 骨架完成 |
| character-service | Go | 8001/9001 | 骨架完成 |
| conversation-service | Go | 8002/9002 | 骨架完成 |
| memory-service | Go | 8003/9003 | 骨架 |
| billing-service | Go | 8004/9004 | 骨架 |
| admin-service | Go | 8005/9005 | 骨架 |
| analytics-service | Go | 8006/9006 | 骨架 |
| llm-agent-service | Python | 8080/50051 | 部分实现 |
| memory-processor | Python | 8081/50052 | 骨架 |

### B. 关键文件路径

```
backend/golang/
├── api/                    # Proto 定义 (权威源)
├── app/                    # 微服务实现
├── migrations/             # 数据库迁移 (统一入口)
├── pkg/                    # 共享库
└── Makefile                # 构建入口

backend/python/
├── llm-agent-service/      # LLM 服务
└── memory-processor/       # 记忆处理服务

frontend/
├── src/
│   ├── components/         # 可复用组件
│   ├── context/            # 状态管理
│   ├── pages/              # 页面组件
│   ├── services/           # API 服务层
│   └── types/              # 类型定义
└── vite.config.ts          # 构建配置

deployments/
├── docker-compose.dev.yml  # 开发环境
├── k8s/                    # Kubernetes 配置
└── apisix/                 # 网关配置
```

---

*本报告由 CCG 多模型协作工作流生成*
