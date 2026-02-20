---
description: "Deployment strategy and environment matrix for AI角色对话平台"
---

# Deployment Strategy: 多环境/多形态部署

本文件补齐本项目在**开发调试**、**单机部署**、**云原生部署**三类场景下的部署策略说明，目标是让交付接口一致、路径清晰、可验收。

> 说明：`go run` **不是部署策略**，它是开发/调试时的本地运行方式（Dev Run）。正式部署形态以容器为基准（Docker Compose / Kubernetes）。

---

## 1. 目标与原则

### 1.1 目标

- 开发/测试阶段可以快速重启单个微服务、查看日志、复现问题。
- 单机部署可以在一台机器上以最少步骤启动完整系统（含依赖服务）。
- 云原生部署支持按服务扩缩容、滚动升级、健康探针、配置与密钥管理、可观测性与回滚。

### 1.2 原则

- **交付单元统一**：每个微服务最终都以容器镜像交付；本地 Dev Run 仅作为内循环加速。
- **依赖常驻**：数据库/缓存/LLM Gateway 等基础设施优先常驻容器，业务服务可本机运行或容器运行。
- **配置同构**：本地/Compose/K8s 的配置项保持一致，差异仅在注入方式（env、ConfigMap、Secret）。
- **可观测性优先**：日志输出到 stdout；K8s 通过探针与指标对齐运行状态；链路追踪保持跨语言一致。
- **迁移可控**：数据库迁移使用一次性 Job/脚本执行，避免服务启动时自动迁移导致不可控。

---

## 2. 部署矩阵（Environment Matrix）

| 场景 | 目标 | 启动方式 | 业务服务运行 | 依赖服务 | 日志查看 | 适用人群 |
|------|------|----------|--------------|----------|----------|----------|
| Dev Run（开发/调试） | 最快内循环，频繁重启单服务 | 本机命令（如 `go run`/Python reload） + 依赖用 Compose | 本机进程（推荐） | Compose 常驻 | 终端 stdout | 开发者 |
| Single-machine（单机部署） | 一台机器完整启动/验收 | Docker Compose | 容器 | Compose 内 | `docker compose logs -f` | 测试/演示/小规模试运行 |
| Cloud-native（生产/扩展） | 扩缩容、滚动升级、回滚、隔离 | Kubernetes + Helm（推荐 GitOps） | Pod（Deployment/StatefulSet） | K8s（或云托管） | `kubectl logs`（集中采集） | 生产/准生产 |

---

## 3. Dev Run（开发/调试运行方式）

### 3.1 推荐做法：依赖用 Compose，业务服务本机运行

- 依赖服务（PostgreSQL/Redis/LiteLLM/APISIX 等）使用 `deployments/docker-compose.dev.yml` 常驻启动。
- 需要频繁调试的微服务使用本机运行（Go: `go run` 或热重载；Python: `--reload` 或文件监听）。
- 日志直接输出到 stdout，配合 grep/rg 过滤。

### 3.2 验收点（开发体验）

- 能在 10 秒内重启单个微服务并看到健康就绪日志。
- 能在不重启依赖服务的情况下反复调试业务服务。

---

## 4. Docker Compose（单机部署）

### 4.1 目标

- 一条命令（或一条 make 目标）启动系统。
- 可重复执行初始化/迁移/种子数据导入。
- 支持查看/聚合日志与快速重启单服务。

### 4.2 验收点（单机部署）

- `docker compose` 启动后，核心入口（Gateway/Frontend）可访问。
- 关键依赖（PostgreSQL/Redis/LiteLLM）具备健康检查。
- DB 初始化/迁移作为独立步骤（脚本或一次性容器）可重复执行且幂等。

---

## 5. Kubernetes（云原生部署）

### 5.1 推荐交付形态

- **Helm Charts**：每个服务一个 chart，环境差异通过 values 管理。
- **探针统一**：每个服务提供 liveness/readiness（HTTP 或 gRPC health）。
- **扩缩容策略**：无状态服务优先 Deployment + HPA；有状态组件用 StatefulSet（或使用云托管服务）。
- **配置/密钥管理**：ConfigMap/Secret 注入；敏感信息不得入镜像与仓库明文。

### 5.2 Snowflake Worker 分配策略

- Compose：由环境变量手工为实例分配 `DATACENTER_ID/WORKER_ID`。
- K8s：使用 StatefulSet 的 pod 序号分配 `WORKER_ID`（例如 `pod-0 → 0`，`pod-1 → 1`），或通过下行 API 注入。

### 5.3 GitOps：Argo CD（推荐）

Argo CD + Kubernetes 是适合本项目的云原生交付组合：

- **适合原因**：多微服务、多环境（dev/staging/prod）、需要可审计回滚与一致性发布。
- **建议组合**：
  - CI：构建镜像、跑测试、推镜像仓库、更新部署仓库的 values/镜像 tag。
  - CD：Argo CD 同步 Helm/Kustomize 到集群，提供差异对比、回滚、审计记录。
- 可选：Argo Rollouts 实现金丝雀/蓝绿发布（上线风险更可控）。

#### 仓库落点（可执行骨架）

本仓库提供最小 GitOps 骨架与示例入口，便于后续补齐各服务 chart/values：

- 云原生目录入口：`deployments/k8s/README.md`
- 通用 Helm chart（可复用）：`deployments/k8s/helm/service/Chart.yaml`
- Argo CD app-of-apps 示例（dev）：`deployments/k8s/argocd/root-app-dev.yaml`
- dev 环境 Applications：`deployments/k8s/argocd/apps/dev/kustomization.yaml`

### 5.4 验收点（云原生）

- Helm 安装后，服务能通过探针进入 Ready。
- 支持滚动升级与回滚（Deployment revision 或 Argo 回滚）。
- 支持水平扩缩容（至少对无状态服务），日志可集中采集。
