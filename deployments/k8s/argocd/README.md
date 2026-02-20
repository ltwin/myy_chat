# Argo CD (GitOps) 示例

本目录提供 **app-of-apps** 模式的最小示例，用于把“环境 → 多服务”组织成一个可导入的根应用。

## 使用步骤（dev 示例）

1. 编辑 `deployments/k8s/argocd/root-app-dev.yaml`：
   - `spec.source.repoURL`：改为你的仓库地址
   - `spec.source.targetRevision`：改为分支或 tag
   - `spec.destination.server`：通常是 `https://kubernetes.default.svc`
   - `spec.destination.namespace`：通常是 `argocd`
2. 执行：`kubectl apply -f deployments/k8s/argocd/root-app-dev.yaml`
3. Argo CD 会同步 `deployments/k8s/argocd/apps/dev/`，并创建其中定义的各服务 Application。

## 约定

- `apps/<env>/` 目录只放 Argo Application（以及 kustomization 组织文件）。
- 每个服务 Application 都指向同一份通用 Helm chart：`deployments/k8s/helm/service`。
- 不同服务/环境通过 `helm.valueFiles` 使用不同 values 文件（本仓库示例放在 chart 内的 `values/<env>/` 下）。

