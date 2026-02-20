# Change Proposal: AI 情感伴侣平台

**Change ID**: `ai-companion-platform`
**Created**: 2026-02-01
**Status**: Planning
**Method**: CCG Multi-Model Collaboration (Codex + Gemini)

---

## Why

当前仓库已具备部分 AI 对话与计费能力，但在 OpenSpec 维度仍缺少可归档的规格增量（delta specs）与统一的跨服务契约，导致方案无法通过规范校验，也无法稳定指导后续实现与验收。

本次变更需要把“预扣后结算计费 + 消息幂等 + 流式对话 + 多端实时同步 + 前端体验升级”收敛为可验证的规范，优先解决上线阻塞的架构一致性问题（健康检查、迁移脚本、契约漂移）。

## What Changes

- 增加 OpenSpec delta specs，覆盖 6 个核心能力：基线治理、账本计费、幂等流式对话、WebSocket 多端同步、长期记忆核心、前端对话体验。
- 明确积分账本边界：`billing-service` 作为唯一 SoT，余额变更必须通过 gRPC。
- 统一消息幂等语义：客户端生成 `client_message_id`（字符串）并全链路透传。
- 固化迁移脚本基线：明确保留/删除清单，消除 `001/002` 重复编号导致的执行歧义，确保新成员可零决策执行。
- 补齐 `memory-service` 交付闭环：增加服务骨架确认任务与 `memory/v1` Proto 扩展任务（不再停留在 Hello 示例）。
- 明确 conversation→memory 的调用关系为“结算后异步事件驱动”，避免记忆链路故障阻塞主对话响应。
- 将“长期记忆系统”提升为当前变更核心，补齐五层记忆架构（压缩、情感、事件、图谱、画像）的规范与任务拆解。
- 补充用户画像字段分级与冲突处理规则：Tier A（单值澄清门禁）、Tier B（多值共存排序）、Tier C（推断弱信号）。
- 明确 `user_profiles`（账号资料）与 `user_portraits`（对话画像）双 SoT 边界，定义来源优先级与非双向自动覆盖策略。
- 新增 `mcp-service` 作为 Agent 通用 MCP 接入层/网关，采用按域插件化模型；当前变更范围先落地 `profile.*` 工具域并强制走 memory-service SoT 边界。
- 补齐充值链路 MVP（套餐查询、下单、支付回调入账、幂等防重）与配置种子数据要求。
- 明确旧 `backend/python/llm-agent-service/` 的冻结、迁移与移除门槛，避免 Go/Python 双实现并行歧义。
- 补齐部署基线：LiteLLM Proxy 独立部署与 APISIX CORS 白名单配置纳入验收。
- 将原 Deferred 的三项能力前置到当前版本：记忆管理基础编辑、情感状态基础版、安全审计最小闭环（无后台 UI）。
- **BREAKING**：计费内部契约从“直接扣减/增加”演进为“预扣/结算/释放”工作流（兼容策略见 design 与 specs）。

## Capabilities

### New Capabilities
- `platform-baseline-hardening`: 统一运行时基线、健康检查路径与幂等迁移规则。
- `billing-reserve-ledger`: 计费账本 SoT、预扣-结算-释放流程与幂等保证。
- `conversation-idempotent-streaming`: 发送消息链路中的幂等、流式回复与失败补偿。
- `realtime-websocket-sync`: 鉴权连接、心跳重连与多设备消息同步。
- `long-term-memory-core`: 五层长期记忆核心系统（写入、检索、压缩、遗忘、图谱同步、可见性隔离）。
- `frontend-chat-experience`: 前端 API 对接、流式加载体验、余额与记忆展示。

### Modified Capabilities
无（当前仓库尚未建立 `openspec/specs/` 主规格，本次均按新增能力建模）。

## Impact

- 后端：`backend/api/`, `backend/app/`, `backend/migrations/`（含 `llm-agent-service` 的 Go 实现）
- 前端：`frontend/src/`（Query、WebSocket、流式渲染、余额与记忆 UI）
- 网关与部署：`deployments/apisix/`, `deployments/docker-compose.dev.yml`, `deployments/k8s/`

