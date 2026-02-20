# Kubernetes Deployment (Helm + Argo CD)

本目录提供云原生部署的“交付骨架”，用于后续逐步补齐各微服务的镜像、探针、HPA 与发布策略。

## 目录结构

```text
deployments/k8s/
├── helm/
│   └── service/          # 通用服务 Helm chart（可被多个微服务复用）
└── argocd/
    ├── root-app-dev.yaml # app-of-apps 入口（dev 示例）
    └── apps/
        └── dev/          # dev 环境的 Argo Applications
```

## 使用方式（推荐 GitOps）

1. 在集群安装 Argo CD（安装方式与集群差异较大，这里不在仓库内固化）。
2. 修改 `deployments/k8s/argocd/root-app-dev.yaml` 中的 `spec.source.repoURL`、`targetRevision`、`destination` 等字段。
3. 执行：`kubectl apply -f deployments/k8s/argocd/root-app-dev.yaml`
4. Argo CD 将自动同步 `deployments/k8s/argocd/apps/dev/`，创建并部署各服务 Application。

## 说明

- Helm chart `helm/service` 是一个通用模板，通过不同 values 文件部署不同微服务。
- 示例 values 放在 `deployments/k8s/helm/service/values/` 下，按环境/服务拆分。

