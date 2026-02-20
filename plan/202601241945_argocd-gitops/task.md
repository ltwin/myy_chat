# 任务清单: Argo CD + Helm（GitOps）交付骨架

目录: `plan/202601241945_argocd-gitops/`

---

## 1. 云原生目录骨架
- [ ] 1.1 新增 `deployments/k8s/README.md`，说明目录结构与使用方式，验证 why.md#核心场景-需求-云原生交付骨架可执行
- [ ] 1.2 新增 `deployments/k8s/helm/service` 通用 Helm chart（Deployment/Service/HPA/Config），验证 why.md#核心场景-场景-新增一个微服务的云原生部署
- [ ] 1.3 新增 `deployments/k8s/argocd` 示例（root app + env apps），验证 why.md#核心场景-需求-云原生交付骨架可执行

## 2. 文档对齐
- [ ] 2.1 在 `specs/001-ai-companion-platform/deployment.md` 中增加 GitOps 文件路径落点与引用，验证 why.md#核心场景-场景-多环境devstagingprod差异管理

