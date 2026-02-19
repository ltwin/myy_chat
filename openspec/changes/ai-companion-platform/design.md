# Design: AI 情感伴侣平台

**Change ID**: `ai-companion-platform`
**Created**: 2026-02-01

---

## Architecture Overview

```
┌────────────────────────────────────────────────────────────────────────────┐
│                              APISIX Gateway                               │
│                         (JWT 认证 + 路由 + 限流)                           │
└────────────────────────────────────────────────────────────────────────────┘
                                      │
        ┌───────────────┬─────────────┼─────────────┬───────────────┬───────────────┐
        ▼               ▼             ▼             ▼               ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│  user-service │ │ conversation  │ │  character    │ │ billing-      │ │ memory-       │
│     (Go)      │ │   -service    │ │   -service    │ │ service (Go)  │ │ service (Go)  │
└───────────────┘ │     (Go)      │ └───────────────┘ └───────────────┘ └───────────────┘
                  └───────┬───────┘          │               ▲               ▲
                          │                  │               │               │
                          │                  │      Reserve/Settle/Release   │
                          │                  │               │               │
                          ▼                  │               │               │
                  ┌───────────────┐          │               │     consume `evt.memory.write.requested`
                  │  llm-agent    │          │               │               │
                  │   -service    │          │               │               │
                  │     (Go)      │          │               │               │
                  └───────┬───┬───┘          │               │               │
                          │   │              │               │               │
                          │   └─ tool call ───────────────────────────────────────┐
                          │                                                       ▼
                          │                                            ┌──────────────────────┐
                          │                                            │     mcp-service      │
                          │                                            │  (MCP Gateway, Go)   │
                          │                                            └─────────┬────────────┘
                          │                                                      │ route tools to
                          ▼                                                      │ memory/user/character/
                  ┌───────────────┐                                              │ conversation/billing
                  │ LiteLLM Proxy │                                              │
                  │  (多模型路由)  │                                              │
                  └───────────────┘                                              │
                                                                                  │
                                   ┌────────────────────────────────────────────┐
                                   │                PostgreSQL + Redis          │
                                   │    (pgvector 向量存储 + 缓存 + Stream)     │
                                   └────────────────────────────────────────────┘
```

Agent 工具调用主链路：`llm-agent -> mcp-service -> 目标微服务 API`。`mcp-service` 不直连业务数据库。

---

## Technical Decisions

### 1. 后端架构

| 决策 | 选择 | 理由 | 被拒绝的替代方案 |
|------|------|------|------------------|
| 后端语言 | Go（单语言） | 降低部署与运维复杂度，减少跨语言契约与排障成本 | 多语言混合后端 |
| 服务框架 | Kratos v2 | 成熟的微服务框架，内置 gRPC/HTTP | Go-Micro, Go Kit |
| 通信协议 | gRPC + Protobuf | 强类型，高性能，多语言支持 | HTTP/JSON |
| Proto 源 | Go api/ 目录 | 单一权威源，避免漂移 | 各服务独立维护 |
| MCP 接入层 | `mcp-service`（通用网关） | Agent 工具统一鉴权、路由、审计，降低 Agent 与业务服务耦合 | 每个工具域独立 MCP 服务 |

### 2. 数据存储

| 决策 | 选择 | 理由 | 被拒绝的替代方案 |
|------|------|------|------------------|
| 主数据库 | PostgreSQL 16 | 成熟稳定，支持 pgvector | MySQL, MongoDB |
| 向量存储 | pgvector | 无需独立服务，简化运维 | Milvus, Pinecone |
| 缓存 | Redis 7 | 高性能，支持分布式锁 | Memcached |
| 主键策略 | Snowflake BIGINT | 有序，支持分布式 | UUID, 自增 |

### 3. 前端架构

| 决策 | 选择 | 理由 | 被拒绝的替代方案 |
|------|------|------|------------------|
| 框架 | React 19 | 最新特性，生态成熟 | Vue, Next.js |
| 构建 | Vite 7 | 快速 HMR，现代化 | Webpack, CRA |
| 状态管理 | Context + TanStack Query | 简单场景 Context，服务端状态 Query | Redux, Zustand |
| 样式 | Tailwind CSS | 原子化，快速开发 | CSS Modules, Styled |

### 4. 实时通信

| 决策 | 选择 | 理由 | 被拒绝的替代方案 |
|------|------|------|------------------|
| 消息同步 | WebSocket/SSE | 实时推送，减少服务器压力 | 短轮询 |
| LLM 响应 | SSE 流式 | TTFT 3s SLA，用户体验好 | 完整响应 |