## Context

基于 `specs/001-ai-companion-platform/spec.md` 需求规格，实现企业级 AI 角色对话平台。

### 用户需求
- 用户认证与多设备同步
- AI 角色创建、自定义与管理
- 智能对话与记忆系统
- 积分计费与在线充值
- 后台管理与运维监控

### 技术约束
- 后端：Go (Kratos v2)
- 前端：React 19 + TypeScript + Vite
- 数据库：PostgreSQL + pgvector + Redis
- 部署：Docker Compose + Kubernetes

---

## User Decisions (Confirmed)

| 决策项 | 用户选择 | 约束参数 |
|--------|----------|----------|
| 后端语言策略 | 单一 Go 后端 | 核心业务与 LLM 编排均使用 Go (Kratos) |
| Proto API 源 | Go api/ 目录为准 | Go 服务内统一生成与使用 |
| 数据库迁移 | 统一到 migrations/ | golang-migrate 工具 |
| 前端 API 切换 | 环境变量切换 | VITE_API_MODE=mock/real |
| 响应 SLA | 流式输出 + TTFT 3s | SSE/WebSocket 传输 |
| 扣费策略 | 预扣后结算 | 多退少补，不允许透支 |
| 幂等粒度 | 按消息级别 | client_message_id 作为幂等键（字符串） |
| 消息同步 | WebSocket/SSE | 替代短轮询 |

---

## Hard Constraints

| ID | 约束 | 参数 |
|----|------|------|
| HC-001 | Go 服务框架 | Kratos v2.8 |
| HC-002 | Go 版本 | 1.24.0 |
| HC-003 | 服务间通信 | gRPC + Protobuf |
| HC-004 | 对外 API | RESTful via APISIX |
| HC-005 | 数据库主键 | Snowflake BIGINT |
| HC-006 | 密码哈希 | bcrypt cost=12 |
| HC-007 | JWT 配置 | TTL=15min, Refresh=7d |
| HC-008 | 测试覆盖率 | ≥70% |
| HC-009 | 健康检查路径 | /api/v1/health |
| HC-010 | 登录锁定 | 5次失败后锁定15分钟 |
| HC-011 | Token 传输策略 | access token 走 Authorization Header；refresh token 走 HttpOnly Secure SameSite Cookie |
| HC-012 | 登出失效语义 | 基于 Redis 黑名单实现“登出立即失效”（不等待 access token 自然过期） |
| HC-013 | 黑名单故障策略 | Redis 黑名单检查默认 fail-closed，受控降级需有会话回查兜底 |
| HC-014 | Cookie 安全 | refresh/logout/logout-all 必须通过 CSRF 防护验收 |

---

## Non-Goals (Out of Scope)

- 语音对话功能
- 视频通话或虚拟形象
- AI 角色主动推送消息
- 多用户群聊
- 移动端原生 APP
- 微信小程序接入

## Deferred Features (Explicit Backlog)

### Pulled Into Current Release (v1.0)

以下需求原位于 Deferred，但因“长期记忆为核心卖点”已纳入当前版本实现：

| 功能 | 原始 User Story | 纳入原因 | 当前范围 |
|------|-----------------|----------|----------|
| 记忆管理基础编辑能力（MemoryList + MemoryEdit） | US2, FR-024, FR-025 | 用户可见与可控是记忆卖点闭环的必要条件 | 列表、基础编辑/删除、审计留痕 |
| 情感状态基础版（Layer 2 基线） | US6, FR-018 | 记忆连续性需要情感上下文支撑 | PAD（valence/arousal/dominance）+ prompt 注入，不含高级演化策略 |
| 安全审计最小闭环（无 UI） | US11, FR-060 | 长期记忆上线必须可追溯风险事件 | 越权检索/冲突澄清/同步失败事件记录与查询 |

### Remaining Deferred Features

以下功能来源于 `specs/001-ai-companion-platform/spec.md`，本轮 MVP 不实现，明确列入后续版本 backlog：

