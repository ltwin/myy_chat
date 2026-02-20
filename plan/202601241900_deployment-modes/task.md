# 任务清单: 多环境部署策略补齐

目录: `plan/202601241900_deployment-modes/`

---

## 1. 规格文档补齐
- [ ] 1.1 在 `specs/001-ai-companion-platform/spec.md` 中补充 `QR-016`（区分 Dev Run 与部署方式），并引用部署矩阵文档，验证 why.md#核心场景-需求-多环境部署矩阵-场景-开发者频繁调试单个服务
- [ ] 1.2 新增 `specs/001-ai-companion-platform/deployment.md`，提供部署矩阵、验收点与 Argo CD 建议，验证 why.md#核心场景-需求-多环境部署矩阵-场景-云原生部署用于生产扩展
- [ ] 1.3 更新 `specs/001-ai-companion-platform/quickstart.md`，修正 Compose 命令入口与 Dev Run 边界说明，验证 why.md#核心场景-需求-多环境部署矩阵-场景-单机部署用于验收演示

## 2. 后续实现任务（不在本变更内）
- [ ] 2.1 增加 Argo CD 应用声明模板（Helm values 管理与镜像 tag 策略）
- [ ] 2.2 补齐各服务 Dockerfile、健康探针与 migration Job（Compose/K8s 验收闭环）