补充：当 WebSocket 在弱网或代理环境不可用时，客户端自动降级到 3-5 秒轮询，确保多端同步延迟口径仍满足 ≤5 秒（95%）。

### 4.1 SLA 口径映射

- `FR-012` 的“3秒内返回回复”在流式模式下映射为 **TTFT ≤ 3s**（正常网络）。
- 完整响应耗时由前端分级反馈兜底：10s 提示“AI 正在努力思考”，30s 提供超时与重试。
- 验收必须同时覆盖：TTFT 目标、10s/30s 分级提示、超时补偿路径。

### 5. 积分系统

| 决策 | 选择 | 理由 | 被拒绝的替代方案 |
|------|------|------|------------------|
| 扣费策略 | 预扣后结算 | 避免透支，多退少补 | 后扣费，允许透支 |
| 幂等粒度 | 消息级别 | 重试不重复扣费 | LLM 调用级别 |
| 账本 SoT | billing-service | 集中治理，便于审计 | 分散在各服务 |

### 6. 长期记忆系统

| 决策 | 选择 | 理由 | 被拒绝的替代方案 |
|------|------|------|------------------|
| 记忆架构 | 五层记忆模型（Layer1-5） | 平衡召回质量、可解释性与可维护性 | 单层向量记忆 |
| 图谱存储策略 | Postgres 图投影（v1.0）+ 外部图库适配层预留（v1.2） | v1.0 先消除新增基础设施复杂度，后续按需扩展图检索增强 | 首版即强依赖 Neo4j |
| 冲突更新策略 | 暂停覆盖 + 人设澄清 | 降低误记风险，符合角色体验 | 新值直接覆盖旧值 |
| 画像采集策略 | AI 自动抽取优先 + 关键字段轻确认 | 降低用户填写成本并控制误记风险 | 强制手填 / 全自动覆盖 |
| 字段冲突策略 | 字段分级（单值澄清、多值共存、时态历史） | 保证画像准确性与长期连续性 | 全量覆盖或全量并存 |
| 优先级演化 | 时间衰减 + 强化唤醒 | 模拟长期使用的自然遗忘与强化 | 固定优先级 |
| 可见性边界 | Owner-Private / Public 强制过滤 | 防止跨身份记忆泄露 | 仅在前端过滤 |
| 用户资料 SoT 边界 | `user_profiles`（账号资料）+ `user_portraits`（对话画像） | 解耦“用户设置”和“对话记忆推断”，避免跨服务互相覆盖 | 所有画像字段单表承载 |

### 7. 旧 Python llm-agent 退役策略

- 现有 `backend/python/llm-agent-service/` 进入冻结模式（不再新增功能，仅保留迁移窗口回滚参考）。
- 生产与预发统一切换到 `backend/app/llm-agent/`（Go）作为唯一运行实现。
- CI 与部署基线移除 Python llm-agent 入口，避免双实现并存造成发布歧义。
- 删除门槛：Go 版本完成等价验收（流式回复、成本回传、故障回退）并稳定一个发布周期后，移除 Python 目录与相关脚本引用。

---

## Portrait Field Tiering Policy

### 画像字段分级

| 层级 | 字段类型 | 典型字段 | 写入策略 | 冲突策略 |
|------|----------|----------|----------|----------|
| Tier A | 关键单值字段 | `full_name`, `birth_date`, `current_occupation` | 先写 claim，不直接覆盖主画像 | 必须澄清后再更新 canonical |
| Tier B | 多值事实字段 | `interests`, `skills`, `role_tags` | 允许增量写入并共存 | 不强制替换，按权重与近期活跃度排序 |
| Tier C | 推断弱信号 | 性格倾向、偏好推断 | 低权重写入，默认可衰减 | 不影响 Tier A canonical 决策 |

### 职业冲突处理规则

- `current_occupation` 作为单值投影字段，仅接受已决议值。
- 历史职业保存在 `career_history`（或等价时态结构）中，保留时间上下文。
- 当用户出现“程序员 → 产品经理”变更时，更新当前职业并保留历史，而不是删除旧事实。

### 输入来源策略

- 默认来源：AI 从对话自动识别并写入 claim。
- 用户手动编辑：作为高可信来源参与决议。
- 关键字段冲突时采用低打扰确认（卡片或人设化澄清），避免强制表单化录入。

---

## User Profile SoT Boundary

### 目标

