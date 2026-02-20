# 变更提案: Argo CD + Helm（GitOps）交付骨架

## 需求背景

项目在云原生部署方面需要“可执行的交付骨架”，以便后续补齐各微服务 Helm Chart、HPA、探针与发布策略时有统一落点，并避免部署方式在文档层面停留于描述。

当前缺口主要集中在：

1. 未形成仓库内 GitOps 目录结构与示例入口（Argo CD Application/app-of-apps）
2. 缺少可复用的 Helm 服务模板（Deployment/Service/HPA/Config 注入）
3. 规格文档虽建议 K8s/Helm，但缺少与仓库实际文件路径的对应关系

## 变更内容

1. 增加 `deployments/k8s/` 云原生目录骨架（Helm + Argo CD）
2. 提供可复用 Helm chart（service chart），支持用不同 values 部署多个服务
3. 提供 Argo CD app-of-apps 示例（root app + env apps 目录）
4. 在部署策略文档中补充“GitOps 落点文件路径”说明

## 影响范围

- **模块:** 部署资源骨架、spec-kit 文档
- **文件:** 新增 `deployments/k8s/**`；更新 `specs/001-ai-companion-platform/deployment.md`
- **API:** 无
- **数据:** 无

## 核心场景

### 需求: 云原生交付骨架可执行
**模块:** 部署/交付
提供一个清晰入口，使团队可以：
- 在集群安装 Argo CD 后，一条 `kubectl apply` 导入 root app
- root app 自动同步并创建各服务 Application
- 每个服务 Application 使用 Helm chart + values 部署到指定 namespace

#### 场景: 新增一个微服务的云原生部署
只需要新增一份 values 文件和一个 Argo Application（或在 env kustomization 中引用），无需从零写 Deployment/Service/HPA。
- 预期结果：新增成本低、结构一致

#### 场景: 多环境（dev/staging/prod）差异管理
不同环境通过 values 文件区分镜像 tag、资源配额、HPA 等配置。
- 预期结果：配置可审计、可回滚

## 风险评估

- **风险:** “通用 chart”过度抽象导致难以理解
- **缓解:** 仅提供最小通用字段（镜像/端口/探针/资源/HPA/config），并在 README 中给出示例