| 功能 | 原始 User Story | 延期原因 | 目标版本 |
|------|-----------------|----------|----------|
| 跨角色全局记忆融合与高级图算法推荐 | US2, US10 | 需在单角色长期记忆稳定后再引入跨角色知识融合 | v1.2 |
| 工具调用能力（天气/搜索/图片生成） | US4 | 需额外 API 集成，非核心对话链路 | v1.1 |
| 情感状态高级演化（长期状态机 + 因果解释） | US6 | 当前仅交付 PAD 基线，演化策略依赖长期线上反馈 | v1.2 |
| 后台管理系统 UI | US7 | 可通过配置文件/数据库直接操作替代 | v1.1 |
| 销售 CRM 系统 | US8 | 商业化运营，产品验证后再建 | v2.0 |
| 日志监控与安全告警 UI | US9 | 基础 OTel + Prometheus 先行 | v1.1 |
| 高级记忆整理策略（个性化遗忘曲线参数自适应） | US10 | 需基于线上长期数据做策略学习与A/B验证 | v1.2 |
| AI 角色公共社区与多人交互 | US11 | 需记忆分区、安全锁等复杂机制 | v2.0 |
| 记忆高级评分参数（`memory_strength/boost_history/surprise_score/connectivity`） | US10 | 涉及策略训练与线上回放评估，当前先保留简化评分参数 | v1.2 |
| 事件高级结构（`title/is_recurring/participants/related_memory_ids/source_message_ids[]`） | US10 | 属于记忆可解释性增强，需在基础事件链路稳定后引入 | v1.2 |
| 图谱增强字段（`relation_description/mention_count/last_mentioned_at`） | US10 | 当前先保证时态边与去重消歧，热度统计后置 | v1.2 |
| 密码重置流程（邮件/token 校验） | US1 | 当前优先注册/登录/锁定主链路，重置链路延后 | v1.1 |


### Deferred Data-Model Clarification (v1.0 Freeze)

为避免“字段缺失”与“有意延期”混淆，本变更明确：

- 已纳入 v1.0：claim 状态机、时态边、硬过滤门禁、Proto 对齐关键字段（`actor_user_id/access_count/source_type`）。
- 延期到 v1.2：高级记忆评分参数、事件高级结构、图谱热度统计与高级解释字段。
- 上述延期项不阻塞当前版本验收，不计入本轮交付 Gate。

## Reference Knowledge Base

本变更以 openspec 为唯一执行计划。`specs/001-ai-companion-platform/` 下的 spec-kit 文档降级为参考知识库，定向吸收以下内容：

- **data-model.md**: 数据模型设计（本轮提炼 MVP 子集合入 design.md）
- **spec.md**: 需求细节（FR-xxx 编号）、边缘用例、验收场景
- **tasks.md**: 前端组件任务定义（ChatLoadingState 等）
- **contracts/**: Proto 合约字段约定

---

## Success Criteria

- [ ] 所有 P0 风险修复完成
- [ ] 核心服务（user/character/conversation/billing）可用
- [ ] LLM 调用链路完整（conversation → llm-agent → LiteLLM）
- [ ] 旧 Python llm-agent 路径完成冻结与部署剔除，Go 实现成为唯一运行入口
- [ ] 积分预扣结算流程正常
- [ ] 充值链路 MVP 可用（套餐查询→下单→支付回调入账，幂等防重）
- [ ] 五层长期记忆核心链路可用（写入→检索→压缩→优先级调整）
- [ ] 记忆可见性隔离与冲突澄清流程通过
- [ ] 画像字段分级与冲突治理通过验收（单值澄清、多值共存、职业时态）
- [ ] `user_profiles` 与 `user_portraits` SoT 边界在实现与验收中保持一致
- [ ] Agent 侧画像 CRUD 通过 `mcp-service` 接入且 SoT 边界有效
- [ ] `mcp-service` 插件注册中心与 `deny-by-default` 策略生效
- [ ] APISIX CORS 白名单与 LiteLLM Proxy 独立部署验收通过
- [ ] WebSocket 消息推送正常
- [ ] 测试覆盖率 ≥70%
- [ ] Docker Compose 部署验收通过