明确 `user-service.user_profiles` 与 `memory-service.user_portraits` 的职责边界，避免数据重叠导致实施歧义。

### 职责分层

| 模型 | 所属服务 | 主要用途 | SoT 范围 |
|------|----------|----------|----------|
| `user_profiles` | user-service | 账号资料与用户设置（用户可直接编辑） | 账户侧资料展示与设置流程 |
| `user_portraits` | memory-service | 对话记忆画像（Agent 检索/推理） | 画像 canonical 与记忆检索注入 |

### 同步与冲突规则

1. 不做双向自动覆盖同步，避免循环写入与竞态放大。
2. 用户在 `user_profiles` 的手动编辑，以 `USER_STATED` 来源写入 claim 流（高优先级），按 Tier 规则进入决议。
3. `memory-service` 决议后的 canonical 更新仅写 `user_portraits`；`user_profiles` 可按需读取投影展示，但不被自动反写覆盖用户手填字段。
4. 来源优先级：`USER_STATED > EXTRACTED > INFERRED`；同级来源按时间与置信度决议。

---

## MCP Service Architecture

### 定位

`mcp-service` 是 Agent 的通用 MCP 接入层/网关，不是新的业务数据 SoT。  
业务数据 SoT 仍保留在各业务微服务中（例如画像与记忆由 `memory-service` 持久化）。

### 工具域（MCP）

当前变更已纳入画像工具域（Profile Tools）：
- `profile.get`
- `profile.claim_upsert`
- `profile.conflict_list`
- `profile.conflict_resolve`
- `profile.fact_delete`
- `profile.audit_query`

并预留扩展到其他工具域（如 user/character/conversation/billing 查询与操作类工具）；本次验收范围仅要求 `profile.*` 全链路可用。

### 工具域路由矩阵

| 工具域 | MCP tool 前缀 | 下游服务 | 写入策略 | 当前状态 |
|------|----------------|----------|----------|----------|
| Profile | `profile.*` | memory-service | 写入走 claim 状态机与字段分级策略 | 本期实现 |
| User | `user.*` | user-service | 默认只读；敏感写操作需显式白名单 | 预留 |
| Character | `character.*` | character-service | 角色资产写操作需 owner 校验 | 预留 |
| Conversation | `conversation.*` | conversation-service | 会话写操作需会话归属校验 | 预留 |
| Billing | `billing.*` | billing-service | 默认只读；余额变更仍只允许 billing gRPC 内部工作流 | 预留 |

路由原则：`tool_name -> 域路由 -> 下游服务 client -> 统一审计`，不允许 MCP 进程内直连业务表。

### 权限边界矩阵

| 调用主体 | 允许工具范围 | 强制校验 | 拒绝策略 |
|----------|--------------|----------|----------|
| Agent（用户会话） | 当前会话授权的 `profile.*`（及后续白名单工具） | `user_id`/`character_id` 归属、scope、幂等键 | 越权直接拒绝并写审计事件 |
| Agent（系统任务） | 平台配置白名单工具（最小权限） | 任务来源签名、租户边界、速率限制 | 非白名单工具默认拒绝 |
| 内部服务调用 | 显式注册的服务到服务工具 | mTLS/服务身份、方法级 ACL | 未注册调用方拒绝 |

默认策略为 `deny-by-default`：未注册工具、未声明 scope、跨用户资源访问一律拒绝。

### 按域插件化模型（Domain Plugin）

`mcp-service` 采用“核心网关 + 域插件”结构：

- **Core（稳定层）**：鉴权、ACL、参数校验、幂等透传、限流、审计、错误规范化。
- **Plugin（变化层）**：按域实现工具集合（`profile.*`、`user.*`、`conversation.*` 等）。
- **Registry（编排层）**：统一维护 `tool_name -> domain -> handler -> required_scopes -> downstream_service` 映射。

插件契约（逻辑约束）：

- 每个插件必须声明 `domain` 与 `tools` 列表。
- 每个工具必须声明：`input schema`、`required scopes`、`target service`、`retry policy`。
- 工具注册时执行冲突检查：禁止重复 `tool_name`、禁止未声明 scope 的写工具上线。

生命周期：

1. 启动阶段：插件注册到 Registry，执行静态校验。
2. 运行阶段：Core 中间件链统一处理后，按 Registry 路由到域 Handler。
3. 下线阶段：先从 Registry 摘除工具，再停止插件执行，避免新请求进入。

### Agent 调用协议（多域工具）

推荐统一两类 MCP 方法：

