# 技术设计: 多环境部署策略补齐

## 技术方案

### 核心技术

- 文档分层：spec（约束/验收口径）+ quickstart（可执行入口）+ deployment matrix（策略与矩阵）
- 部署形态：Docker Compose（单机）+ Kubernetes（云原生）
- GitOps：Argo CD（推荐，用于部署与回滚审计）

### 实现要点

- 在 `spec.md` 里补充可验收的部署类质量要求：
  - 明确 Dev Run（开发/调试）不属于部署策略
  - Compose/K8s 必须有可执行路径与验收点
- 新增 `deployment.md` 做统一“部署矩阵”：
  - 场景/目标/启动方式/组件范围/日志观测/验收点
  - Snowflake Worker 分配策略在不同环境的规则
  - 云原生交付推荐 Argo CD + Helm（可选 Argo Rollouts）
- 修正 `quickstart.md` 中与仓库目录不一致的命令示例，统一以 `deployments/docker-compose.dev.yml` 作为 Compose 入口。

## 架构决策 ADR

### ADR-001: 区分 Dev Run 与正式部署方式
**上下文:** 项目需要同时支持开发调试、单机部署、云原生部署，但 `go run` 容易被误解为“部署方式”。  
**决策:** 明确 Dev Run 仅用于开发/调试，正式部署以容器为交付单元，采用 Compose/K8s。  
**理由:** 交付接口一致、环境差异更可控、便于测试与运维标准化。  
**替代方案:** 把 `go run` 作为“第三种部署方式” → 拒绝原因: 无法覆盖依赖与多语言服务的一致交付，验收口径不稳定。  
**影响:** 文档与验收口径更清晰；后续实现需补齐 Compose/K8s 的交付任务。

### ADR-002: 云原生交付采用 Argo CD（GitOps）
**上下文:** 多微服务、多环境需要可审计发布与回滚。  
**决策:** 推荐使用 Argo CD 管理 Helm/Kustomize，必要时引入 Argo Rollouts。  
**理由:** Git 即事实来源，发布历史可追溯，回滚/对比/权限体系成熟。  
**替代方案:** 手工 `kubectl apply` → 拒绝原因: 容易漂移、缺审计、回滚成本高。  
**影响:** 需要在 tasks 中补齐 Argo Application/values 管理策略（后续实现任务）。

## 测试与部署

- **测试:** 本变更为文档补齐，无代码测试；后续实现任务需为 Compose/K8s 交付增加集成验收。
- **部署:** 文档更新随仓库版本发布；云原生实现阶段再落地 Argo/Helm 资源。

