# 变更提案: 多环境部署策略补齐

## 需求背景

当前规格文档已包含“支持 Docker Compose 与 Kubernetes”的质量要求，但对以下问题缺少统一、可验收的说明：

1. “开发/调试运行方式”（例如 `go run`）与“正式部署方式”的边界不清晰
2. 多环境（开发/测试/生产）应如何选择启动形态、包含哪些组件、如何观测与排障没有矩阵化描述
3. 云原生交付路径未明确是否采用 GitOps（例如 Argo CD）以及验收口径

## 变更内容

1. 增加部署矩阵与验收点文档：Dev Run / Docker Compose / Kubernetes
2. 在 spec 中补充可验收的部署类质量要求（区分 Dev Run 与部署方式）
3. 在 quickstart 中补齐并修正本仓库实际路径（`deployments/docker-compose.dev.yml`）并明确 Dev Run 边界

## 影响范围

- **模块:** 规格文档（spec-kit 输出）
- **文件:** `specs/001-ai-companion-platform/spec.md`、`specs/001-ai-companion-platform/quickstart.md`、新增 `specs/001-ai-companion-platform/deployment.md`
- **API:** 无
- **数据:** 无

## 核心场景

### 需求: 多环境部署矩阵
**模块:** 文档/部署
项目需要对不同环境下的启动形态、组件构成、配置注入方式、日志与观测方式给出明确说明。

#### 场景: 开发者频繁调试单个服务
依赖服务常驻，业务服务可本机快速重启并查看日志。
- 预期：不把 `go run` 误认为生产部署方式

#### 场景: 单机部署用于验收/演示
一台机器能用 Compose 启动完整系统并可重复初始化数据。
- 预期：具备可执行入口与验收点

#### 场景: 云原生部署用于生产/扩展
使用 K8s 提供探针、滚动升级与扩缩容，并通过 GitOps 管理发布。
- 预期：明确 Argo CD 的适用性与建议组合

## 风险评估

- **风险:** 文档与仓库实际目录结构不一致导致误导
- **缓解:** 在 quickstart 中统一以 `deployments/docker-compose.dev.yml` 作为 Compose 入口，并集中引用部署矩阵文档