- `tools.list`: 返回当前调用主体可见工具（已过 ACL 过滤）。
- `tools.call`: 执行指定工具，由 mcp-service 注入服务端上下文（`user_id`、`character_id`、`tenant_id`、`request_id`）。

执行语义：

1. Agent 先 `tools.list` 获取可调用工具与参数 schema。
2. Agent 发起 `tools.call(tool_name, args)`。
3. mcp-service 完成鉴权/ACL/校验后路由到对应域插件。
4. 域插件调用下游微服务 API，返回标准化响应。

关键约束：

- `user_id`/`tenant_id` 等身份字段以服务端注入为准，不信任 Agent 透传值。
- 未注册工具、跨域越权、缺失 scope 请求均按 `deny-by-default` 拒绝并审计。

### 错误与重试语义（网关统一）

| 错误码 | 含义 | 是否可重试 | 处理策略 |
|--------|------|------------|----------|
| `UNAUTHORIZED` | 调用方身份无效 | 否 | 直接拒绝并审计 |
| `FORBIDDEN_SCOPE` | scope 不满足 | 否 | 直接拒绝并审计 |
| `TOOL_NOT_FOUND` | 工具未注册/不可见 | 否 | 返回工具列表刷新建议 |
| `VALIDATION_FAILED` | 入参不符合 schema | 否 | 返回字段级错误 |
| `UPSTREAM_TIMEOUT` | 下游服务超时 | 是 | 按工具策略指数退避重试 |
| `UPSTREAM_UNAVAILABLE` | 下游不可用 | 是 | 熔断+降级（只读优先） |
| `CONFLICT_PENDING` | 业务冲突待澄清（如画像 claim） | 条件可重试 | 引导进入澄清流程 |

### 边界与调用链

```text
Agent
  -> MCP tools (mcp-service)
      -> target microservice APIs (memory/user/character/conversation/billing)
          -> PostgreSQL/Redis
```

- Agent 不允许直接访问数据库或绕过业务服务边界。
- MCP 层负责鉴权、参数校验、路由分发、幂等键透传（`client_message_id`）、审计补充。
- 画像写操作必须通过 `mcp-service -> memory-service`，并经过 claim 状态机与字段分级策略。
- 当前范围内，`profile.*` 工具审计统一落到 `memory_audit_events`；跨域工具审计模型（`mcp_audit_events`）在后续版本扩展。

## PBT Properties

### 积分系统

| 属性 | 不变量 | 证伪策略 |
|------|--------|----------|
| 余额非负 | `balance >= 0` | 并发扣费+重试风暴，断言余额≥0 |
| 扣费幂等 | 同一 client_message_id 只产生一条流水 | 重复请求验证流水唯一性 |
| 预扣结算一致 | 预扣金额 ≥ 实际扣费 | 随机 cost_usd 验证多退少补 |

### 消息系统

| 属性 | 不变量 | 证伪策略 |
|------|--------|----------|
| 消息单调 | 同一游标分页不丢不重 | 并发写入+分页查询交错验证 |
| 多端一致 | 最终消息列表完全一致 | 多端并发发送后比对 |
| 响应时限 | TTFT ≤ 3s 或返回超时错误 | 注入延迟验证超时处理 |

### 长期记忆系统

| 属性 | 不变量 | 证伪策略 |
|------|--------|----------|
| 记忆隔离 | 非 Owner 不可读 Owner-Private 记忆 | 构造跨身份检索请求验证结果过滤 |
| 冲突澄清 | 矛盾信息不直接覆盖 | 注入互斥画像信息，验证进入澄清流程 |
| 单值字段安全更新 | Tier A 字段未经决议不得覆盖 canonical | 注入矛盾单值信息并验证仍需澄清 |
| 多值字段共存 | Tier B 字段冲突不触发 destructive overwrite | 注入多兴趣/多技能事实并验证共存 |
| 压缩触发 | 上下文达阈值时触发压缩 | 压测长对话并验证 summary 产出 |
| 图谱同步幂等 | 同一节点/边重复同步不产生重复关系 | 重复同步任务验证版本与唯一约束 |

### 安全系统

| 属性 | 不变量 | 证伪策略 |
|------|--------|----------|
| 删除原子 | 账号删除后无半删状态 | 删除作业中断+重启验证 |
| MCP SoT 边界 | Agent 画像写入不得绕过 memory-service | 模拟直写路径并断言被拒绝/审计 |

> 注：记忆隔离属性统一定义在“长期记忆系统”小节，避免重复验收口径。

---

## Risk Mitigation

