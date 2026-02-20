# 技术设计: Argo CD + Helm（GitOps）交付骨架

## 技术方案

### 核心技术

- Kubernetes：云原生运行时
- Helm：服务部署模板化
- Argo CD：GitOps 持续交付（同步/差异/回滚/审计）
- Kustomize：用于 app-of-apps 组织多个 Argo Application（仅资源编排，不与 Helm 冲突）

### 实现要点

#### 1) 仓库目录约定

```text
deployments/k8s/
├── README.md
├── helm/
│   └── service/                 # 通用服务 Helm chart（可被多次复用）
│       ├── Chart.yaml
│       ├── values.yaml          # 默认值
│       ├── values/              # 示例 values（按环境/服务拆分）
│       │   ├── dev/
│       │   └── prod/
│       └── templates/
└── argocd/
    ├── README.md
    ├── root-app-dev.yaml        # app-of-apps 入口（dev 示例）
    └── apps/
        └── dev/
            ├── kustomization.yaml
            ├── user.yaml
            └── llm-agent-service.yaml
```

#### 2) Helm chart（service）能力边界

- 负责：Deployment/Service（可选 HPA、ConfigMap、ServiceAccount）
- 不负责：数据库等有状态依赖（应独立 chart 或托管）
- 目标：让“新增服务”主要通过 values 完成，而非复制粘贴 YAML

#### 3) Argo CD app-of-apps

- `root-app-dev.yaml` 指向 `deployments/k8s/argocd/apps/dev`
- `apps/dev/kustomization.yaml` 收集多个 Application
- 每个 Application 指向同一个 Helm chart，但使用不同 values 文件（在 chart 内部路径）

## 架构决策 ADR

### ADR-003: GitOps 采用 Argo CD app-of-apps
**上下文:** 多微服务、多环境需要统一入口、可扩展、可审计。  
**决策:** 用 Argo CD root app 管理 env apps 目录，env apps 以 Kustomize 组织多个 Application。  
**理由:** 最少样板代码，扩展新服务只需增加 values + Application 并引用。  
**替代方案:** 单个 Application 内写大量 Helm releases → 拒绝原因: 可维护性差，难以分权与独立回滚。  
**影响:** 需要约定 repoURL/targetRevision/namespace 等参数填写方式（README 说明）。

## 测试与部署

- **测试:** 本变更提供骨架与示例，不做集群联调；后续在引入真实 chart/镜像时补齐集群验收。
- **部署:** 安装 Argo CD 后，应用 `deployments/k8s/argocd/root-app-dev.yaml` 即可导入 dev 环境应用集合。