| 优先级 | 风险 | 缓解措施 |
|--------|------|----------|
| P0 | Go Dockerfile 版本不匹配 | 更新为 golang:1.24-alpine |
| P0 | 健康检查路径不一致 | 统一为 /api/v1/health |
| P0 | 积分并发扣费 | 幂等键 + 原子条件更新 |
| P0 | 记忆隔离失败 | 数据层 + Prompt 层双重过滤 |
| P1 | mcp-service 成为工具调用瓶颈 | 插件化分域 + 域级限流/熔断 + 热点域可远程化拆分 |
| P1 | LLM 故障切换 | Redis 集中状态 + 熔断器 |
| P1 | 账号删除半删 | 事务 + 可重入删除作业 |
| P1 | LiteLLM cost 缺失 | 兜底估算 + 审计日志 |

---

## Data Model

完整数据库表设计与 ER 图已前置落地到 `openspec/changes/ai-companion-platform/data-model.md`（作为开发前基线）。

本节保留跨模块核心摘要，详细字段、DDL、索引与迁移映射以 `data-model.md` 为准。

### 设计原则

- **主键**: Snowflake BIGINT（应用层生成）
- **无外键约束**: 应用层维护引用完整性，支持未来分库分表
- **软删除**: 关键数据使用 `is_deleted` 标记
- **分区策略**: messages 与 conversations 均按月分区（与现有迁移实现保持一致）

### 核心表清单

| 表名 | 用途 | 所属服务 | 说明 |
|------|------|----------|------|
| users | 用户基本信息 | user-service | 邮箱/手机号注册，软删除 |
| user_profiles | 账户资料与用户设置 | user-service | 1:1 关联 users；用户手动编辑入口 SoT |
| sessions | 登录会话/JWT | user-service | refresh_token 存储 |
| characters | AI 角色 | character-service | 预设 + 自定义，owner_id |
| conversations | 会话 | conversation-service | user_id + character_id |
| messages | 消息 | conversation-service | 按月分区 |
| memories | 记忆主表 | memory-service | 向量检索 + 可见性控制 + 优先级 |
| user_portraits | 对话画像 canonical | memory-service | Layer 5；Agent 记忆检索 SoT |
| emotional_states | 情感状态时序 | memory-service | Layer 2（PAD 模型：valence/arousal/dominance） |
| important_events | 重要事件 | memory-service | Layer 3（高价值事件） |
| memory_graph_nodes | 图记忆节点 | memory-service | Layer 4（实体知识） |
| memory_graph_edges | 图记忆边 | memory-service | Layer 4（关系知识，含时态字段） |
| conversation_summaries | 会话摘要 | memory-service | Layer 1（上下文压缩） |
| memory_claims | 记忆声明 | memory-service | 冲突前置缓冲与状态跟踪 |
| memory_claim_conflicts | 记忆冲突组 | memory-service | 澄清与解决闭环追踪 |
| credit_accounts | 积分账户 | **billing-service** | balance 非负约束 |
| credit_transactions | 积分流水 | **billing-service** | 幂等键（client_message_id/order_id） |
| system_configs | 系统配置 | shared | KV 存储 |

字段级约定（类型、默认值、约束、索引、迁移映射）统一以 `openspec/changes/ai-companion-platform/data-model.md` 为唯一基线，避免双份维护产生漂移。

### 级联删除策略

用户删除（30天后执行）顺序：
1. user_profiles → 2. characters (用户创建的) → 3. messages → 4. conversations → 5. memories → 6. credit_transactions → 7. credit_accounts → 8. sessions → 9. users

所有级联操作在事务中执行。

---

## Auth Persistence Notes

- 登录失败锁定（5 次失败 / 15 分钟）使用 Redis TTL 计数实现，不新增 `login_attempts` 关系表。
- 密码重置 token 与邮件链路不纳入当前 MVP，实现延期到 v1.1（见 proposal Deferred Features）。

---

## Billing Service Boundary (SoT Definition)

### 核心原则

**billing-service 是积分账本的唯一 Source of Truth (SoT)**。

任何涉及积分余额变更的操作，必须且只能通过 billing-service 的 gRPC 接口执行。其他服务（user-service, conversation-service）不得直写 credit_accounts 或 credit_transactions 表。
LiteLLM Proxy 仅负责返回 `cost_usd/tokens` 计量数据，不得直接执行积分余额扣减或充值入账。

### 服务边界

| 操作 | 责任服务 | 调用关系 |
|------|----------|----------|
| 注册赠送积分 | billing-service | user-service 注册成功后 → billing.GrantCredits（兼容期可映射到 AddCredits） |
| 预扣积分 | billing-service | conversation-service 发消息前 → billing.ReserveCredits |
| 结算积分 | billing-service | conversation-service LLM返回后 → billing.SettleCredits |
| 释放预扣 | billing-service | 超时/失败时 → billing.ReleaseCredits |
| 查询余额 | billing-service | 前端/其他服务 → gRPC 调用 billing.GetBalance |
| 充值 | billing-service | 支付回调 → billing.AddCredits（后续可收敛为 RechargeCredits） |

### 当前实现差距

现有 `backend/api/billing/v1/billing.proto` 仅有 `DeductCredits/AddCredits/SyncLLMCost` 等接口，尚未完整表达“预扣→结算→释放”语义。迁移步骤：
1. 在 `backend/api/billing/v1/billing.proto` 增补 `GrantCredits/ReserveCredits/SettleCredits/ReleaseCredits`（保留旧接口兼容）
2. user-service 注册流程改为优先调用 `billing.GrantCredits`
3. conversation-service 扣费链路改为 `billing.ReserveCredits` / `billing.SettleCredits` / `billing.ReleaseCredits`
4. 在兼容窗口结束后下线旧扣减调用（`DeductCredits` 直扣路径）

### 并发与锁策略（避免实现分歧）

- Reserve/Settle/Release 统一采用单事务行锁路径（`SELECT ... FOR UPDATE` + 条件校验），以账本行内原子更新为准。
- `credit_accounts.version` 保留为管理后台人工调账/并发审计字段，不作为主扣费链路 CAS 前提。
- 账本主链路不混用“version CAS + 条件更新”双语义，避免并发实现漂移。

### 充值链路（MVP）

- 充值套餐由 billing-service 提供只读查询接口（`credit_packages`）。
- 下单写入 `recharge_orders(PENDING)`，支付回调幂等推进到 `PAID` 并触发 `AddCredits/RechargeCredits` 入账。
- 回调重复通知必须通过 `order_no/idempotency_key` 幂等保护，禁止重复入账。

---

## Client Message ID Idempotent Protocol

### 设计目标

防止网络重试、客户端重发导致的重复扣费和重复消息。

### 协议规范

```
客户端生成 client_message_id → 发送请求(client_message_id) → 服务端检查幂等
                                                  ├─ 首次: 执行完整流程
                                                  └─ 重复: 返回已有结果
```

### client_message_id 生成规则

- **生成方**: 前端客户端
- **格式**: `{user_id}_{timestamp_ms}_{random_4chars}` 或 UUID v7
- **唯一性**: 全局唯一，作为 credit_transactions.idempotency_key
- **生命周期**: 一条用户消息对应一个 client_message_id，贯穿 预扣→LLM调用→结算 全链路

### 幂等单一真源（Single Source of Truth）

- 幂等判定统一由 `message_dedup_keys` 承担。
- `messages` 分区表不依赖全局 `UNIQUE(client_message_id)`。
- 重试请求必须先查 dedup 表，再决定返回已有结果或继续执行。

### 全链路透传

```
Frontend                    conversation-service        billing-service         llm-agent-service
   │                              │                          │                       │
   ├── SendMessage(client_message_id) ──→│                   │                       │
   │                              ├── ReserveCredits ───────→│                       │
   │                              │   (idempotency_key =     │                       │
   │                              │    client_message_id)    │                       │
   │                              │←── reserve_id ──────────│                       │
   │                              │                          │                       │
   │                              ├── Chat(client_message_id) ──────────────────────→│
   │                              │←── response + cost_usd ─────────────────────────│
   │                              │                          │                       │
   │                              ├── SettleCredits ────────→│                       │
   │                              │   (reserve_id,           │                       │
   │                              │    actual_cost)          │                       │
   │                              │←── settled ─────────────│                       │
   │←── AI Response ─────────────│                          │                       │
```

### 重试语义

| 场景 | 行为 |
|------|------|
| 相同 client_message_id 重复预扣 | 返回已有 reserve_id，不重复扣费 |
| 预扣成功但 LLM 超时 | 5分钟后自动释放预扣金额 |
| 结算时 reserve_id 已结算 | 返回已有结算结果 |
| 客户端断连重发 | 服务端查到已有消息记录，直接返回 |

---

## Conversation ↔ Memory Integration

### 设计决策

SendMessage 主链路与长期记忆写入链路解耦，采用 **异步事件驱动 + Outbox 可靠投递**：

1. conversation-service 在同一事务内提交：消息持久化 + `outbox_events(status=PENDING)`。
   同时维护 `conversations.message_count/token_count/last_message_at` 聚合字段，避免统计字段长期漂移。
2. relay worker 异步读取 outbox 并发布 `evt.memory.write.requested` 到 Redis Stream。
3. memory-service 消费事件并执行 Layer1-5 写入编排。
4. 投递/消费失败进入重试队列，不阻塞用户主响应。

### 事件契约（最小集）

| 字段 | 说明 |
|------|------|
| user_id | 用户 ID |
| conversation_id | 会话 ID |
| message_id | 服务端消息 ID（Snowflake） |
| client_message_id | 客户端幂等键 |
| assistant_reply | AI 回复文本（可裁剪） |

### Claim 冲突异步编排（非阻塞主链路）

`memory_claims + 冲突状态机` 复杂度被显式放在 memory-service 的事件消费链路，不进入 SendMessage 同步路径：

1. **Extract**：从 `evt.memory.write.requested` 提取候选 claim，写入 `memory_claims(status=PENDING)`
2. **Detect**：按 `(user_id, character_id, claim_key)` 比对当前有效 claim；冲突时创建 `memory_claim_conflicts(status=OPEN)`，并将新 claim 标记为 `CONFLICTING`
3. **Clarify**：claim 进入 `CLARIFYING`，并发布 `evt.memory.conflict.clarify.requested`；由 llm-agent 生成人设化澄清问句，conversation-service 作为下一轮对话发送
4. **Resolve**：`memory_claim_conflicts` 转移到 `RESOLVED_REPLACED` / `RESOLVED_APPENDED` / `RESOLVED_REJECTED`，并驱动 claim 转移到 `CONFIRMED` / `MERGED` / `REJECTED`，再回写 `user_portraits` 与 `memories`

### 检索注入安全门（Hard Filter First）

长期记忆检索采用固定三段式，权限过滤为硬门禁，不降级为评分因子：

1. **Stage 1 Recall**：Vector + BM25 + Graph 多路召回
2. **Stage 2 Hard Gate**：强制 `visibility_filter + taboo_filter`，未通过项直接丢弃并记录审计事件
3. **Stage 3 Re-rank**：只对通过 Stage 2 的候选做相关性重排并注入 Prompt

### 时态关系表达（补回 001 基线）

对会变化的关系边（friend/enemy/blocked）不做覆盖写，统一使用 `memory_graph_edges.valid_from / valid_until` 表达演化，支持“曾经是…后来变成…”的可追溯叙事。

### 失败策略

- outbox 投递失败：保留 `outbox_events` 记录并指数退避重试，主链路仍返回成功
- outbox 重试上限：按事件记录 `max_retries` 与全局默认阈值共同生效，超过阈值标记 `DEAD`
- 消费失败：memory-service 按指数退避重试，超过阈值写入审计事件
- Claim 编排失败：Extract/Detect/Clarify/Resolve 各阶段独立重试，超过阈值写入 `evt.memory.claim.dlq`
- 幂等保障：memory-service 以 `client_message_id + claim_key + stage` 作为去重键

---

## Gateway & Deployment Baseline

- APISIX 必须启用 CORS 白名单（按环境区分 `dev/staging/prod` Origin），允许前端 SPA 跨域访问 `/api/v1/*`。
- LiteLLM Proxy 作为独立部署单元（Compose/K8s），llm-agent 通过内网地址调用并使用密钥配置。
- LiteLLM 健康检查与超时策略纳入部署验收：不可用时 llm-agent 返回可观测错误并触发降级策略。

---

## WebSocket Architecture

### MVP 单实例方案

```
Frontend ──WebSocket──→ conversation-service (Go)
                              │
                              ├── 维护连接池 (sync.Map)
                              ├── JWT 鉴权 (首次连接)
                              └── 消息推送 (新消息/状态变更)
```

### 连接管理

- **鉴权**: WebSocket 连接建立时通过 query param 或首条消息携带 JWT
- **心跳**: 30s ping/pong，超时断开
- **重连**: 客户端指数退避重连（1s, 2s, 4s, 8s, max 30s）
- **多端**: 同一 user_id 可维护多个连接，消息广播到所有连接

### 多实例扩展方案（v1.1）

```
Frontend ──→ APISIX (sticky session) ──→ conversation-service (N instances)
                                              │
                                              ├── Redis Pub/Sub (跨实例广播)
                                              └── 连接注册表 (Redis Hash)
```

MVP 阶段先做单实例可用，连接池存内存。v1.1 引入 Redis Pub/Sub 实现跨实例广播。

### 连接降级策略

- **Primary**: WebSocket 实时推送。
- **Fallback**: WebSocket 建连失败或连续心跳超时时，客户端自动降级到 3-5 秒轮询。
- **恢复**: 轮询期间持续指数退避重连 WebSocket，连接恢复后停止轮询。

---

## Frontend Architecture Upgrade

### 状态管理改造

| 层面 | 当前 (spec-kit) | 目标 (openspec) |
|------|------------------|-----------------|
| UI 状态 | Context API | **Context API** (保留) |
| 服务端数据 | Context API (ChatContext) | **TanStack Query** |
| 认证状态 | AuthContext | **AuthContext** (保留，对接真实API) |
| 实时消息 | useMessagePolling | **useChatSocket** (WebSocket) |

### TanStack Query 集成

```typescript
// 查询键约定
const queryKeys = {
  characters: ['characters'] as const,
  characterDetail: (id: string) => ['characters', id] as const,
  conversations: (characterId: string) => ['conversations', characterId] as const,
  messages: (conversationId: string) => ['messages', conversationId] as const,
  userProfile: ['user', 'profile'] as const,
  creditBalance: ['user', 'credits'] as const,
}
```

### 需保留的 Context

- `ThemeContext` - 深色/浅色模式
- `LanguageContext` - 中英文切换
- `AuthContext` - JWT 管理（改为对接真实 API）

### 需废弃/重构的 Context

- `ChatContext` - 消息列表管理改为 TanStack Query + WebSocket 事件驱动

---

## Test Matrix

### 测试策略

每个实现任务必须配最小测试单元，不允许仅依赖最终阶段验收。

| 层面 | 工具 | 覆盖目标 | 触发时机 |
|------|------|----------|----------|
| Go 单元测试 | `go test` + testify | ≥70% per service | 每个任务完成时 |
| Go 集成测试 | testcontainers-go | 跨服务调用 | Phase 完成时 |
| 前端单元测试 | Vitest + @testing-library/react | 关键组件 | 每个组件完成时 |
| 前端 E2E 测试 | Playwright | 核心用户流程 | Phase 4/5 |
| PBT 属性测试 | rapid (Go) | 不变量验证 | Phase 5 |

### PBT 不变量清单

| 不变量 | 对应 Design 属性 | 测试文件 |
|--------|------------------|----------|
| credit_accounts.balance >= 0 | 余额非负 | `tests/pbt/billing_test.go` |
| 同一 client_message_id 只产生一条 RESERVE 流水 | 扣费幂等 | `tests/pbt/billing_test.go` |
| RESERVE 金额 >= SETTLE 金额 | 预扣结算一致 | `tests/pbt/billing_test.go` |
| 分页查询消息不丢不重 | 消息单调 | `tests/pbt/message_test.go` |
| Owner-Private 记忆不出现在非 Owner 查询 | 记忆隔离 | `tests/pbt/memory_test.go` |

### Proto 合约关键字段

```protobuf
// backend/api/billing/v1/billing.proto 目标关键 RPC（兼容期保留旧接口）
service BillingService {
  rpc GrantCredits(GrantCreditsRequest) returns (GrantCreditsReply);
  rpc ReserveCredits(ReserveCreditsRequest) returns (ReserveCreditsReply);
  rpc SettleCredits(SettleCreditsRequest) returns (SettleCreditsReply);
  rpc ReleaseCredits(ReleaseCreditsRequest) returns (ReleaseCreditsReply);
  rpc GetBalance(GetBalanceRequest) returns (GetBalanceReply);
  rpc DeductCredits(DeductCreditsRequest) returns (DeductCreditsResponse); // legacy
  rpc AddCredits(AddCreditsRequest) returns (AddCreditsResponse);           // legacy
}

message ReserveCreditsRequest {
  int64 user_id = 1;
  string idempotency_key = 2;  // = client_message_id
  int64 estimated_amount = 3;  // 预估积分（基于历史平均或固定值）
}

message ReserveCreditsReply {
  int64 reserve_id = 1;
  int64 reserved_amount = 2;
  int64 balance_after = 3;
}

message SettleCreditsRequest {
  int64 reserve_id = 1;
  int64 actual_amount = 2;     // 实际积分消耗
  string cost_usd = 3;         // LLM 原始 USD 成本
  int32 token_count = 4;
}
```
