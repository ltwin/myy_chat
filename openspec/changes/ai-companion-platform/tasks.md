# Tasks: AI 情感伴侣平台

**Change ID**: `ai-companion-platform`
**Created**: 2026-02-01
**Method**: Zero-Decision Implementation Plan

---

## Phase 0: P0 风险修复 (阻塞项)

### T0.1 更新 Go Dockerfile 基础镜像
- **文件**: `backend/app/*/Dockerfile`
- **操作**: 将 `FROM golang:1.23-alpine` 改为 `FROM golang:1.24-alpine`
- **验收**: 所有 Dockerfile 使用 1.24 版本

### T0.2 统一健康检查路径
- **文件**:
  - Go 服务: `backend/app/*/internal/server/http.go`
  - docker-compose: `deployments/docker-compose.dev.yml`
  - Helm: `deployments/k8s/*/values.yaml`
- **操作**: 统一使用 `/api/v1/health`
- **验收**: 所有探针配置一致

### T0.3 整理数据库迁移脚本（消除编号歧义）
- **文件**: `backend/migrations/`
- **操作**:
  - 保留唯一执行链脚本：`000_init.sql`、`001_create_users.sql`、`002_create_characters.sql`、`004_create_conversations_table.sql`、`005_create_messages_table.sql`、`006_seed_characters.sql`
  - 删除重复脚本：`001_create_users_table.sql`、`002_create_user_profiles_table.sql`、`003_create_characters_table.sql`
  - 修正 `004_create_conversations_table.sql` 头部 Dependencies 注释为 `001_create_users.sql, 002_create_characters.sql`
  - 添加 `IF NOT EXISTS` 幂等检查，并将历史 trigger 创建改为幂等模式（`DROP TRIGGER IF EXISTS` + `CREATE TRIGGER`）
- **验收**:
  - `migrate up` 可重复执行无错误
  - 新成员按清单执行时无需再做“保留哪个脚本”的临场决策

### T0.4 校验 Go 版本一致性
- **文件**: `backend/go.mod`, `backend/app/*/Dockerfile`
- **操作**: 确保 `go.mod` 固定为 `go 1.24.0`，且全部构建镜像统一为 `golang:1.24-alpine`
- **验收**: 仓库中不再出现 `1.23` / `1.24` 混用

### T0.5 旧 Python llm-agent 冻结与标识
- **文件**: `backend/python/llm-agent-service/`, `README.md`, `docs/`
- **操作**: 标记 legacy 路径为冻结状态（只读维护），新增迁移说明与禁改提示
- **验收**: 文档明确“Go 版本为唯一演进路径”，Python 目录不再新增功能任务

### T0.6 Python llm-agent 构建与部署剔除
- **文件**: `Makefile`, `deployments/`, CI 工作流配置
- **操作**: 移除 Python llm-agent 的构建/发布入口，仅保留 Go llm-agent 交付链路
- **验收**: CI/CD 与部署脚本不再引用 Python llm-agent 服务

---

## Phase 1: 基础设施层

### T1.1 Snowflake ID 生成器
- **文件**: `backend/pkg/snowflake/snowflake.go`
- **操作**: 实现支持 DATACENTER_ID + WORKER_ID 的 Snowflake 生成器
- **验收**: 单元测试通过，ID 唯一且有序

### T1.2 统一错误码规范
- **文件**: `backend/pkg/errors/errors.go`
- **操作**: 定义 code/message/details schema
- **验收**: 所有服务使用统一错误格式

### T1.3 Redis 连接池
- **文件**: `backend/pkg/redis/client.go`
- **操作**: 实现连接池，支持分布式锁
- **验收**: 连接池测试通过

### T1.4 OTel 集成
- **文件**: `backend/pkg/tracing/otel.go`
- **操作**: 集成 OpenTelemetry，支持跨服务 trace
- **验收**: trace 可在 Jaeger 中查看

### T1.5 Redis Stream Consumer Group 初始化
- **文件**: `backend/pkg/redis/stream_bootstrap.go`
- **操作**: 启动时幂等创建 `evt.memory.write.requested` 与 `evt.memory.conflict.clarify.requested` 对应 consumer group
- **验收**: 重复启动不报错，memory-service 可稳定消费

### T1.6 LiteLLM 基础设施配置
- **文件**: `deployments/docker-compose.dev.yml`, `deployments/k8s/`, `backend/app/llm-agent/configs/`
- **操作**: 定义 LiteLLM Proxy 服务、内网地址、鉴权密钥与健康检查
- **验收**: llm-agent 可稳定探测 LiteLLM 健康状态并成功发起调用

---

## Phase 2: 核心服务实现

### 2.1 user-service (Go)

#### T2.1.1 用户注册
- **文件**: `backend/app/user/internal/biz/user.go`
- **操作**: 实现邮箱/手机号注册，格式验证
- **测试**: 单元测试覆盖正常注册、重复邮箱、格式无效
- **验收**: 注册 API 返回用户 ID

#### T2.1.2 密码哈希
- **文件**: `backend/app/user/internal/biz/user.go`
- **操作**: 使用 bcrypt cost=12 哈希密码
- **测试**: 验证哈希可逆性、cost 参数
- **验收**: 密码存储为哈希值

#### T2.1.3 JWT 签发
- **文件**: `backend/app/user/internal/biz/session.go`
- **操作**: 签发 JWT，TTL=15min，claims 含 user_id
- **测试**: 验证 token 解析、过期、无效签名
- **验收**: 登录返回有效 JWT

#### T2.1.4 登录锁定
- **文件**: `backend/pkg/middleware/login_lockout.go`
- **操作**: 使用 Redis TTL 计数实现“5次失败后锁定15分钟”
- **测试**: 边界测试（4次不锁、5次锁、15分钟后解锁）
- **验收**: 锁定机制测试通过

#### T2.1.5 账号删除
- **文件**: `backend/app/user/internal/biz/user.go`
- **操作**: 标记 pending_delete，30天后删除
- **测试**: 状态机测试（标记→冷静期→恢复/删除）
- **验收**: 删除状态机测试通过

#### T2.1.6 注册时调用 billing 赠送积分
- **文件**: `backend/app/user/internal/biz/user_service.go`
- **操作**: 注册成功后 gRPC 调用 `billing.GrantCredits(user_id, 100)`
- **依赖**: T2.2.1, T2.2.2
- **测试**: mock billing client，验证调用参数
- **验收**: 注册后 billing-service 记录赠送流水

#### T2.1.7 user_profiles 手动编辑与画像 claim 同步
- **文件**: `backend/app/user/internal/service/profile_service.go`
- **操作**: 用户手动编辑 `user_profiles` 后发布 `evt.profile.user_stated.updated`，供 memory-service 以 `USER_STATED` 来源进入 claim 决议流
- **依赖**: T2.4M.2a
- **测试**: 编辑触发事件、重复提交幂等、字段级 payload 校验
- **验收**: 账号资料编辑可驱动画像决议链路，且不直接覆盖 `user_portraits` canonical

### 2.2 billing-service (Go) — 新建独立服务

#### T2.2.0 billing-service 服务骨架
- **文件**: `backend/app/billing/cmd/`, `backend/app/billing/internal/`, `backend/app/billing/configs/`
- **操作**: 使用 Kratos 模板创建独立服务，含 gRPC + HTTP server
- **验收**: `make run-billing` 启动无错误，健康检查通过

#### T2.2.0a billing Proto 定义
- **文件**: `backend/api/billing/v1/billing.proto`
- **操作**: 扩展 BillingService RPC（新增 GrantCredits, ReserveCredits, SettleCredits, ReleaseCredits, 保留 DeductCredits/AddCredits 兼容）
- **验收**: proto 编译通过，Go 代码生成完整

#### T2.2.1 积分账户模型
- **文件**: `backend/app/billing/internal/biz/credit_account.go`
- **操作**: balance BIGINT，非负约束，CHECK (balance >= 0)
- **测试**: 余额非负边界测试
- **验收**: 数据模型测试通过

#### T2.2.2 注册赠送积分
- **文件**: `backend/app/billing/internal/biz/credit_account.go`
- **操作**: GrantCredits RPC，新用户 +100 积分，幂等（同一 user_id 只赠一次）
- **测试**: 重复调用幂等验证
- **验收**: 注册后余额为 100

#### T2.2.3 预扣接口
- **文件**: `backend/app/billing/internal/service/billing_service.go`
- **操作**: ReserveCredits RPC，幂等键=client_message_id，原子条件更新 `balance >= amount`
- **测试**: 并发预扣测试、幂等重复请求测试、余额不足测试
- **验收**: 重复请求返回相同 reserve_id

#### T2.2.4 结算接口
- **文件**: `backend/app/billing/internal/service/billing_service.go`
- **操作**: SettleCredits RPC，多退少补逻辑，关联 reserve_id
- **测试**: 实际 < 预扣（退差额）、实际 = 预扣、重复结算幂等
- **验收**: 结算后余额正确

#### T2.2.5 释放预扣
- **文件**: `backend/app/billing/internal/service/billing_service.go`
- **操作**: ReleaseCredits RPC，超时/失败时释放预扣金额
- **测试**: 正常释放、已结算的释放应拒绝
- **验收**: 释放后余额恢复

#### T2.2.6 交易流水
- **文件**: `backend/app/billing/internal/data/transaction.go`
- **操作**: 完整审计日志，transaction_type: RESERVE/SETTLE/RELEASE/RECHARGE/GRANT
- **测试**: 每种类型的流水写入验证
- **验收**: 所有操作有流水记录

#### T2.2.7 预扣超时自动释放
- **文件**: `backend/app/billing/internal/biz/credit_account.go`
- **操作**: 定时任务扫描 5 分钟未结算的 RESERVE 记录，自动释放
- **测试**: 模拟超时场景
- **验收**: 超时预扣自动释放

#### T2.2.8 充值套餐查询
- **文件**: `backend/app/billing/internal/service/billing_service.go`
- **操作**: 提供 `ListCreditPackages` 只读接口，返回启用中的充值套餐
- **测试**: 上下架过滤、排序稳定性、空结果
- **验收**: 前端可获取可售套餐列表

#### T2.2.9 充值下单与幂等
- **文件**: `backend/app/billing/internal/biz/recharge.go`
- **操作**: 创建 `recharge_orders(PENDING)`，按 `idempotency_key` 幂等去重
- **测试**: 重复下单返回同一订单、非法套餐拒绝、金额校验
- **验收**: 下单接口可稳定返回订单号与待支付状态

#### T2.2.10 支付回调入账
- **文件**: `backend/app/billing/internal/service/payment_callback.go`
- **操作**: 支付成功回调将订单推进 `PAID` 并调用入账流程（`RECHARGE` 流水），重复回调不得重复入账
- **测试**: 正常回调、重复回调幂等、失败回调状态保持
- **验收**: 充值后余额与流水一致，重复通知不重复加币

### 2.3 character-service (Go)

#### T2.3.1 预设角色列表
- **文件**: `backend/app/character/internal/biz/character.go`
- **操作**: 至少 5 个预设角色，含 seed 迁移脚本
- **测试**: 列表返回数量、预设角色字段完整性
- **验收**: 列表 API 返回 ≥5 个角色

#### T2.3.2 自定义角色 CRUD
- **文件**: `backend/app/character/internal/service/character_service.go`
- **操作**: 支持名称/性格/背景/system_prompt
- **测试**: CRUD 全路径 + 软删除验证
- **验收**: CRUD API 测试通过

#### T2.3.3 Owner 归属
- **文件**: `backend/app/character/internal/biz/character.go`
- **操作**: owner_id 字段，权限检查（非 Owner 不能改/删）
- **测试**: 跨用户操作拒绝测试
- **验收**: 非 Owner 无法修改

### 2.4 llm-agent-service (Go)

#### T2.4.1 服务骨架
- **文件**: `backend/app/llm-agent/cmd/`, `backend/app/llm-agent/internal/`, `backend/app/llm-agent/configs/`
- **操作**: 使用 Kratos 模板创建 llm-agent-service（gRPC + HTTP health）
- **验收**: `make run-llm-agent` 启动无错误，健康检查通过

#### T2.4.2 gRPC 服务实现
- **文件**: `backend/app/llm-agent/internal/service/llm_agent_service.go`
- **操作**: 实现 Chat RPC（含 streaming），对接 LiteLLM Proxy
- **测试**: mock LiteLLM 的 gRPC/HTTP 调用测试
- **验收**: gRPC Chat 服务可调用

#### T2.4.3 Prompt 组装
- **文件**: `backend/app/llm-agent/internal/biz/prompt_builder.go`
- **操作**: 角色设定 + 基础记忆 + 安全约束 + 角色一致性指令
- **测试**: 验证 prompt 包含所有必要 section
- **验收**: Prompt 包含所有必要部分

#### T2.4.4 流式输出
- **文件**: `backend/app/llm-agent/internal/biz/streaming.go`
- **操作**: gRPC streaming 按 chunk 顺序返回
- **测试**: 验证流式 chunk 顺序和完整性
- **验收**: TTFT ≤ 3s

#### T2.4.5 成本回传
- **文件**: `backend/app/llm-agent/internal/biz/cost.go`
- **操作**: 返回 cost_usd + tokens（从 LiteLLM response 提取）
- **测试**: 有 cost 和无 cost（兜底估算）场景
- **验收**: 响应包含成本信息

### 2.4M long-term-memory-core (Go)

#### T2.4M.0 [P0] memory-service 服务骨架
- **文件**: `backend/app/memory/cmd/`, `backend/app/memory/internal/`, `backend/app/memory/configs/`
- **操作**:
  1. 按 Kratos 标准骨架清单核对 memory-service（`cmd/main.go`、`wire.go`、`wire_gen.go`、`internal/{biz,data,service,server,conf}`、`configs/config.yaml`）
  2. 若已有目录则做“缺项补齐”，若缺失则按模板创建
  3. 校验 `make run-memory` 启动目标与 `/api/v1/health` 健康检查
- **验收**: `make run-memory` 启动无错误，健康检查通过

#### T2.4M.0a [P0] memory Proto 定义
- **文件**: `backend/api/memory/v1/memory.proto`
- **操作**: 将当前 Hello 示例升级为 MemoryService，至少覆盖 `MemoryWrite`、`MemoryQuery`、`MemoryList`、`MemoryUpdate`、`MemoryDelete`、`AuditEventQuery` RPC，并补齐 `client_message_id`、`visibility`、审计元数据字段
- **验收**: proto 编译通过，Go 代码生成完整，前后端接口命名一致

#### T2.4M.1 [P0] memory-service 核心模型
- **文件**: `backend/app/memory/internal/biz/`
- **操作**: 建立 user_portraits / emotional_states / important_events / memory_graph_nodes / memory_graph_edges / conversation_summaries / memory_claims / memory_claim_conflicts 领域模型；补齐 `memories.actor_user_id/access_count/source_type/target_table/target_id` 与 `memory_graph_nodes.normalized_name/aliases`，并补回 `memory_graph_edges.valid_from/valid_until/confidence` 语义
- **依赖**: T2.4M.0, T2.6.3, T2.6.4
- **测试**: 各实体约束与状态迁移单元测试
- **验收**: 五层实体模型 + claim 模型测试通过

#### T2.4M.2 [P0] 记忆写入流水线
- **文件**: `backend/app/memory/internal/service/memory_service.go`, `backend/app/memory/internal/biz/pipeline.go`
- **操作**:
  1. 消费 `evt.memory.write.requested` 异步事件并执行多层写入编排（摘要、画像、事件、情感、图节点）
  2. 先落 `memory_claims(status=PENDING)`，再进入冲突编排，避免阻塞 SendMessage 主链路
  3. 以 `client_message_id` 做跨层幂等去重
- **依赖**: T2.4M.0a, T2.4M.1, T1.3, T1.5
- **测试**: 写入成功、部分失败补偿、重复消息幂等
- **验收**: 新消息完成后可通过异步事件触发记忆写入链路

#### T2.4M.2a [P0] Claim 冲突检测编排
- **文件**: `backend/app/memory/internal/biz/claim_pipeline.go`
- **操作**:
  1. 按 `(user_id, character_id, claim_key)` 与当前生效 claim 做冲突检测
  2. 冲突时创建 `memory_claim_conflicts(status=OPEN)`，并将新 claim 标记为 `CONFLICTING`
  3. 发布 `evt.memory.conflict.clarify.requested`，等待人设化澄清链路
- **依赖**: T2.4M.1, T2.4M.2
- **测试**: 冲突检测准确性、重复事件幂等、无冲突自动确认
- **验收**: claim/conflict 双状态机稳定流转（claim: `PENDING/CONFLICTING/CLARIFYING/CONFIRMED|REJECTED|MERGED`; conflict: `OPEN/CLARIFYING/RESOLVED_*`）

#### T2.4M.2b [P0] Claim 状态机重试与死信
- **文件**: `backend/app/memory/internal/biz/claim_retry_worker.go`
- **操作**: 为 Extract/Detect/Clarify/Resolve 四阶段配置指数退避重试，统一幂等键 `client_message_id + claim_key + stage`，超阈值写入 `evt.memory.claim.dlq`
- **依赖**: T2.4M.2a, T1.3
- **测试**: 阶段失败重试、死信入队、重放恢复
- **验收**: 冲突链路失败可恢复且主链路不受阻

#### T2.4M.3 [P0] 记忆冲突澄清
- **文件**: `backend/app/llm-agent/internal/biz/prompt_builder.go`
- **操作**: 消费 `evt.memory.conflict.clarify.requested`，输出人设化澄清问句与结构化确认选项，禁止暴露“冲突检测”等系统措辞
- **依赖**: T2.4M.2a
- **测试**: 矛盾信息场景 prompt 结构校验
- **验收**: 冲突场景回复符合角色人设并可驱动状态机决议

#### T2.4M.4 [P0] 可见性隔离检索
- **文件**: `backend/app/memory/internal/data/memory_query.go`
- **操作**: 固化 Recall → Hard Gate → Re-rank 检索链路；按 Owner-Private/Public/Visitor-Private 强制过滤并在 prompt 注入前二次校验，`visibility_filter` 不参与评分
- **测试**: owner/visitor 混合访问隔离测试
- **验收**: 非 Owner 无法检索私密记忆，越权候选在重排前已被剔除

#### T2.4M.4a [P0] 安全审计事件最小闭环
- **文件**: `backend/app/memory/internal/biz/audit_event.go`, `backend/app/memory/internal/service/memory_service.go`
- **操作**: 统一记录越权检索拒绝、冲突状态迁移、claim 死信、图谱同步失败四类事件，提供按 `user_id + time_range` 查询接口
- **测试**: 事件写入、重复上报幂等、查询分页
- **验收**: 安全事件可追溯查询

#### T2.4M.5 压缩与优先级维护任务
- **文件**: `backend/app/memory/internal/biz/maintenance.go`
- **操作**: 上下文阈值触发对话压缩，定时执行遗忘衰减与强化提升
- **测试**: 压缩触发、衰减曲线、重要事件保留
- **验收**: 高价值记忆稳定保留，低价值旧记忆优先级下降

#### T2.4M.6 图投影同步作业（Postgres 内部）
- **文件**: `backend/app/memory/internal/biz/graph_sync.go`
- **操作**: 基于 `memory_graph_nodes/edges` 执行 Postgres 内部图投影同步与状态收敛（`sync_status/sync_version`），支持重试与幂等版本控制
- **测试**: 重复同步不重复建边、失败重试恢复
- **验收**: sync pending 队列可清空

#### T2.4M.7 [P0] 记忆管理 API
- **文件**: `backend/app/memory/internal/service/memory_service.go`
- **操作**: 提供 MemoryList / MemoryUpdate / MemoryDelete 接口（带审计字段）
- **测试**: 列表分页、编辑授权、删除回读一致性
- **验收**: 前端可用记忆管理接口稳定返回

#### T2.4M.8 [P0] 画像字段分级策略
- **文件**: `backend/app/memory/internal/biz/profile_field_policy.go`
- **操作**: 实现 Tier A/B/C 字段分级规则（单值关键字段、多值共存字段、推断弱信号）
- **测试**: 字段分类命中、未知字段回退策略、策略加载测试
- **验收**: 写入流水线可按字段分级执行不同更新策略

#### T2.4M.9 [P0] Tier A 单值字段冲突门禁
- **文件**: `backend/app/memory/internal/biz/claim_pipeline.go`
- **操作**: 对 `full_name`、`birth_date`、`current_occupation` 等单值关键字段启用“先 claim 后澄清”门禁，未决议前禁止覆盖 canonical
- **依赖**: T2.4M.2a, T2.4M.8
- **测试**: 冲突单值字段不覆盖、澄清后覆盖、拒绝后保持旧值
- **验收**: Tier A 字段仅在明确决议后更新主画像

#### T2.4M.10 [P1] Tier B 多值字段共存与排序
- **文件**: `backend/app/memory/internal/biz/portrait_merge.go`
- **操作**: interests/skills/role_tags 等多值事实支持共存，按时间与权重排序输出
- **依赖**: T2.4M.8
- **测试**: 多值追加、重复去重、排序稳定性
- **验收**: 多值冲突不触发 destructive overwrite

#### T2.4M.11 [P1] 职业时态投影
- **文件**: `backend/app/memory/internal/biz/occupation_projection.go`
- **操作**: 维护 `current_occupation` 与 `career_history`（或等价时态结构），支持职业变更保留历史
- **依赖**: T2.4M.9
- **测试**: 程序员→产品经理变更链路、当前值投影、历史可追溯
- **验收**: 职业变更场景下“当前值正确 + 历史不丢失”

#### T2.4M.12 [P0] mcp-service 通用网关骨架
- **文件**: `backend/api/mcp/v1/mcp.proto`, `backend/app/mcp/cmd/`, `backend/app/mcp/internal/`, `backend/app/mcp/configs/`
- **操作**:
  1. 定义 `mcp/v1` Proto 合约（`tools.list` / `tools.call` + 标准错误码）并完成代码生成
  2. 基于 Kratos 建立 mcp-service 核心网关（认证/ACL/审计/限流/错误规范化中间件）
  3. 实现工具注册中心（`tool_name -> domain -> handler -> required_scopes -> downstream_service`）
  4. 提供 `tools.list` 与 `tools.call` 统一入口，并支持按 scope 过滤可见工具
- **验收**: `make run-mcp` 启动正常，健康检查通过

#### T2.4M.13 [P0] MCP 画像工具与 SoT 边界
- **文件**: `backend/app/mcp/internal/service/profile_tools.go`
- **操作**:
  1. 在当前变更范围内，以 `profile` 域插件方式注册 `profile.get/profile.claim_upsert/profile.conflict_list/profile.conflict_resolve/profile.fact_delete/profile.audit_query`
  2. 写入统一透传到 memory-service，禁止直写数据库
  3. 服务端强制注入身份上下文（user/tenant/character/request_id），忽略 client 伪造身份字段
- **依赖**: T2.4M.7, T2.4M.8, T2.4M.12
- **测试**: 工具调用鉴权、幂等透传、越权拒绝、SoT 边界校验
- **验收**: Agent 侧画像 CRUD 全部经 MCP 层完成且可审计

### 2.5 conversation-service (Go)

#### T2.5.1 会话管理
- **文件**: `backend/app/conversation/internal/biz/conversation.go`
- **操作**: CRUD + 游标分页（基于 Snowflake ID 单调递增特性）
- **测试**: 分页不丢不重、空列表、边界条件
- **验收**: 分页 API 测试通过

#### T2.5.2 消息持久化
- **文件**: `backend/app/conversation/internal/data/message.go`
- **操作**: 服务端 Snowflake `message.id` + 客户端 `client_message_id` 唯一键，按月分区表写入
- **测试**: 消息写入和读取顺序一致性
- **验收**: 消息有序存储

#### T2.5.3 LLM 客户端
- **文件**: `backend/app/conversation/internal/data/llm_client.go`
- **操作**: gRPC 调用 llm-agent，支持 streaming response
- **测试**: mock llm-agent 的流式响应测试
- **验收**: 调用返回 AI 回复

#### T2.5.4 扣费集成（预扣→LLM→结算）
- **文件**: `backend/app/conversation/internal/biz/conversation.go`
- **操作**:
  1. 预扣: gRPC 调用 billing.ReserveCredits(client_message_id)
  2. LLM 调用: gRPC 调用 llm-agent.Chat
  3. 结算: gRPC 调用 billing.SettleCredits(reserve_id, actual_cost)
  4. 异常: LLM 失败时 billing.ReleaseCredits(reserve_id)
- **依赖**: T2.2.3, T2.2.4, T2.2.5, T2.4.2
- **测试**: 正常流程、LLM 失败释放、余额不足拒绝
- **验收**: 扣费流程端到端测试通过

#### T2.5.5 WebSocket 推送
- **文件**: `backend/app/conversation/internal/server/websocket.go`
- **操作**:
  - gorilla/websocket 集成
  - JWT 鉴权（连接建立时）
  - 连接池管理（sync.Map）
  - 心跳 30s ping/pong
  - 新消息/流式 chunk 推送
- **测试**: 连接建立、消息推送、断线重连
- **验收**: 多端实时同步

#### T2.5.6 SendMessage 端到端流程
- **文件**: `backend/app/conversation/internal/service/conversation_service.go`
- **操作**:
  1. 整合 T2.5.2 + T2.5.3 + T2.5.4 + T2.5.5，实现发送主链路
  2. 在同一数据库事务内完成“消息落库 + `outbox_events` 插入”（event_type=`evt.memory.write.requested`，payload 固定：`user_id`、`conversation_id`、`message_id`、`client_message_id`、`assistant_reply`）
  3. 订阅 `evt.memory.conflict.clarify.requested`，在后续对话轮次注入人设化澄清问句
  4. outbox 投递/消费失败不阻塞主链路响应，写入对应重试队列并记录告警日志
- **参数**: 前端传入 client_message_id（幂等键）
- **依赖**: T2.4M.2, T2.4M.2a, T1.3
- **测试**: 端到端集成测试（含 mock billing + mock llm-agent + mock Redis Stream + outbox relay）
- **验收**: 发送消息 → 预扣 → LLM → 结算 → 异步记忆事件发布/澄清回流 → WebSocket 推送

#### T2.5.7 Outbox Relay Worker
- **文件**: `backend/app/conversation/internal/biz/outbox_relay.go`
- **操作**:
  1. 扫描 `outbox_events` 中 `PENDING/FAILED` 记录并投递 Redis Stream
  2. 成功投递后更新为 `PUBLISHED`，失败按指数退避重试
  3. 达到 `max_retries` 或全局阈值后标记 `DEAD` 并记录告警
- **依赖**: T1.3, T2.5.6
- **测试**: 重试语义、幂等投递、服务重启后的续传恢复
- **验收**: 主链路成功后 `evt.memory.write.requested` 最终可达

#### T2.5.8 会话聚合字段维护
- **文件**: `backend/app/conversation/internal/biz/conversation.go`
- **操作**: 在消息写入事务内同步维护 `conversations.message_count/token_count/last_message_at`
- **依赖**: T2.5.2
- **测试**: 多轮对话累计、并发发送一致性、回滚场景不脏写
- **验收**: 会话聚合统计与消息明细一致

### 2.6 数据库迁移脚本补充（执行顺序前置）

> 说明：虽然编号位于 2.5 后，但实现顺序上应在 2.2~2.5 核心功能开发前完成并验收。

#### T2.6.1 billing 相关表迁移
- **文件**: `backend/migrations/007_billing_reserve_settle.sql`
- **操作**: 在现有账本表基础上补充预扣/结算字段与索引（幂等执行，基于 T0.3 固定基线，避免与 `000-006` 冲突）
- **验收**: `migrate up` 幂等执行通过

#### T2.6.2 messages 分区表
- **文件**: `backend/migrations/008_messages_client_message_id.sql`
- **操作**: 在现有分区 messages 表上补齐 `client_message_id` 索引；新增 `message_dedup_keys` 全局幂等键表与 `outbox_events` 可靠投递表（含 `max_retries` 字段），并补齐 2026 分区
- **验收**: 写入不同月份的消息到正确分区

#### T2.6.3 长期记忆核心表迁移
- **文件**: `backend/migrations/009_long_term_memory_core.sql`
- **操作**: 新增 user_portraits / emotional_states / important_events / memory_graph_nodes / memory_graph_edges / conversation_summaries 表及索引；`user_portraits` 明确包含 `current_occupation` 与 `career_history`，`emotional_states.dominance` 增加范围约束，并补回 `memory_graph_edges.valid_from/valid_until` 时态字段
- **验收**: 六张长期记忆表与时态边字段迁移成功

#### T2.6.4 记忆主表扩展与约束
- **文件**: `backend/migrations/010_memory_extensions.sql`
- **操作**: 扩展 memories 表（`actor_user_id/access_count/source_type/target_table/target_id/visibility_scope`）与审计事件类型；新增 `memory_claims` / `memory_claim_conflicts`，并补齐 `field_tier/rank_score/valid_from/valid_until` 字段、`uq_memory_claims_dedup` 唯一索引、HNSW 向量索引与必要查询索引
- **验收**: 记忆检索、冲突状态机与维护任务字段齐备

#### T2.6.5 system_configs 种子迁移
- **文件**: `backend/migrations/011_system_configs_seed.sql`
- **操作**: 初始化 `billing.credit_formula`、`memory.compression.threshold`、`memory.decay.default_factor` 等关键配置（`ON CONFLICT` 幂等）
- **验收**: 新环境启动无需手工补关键配置

---

## Phase 3: 集成与网关

### T3.1 APISIX 路由配置
- **文件**: `deployments/apisix/apisix.yaml`
- **操作**: 显式配置 user/character/conversation/billing/memory/mcp 路由（`/api/v1/*`）及 upstream 健康检查
- **验收**: 路由测试通过

### T3.2 JWT 插件配置
- **文件**: `deployments/apisix/apisix.yaml`
- **操作**: 与 user-service 对齐 secret/issuer
- **验收**: JWT 验证通过

### T3.3 限流配置
- **文件**: `deployments/apisix/apisix.yaml`
- **操作**: RPM/RPS 限制
- **验收**: 限流测试通过

### T3.4 熔断配置
- **文件**: `deployments/apisix/apisix.yaml`
- **操作**: 后端不可用时友好错误
- **验收**: 熔断测试通过

### T3.5 APISIX CORS 白名单配置
- **文件**: `deployments/apisix/apisix.yaml`, `deployments/*/.env*`
- **操作**: 为前端 SPA 配置按环境区分的 Origin 白名单、允许方法/头与凭据策略
- **验收**: 浏览器跨域请求正常，通过非法 Origin 拒绝测试

### T3.6 LiteLLM 网关接入配置
- **文件**: `backend/app/llm-agent/configs/config.yaml`, `deployments/docker-compose.dev.yml`, `deployments/k8s/`
- **操作**: 配置 llm-agent 到 LiteLLM 的连接地址、超时、重试与密钥注入策略
- **验收**: llm-agent 在降级/超时场景可返回标准化错误并可观测

---

## Phase 4: 前端集成

### 4.0 前端架构升级

#### T4.0.1 引入 TanStack Query
- **文件**: `frontend/package.json`, `frontend/src/main.tsx`
- **操作**: 安装 `@tanstack/react-query`，配置 `QueryClientProvider`
- **验收**: QueryClient 正常初始化

#### T4.0.2 定义 Query Keys 和 API Hooks
- **文件**: `frontend/src/services/queries.ts`
- **操作**: 定义 queryKeys 常量和 useCharacters/useConversations/useMessages 等 hooks
- **验收**: hooks 可正常调用 mock API

#### T4.0.3 重构 ChatContext
- **文件**: `frontend/src/context/ChatContext.tsx`
- **操作**: 移除消息列表管理，改为 TanStack Query + WebSocket 事件驱动
- **验收**: ChatPage 正常工作

#### T4.0.4 全局错误处理与通知组件
- **文件**: `frontend/src/components/GlobalErrorBoundary.tsx`, `frontend/src/components/Toast.tsx`
- **操作**: 建立统一错误捕获、接口错误映射与用户提示组件
- **验收**: 关键异常有统一展示与可恢复提示

### 4.1 API 切换与认证

#### T4.1.1 API 切换机制
- **文件**: `frontend/src/services/api.ts`
- **操作**: VITE_API_MODE 控制 mock/real，统一 ApiClient 封装
- **测试**: 环境变量切换后请求路径正确
- **验收**: 环境变量切换正常

#### T4.1.2 JWT 拦截器
- **文件**: `frontend/src/services/api.ts`
- **操作**: 自动附加 Token，401 自动登出，token refresh
- **测试**: 过期 token 自动登出、refresh 续期
- **验收**: 认证流程正常

#### T4.1.3 AuthContext 对接真实 API
- **文件**: `frontend/src/context/AuthContext.tsx`
- **操作**: 对接 user-service 登录/注册/登出 API
- **验收**: 真实注册→登录→获取用户信息

#### T4.1.4 角色列表/会话列表对接真实 API
- **文件**: `frontend/src/pages/CharacterListPage.tsx`, `frontend/src/pages/ChatPage.tsx`, `frontend/src/services/queries.ts`
- **操作**: 将角色列表与会话列表从 mock 数据切换到真实接口，并补齐加载/空态/错误态
- **依赖**: T4.0.2, T4.1.1
- **验收**: 角色与会话列表在 real 模式下可稳定加载

### 4.2 实时通信

#### T4.2.1 WebSocket 客户端
- **文件**: `frontend/src/services/websocket.ts`
- **操作**:
  - useChatSocket hook
  - JWT 鉴权（连接参数）
  - 心跳检测
  - 指数退避重连（1s, 2s, 4s, 8s, max 30s）
  - 消息事件 → queryClient.invalidateQueries
- **测试**: 连接状态管理、断线重连
- **验收**: 多端同步正常

#### T4.2.2 连接状态指示器
- **文件**: `frontend/src/components/ConnectionIndicator.tsx`
- **操作**: 展示 WebSocket 连接状态（已连接/连接中/已断开/重连中）
- **验收**: 状态实时反映

### 4.3 对话体验

#### T4.3.1 流式渲染（打字机效果）
- **文件**: `frontend/src/components/StreamingMessage.tsx`
- **操作**: SSE/WebSocket 流式文本追加 + Markdown 渲染
- **测试**: chunk 拼接正确性、特殊字符处理
- **验收**: 流式显示正常

#### T4.3.2 ChatLoadingState（对话加载状态）
- **文件**: `frontend/src/components/ChatLoadingState.tsx`
- **操作**:
  - 连接建立前: "正在连接..."
  - 等待 AI 响应: "AI 正在输入..." (含动画)
  - 超过 10s: "AI 正在努力思考，请稍候..."
  - 超过 30s: 超时提示 + 重试按钮
- **来源**: spec-kit FR-012a/FR-012b，适配流式模式
- **验收**: 各阶段提示正确显示

#### T4.3.3 client_message_id 客户端生成
- **文件**: `frontend/src/lib/messageId.ts`
- **操作**: 生成唯一 client_message_id（UUID v7 或 `{userId}_{timestamp}_{random}`）
- **测试**: 唯一性验证
- **验收**: 每条消息携带唯一 client_message_id

### 4.4 记忆管理 UI（基础版）

#### T4.4.1 [P1] MemoryList 组件
- **文件**: `frontend/src/components/MemoryList.tsx`
- **操作**:
  1. Iteration 1：完成 UI 骨架与 mock 适配器（`VITE_API_MODE=mock`）
  2. Iteration 2 API 冻结后切换到 `MemoryList` 真接口（`T2.4M.7`）
- **依赖**: mock 阶段无后端依赖；真实接口阶段依赖 T2.4M.7
- **来源**: spec-kit FR-024（已前置到本版本）
- **验收**: mock/real 两种模式下列表均可正常展示

#### T4.4.1a [P1] MemoryEditDialog 基础编辑
- **文件**: `frontend/src/components/MemoryEditDialog.tsx`
- **操作**: 支持可编辑记忆的修改/删除提交，调用 MemoryUpdate/MemoryDelete API
- **依赖**: T4.4.1, T2.4M.7
- **测试**: 表单校验、提交成功/失败回显、删除二次确认
- **验收**: 用户可完成记忆基础编辑与删除

#### T4.4.1b [P1] Memory 审计事件展示（轻量）
- **文件**: `frontend/src/components/MemoryAuditPanel.tsx`
- **操作**: 展示最近记忆相关安全事件（越权检索拒绝、冲突澄清、同步失败）
- **依赖**: T4.4.1, T2.4M.4a
- **测试**: 空状态、事件映射、分页加载
- **验收**: 用户可查看最近记忆安全事件摘要

#### T4.4.2 积分余额显示
- **文件**: `frontend/src/components/CreditBalance.tsx`
- **操作**: 对话界面显示当前积分余额，余额为 0 时阻止发送并提示充值
- **来源**: spec-kit FR-037
- **验收**: 余额实时更新，0 积分时显示充值引导

---

## Phase 5: 测试与部署

### 5.1 单元测试覆盖

#### T5.1.1 Go 服务单元测试
- **文件**: 各服务 `*_test.go`
- **操作**: 补全各 Phase 2 任务的单元测试，达到 ≥70% 覆盖率
- **验收**: `go test -cover ./...` 各服务 ≥70%

#### T5.1.2 llm-agent 服务单元测试（Go）
- **文件**: `backend/app/llm-agent/**/*_test.go`
- **操作**: prompt_builder、streaming、cost、service 层测试
- **验收**: `go test -cover ./backend/app/llm-agent/...` ≥70%

#### T5.1.3 前端单元测试
- **文件**: `frontend/src/**/*.test.tsx`
- **操作**: Vitest + @testing-library/react，覆盖关键组件
- **工具**: Vitest (已在 spec-kit Phase 1 配置)
- **验收**: `npm run test` 通过

### 5.2 集成测试

#### T5.2.1 用户注册→登录→JWT 全链路
- **文件**: `tests/integration/auth_test.go`
- **操作**: 注册 → 登录 → 获取 JWT → 调用受保护 API
- **验收**: 全链路通过

#### T5.2.2 发送消息→扣费→回复 全链路
- **文件**: `tests/integration/conversation_test.go`
- **操作**: 发送消息 → 预扣 → LLM 调用 → 结算 → 消息存储
- **验收**: 端到端流程通过

#### T5.2.3 SendMessage Outbox 可靠投递链路
- **文件**: `tests/integration/outbox_delivery_test.go`
- **操作**: 验证消息落库后 outbox 事件最终投递到 `evt.memory.write.requested`，并覆盖 relay 重启恢复场景
- **验收**: 无“主链路成功但记忆事件丢失”窗口

#### T5.2.4 WebSocket 降级轮询与同步口径
- **文件**: `frontend/tests/e2e/realtime-fallback.spec.ts`
- **操作**: 模拟 WebSocket 建连失败，验证自动降级 3-5 秒轮询，且跨设备同步延迟口径 ≤5 秒（95%）
- **验收**: 实时链路在 WS 不可用时仍可用

#### T5.2.5 账本 SoT 防直写校验
- **文件**: `tests/integration/ledger_boundary_test.go`
- **操作**: 校验非 billing-service 路径无法直接写入 `credit_accounts/credit_transactions`（仓储封禁 + 集成验证）
- **验收**: 积分余额变更仅通过 billing gRPC 发生

#### T5.2.6 画像字段分级与冲突策略集成测试
- **文件**: `tests/integration/profile_tiering_conflict_test.go`
- **操作**: 覆盖 Tier A（澄清后更新）、Tier B（共存排序）、职业时态（current+history）三类场景
- **验收**: 字段分级规则与冲突策略端到端生效

#### T5.2.7 MCP 网关插件路由与画像工具边界测试
- **文件**: `tests/integration/mcp_gateway_profile_boundary_test.go`
- **操作**: 在当前范围（`profile.*`）校验 `tools.list` scope 过滤、插件路由分发、agent 通过 MCP tools 访问画像、越权拒绝、绕过 memory-service 的写路径被阻断、未注册工具默认拒绝
- **验收**: MCP 接入层（profile 域）与 SoT 边界满足设计约束

#### T5.2.8 充值下单与回调入账集成测试
- **文件**: `tests/integration/recharge_flow_test.go`
- **操作**: 覆盖套餐查询→下单→回调→入账全链路，并验证重复回调幂等
- **验收**: 充值链路余额与流水一致，重复回调不重复入账

### 5.3 PBT 测试

#### T5.3.1 积分系统 PBT
- **文件**: `tests/pbt/billing_test.go`
- **操作**:
  - 余额非负不变量
  - 扣费幂等不变量
  - 预扣结算一致不变量
- **工具**: rapid (Go PBT 库)
- **验收**: 属性测试通过 (≥1000 次)

#### T5.3.2 消息系统 PBT
- **文件**: `tests/pbt/message_test.go`
- **操作**: 分页查询不丢不重不变量
- **验收**: 属性测试通过

### 5.4 前端 E2E 测试

#### T5.4.1 核心用户流程 E2E
- **文件**: `frontend/tests/e2e/`
- **操作**: Playwright 自动化测试
  - 注册 → 登录 → 选择角色 → 发送消息 → 收到回复
  - 积分余额变化
- **验收**: E2E 测试通过

### 5.5 部署

#### T5.5.1 Docker Compose 服务编排补齐
- **文件**: `deployments/docker-compose.dev.yml`
- **操作**: 补齐并编排 `user/character/conversation/billing/llm-agent/memory/mcp/litellm/postgres/redis/apisix` 服务定义
- **验收**: `docker-compose up` 能拉起完整服务拓扑

#### T5.5.2 Docker Compose 健康检查验收
- **文件**: `deployments/docker-compose.dev.yml`, `scripts/healthcheck/`
- **操作**: 统一健康检查路径与启动顺序，验证服务依赖探针（含 LiteLLM 与 mcp/memory/llm-agent）
- **验收**: 所有容器 healthy，关键调用链可通

#### T5.5.3 K8s Helm 部署
- **文件**: `deployments/k8s/`
- **操作**: 探针通过，滚动升级，覆盖 `mcp/memory/llm-agent/litellm` 新增服务 chart 或 values
- **验收**: Helm install 成功

---

## Dependency Graph

### 前置需求优先级（Pulled-in Deferred）

| 优先级 | 目标 | 任务 | 依赖 |
|-------|------|------|------|
| P0 | 情感状态基础版 | T2.4M.1, T2.4M.2, T2.4M.2a, T2.4M.3 | T2.4M.0, T2.4M.0a, T2.6.3, T2.6.4, T1.5 |
| P0 | 安全审计最小闭环 | T2.4M.2b, T2.4M.4, T2.4M.4a, T2.4M.7 | T2.4M.1, T2.4M.2a, T2.6.4 |
| P0 | 画像字段分级与 MCP 接入层 | T2.4M.8, T2.4M.9, T2.4M.12, T2.4M.13 | T2.4M.1, T2.4M.2a, T2.4M.7 |
| P0 | 资料双 SoT 边界落地 | T2.1.7, T2.4M.8, T2.4M.9 | T2.1.1, T2.4M.2a |
| P1 | 记忆管理基础编辑 | T4.4.1, T4.4.1a, T4.4.1b | T2.4M.7, T2.4M.4a |

```
Phase 0 (P0 修复)
    │
    ▼
Phase 1 (基础设施)
    │
    ▼
Phase 2.6 (迁移脚本补充，前置执行)
    │
    ├──────────────────────────────────────┐
    ▼                                      ▼
Phase 2.1 (user)                    Phase 2.3 (character)
    │                                      │
    ▼                                      │
Phase 2.2 (billing) ◄─── T2.1.6 ──────────┤
    │                                      │
    ├──────────────────────────────────────┤
    ▼                                      ▼
Phase 2.4 (llm-agent) ◄────────────────────┘
    │
    ├── Phase 2.4M (memory-core)
    │      ├── P0-A: emotional baseline (T2.4M.0/0a/1/2/2a/3)
    │      ├── P0-B: audit loop (T2.4M.2b/4/4a/7)
    │      └── P1: frontend memory edit readiness
    ▼
Phase 2.5 (conversation) ◄── T2.5.4 依赖 billing + llm-agent + memory-core
    │
    ├── T2.5.6（SendMessage）与 T2.5.7（Relay）可并行；
    │   Relay 依赖事件契约冻结，不阻塞主链路编码
    │
    ▼
Phase 3 (网关集成)
    │
    ▼
Phase 4.0 (前端架构升级) ─→ Phase 4.1-4.4 (前端功能)
                               └── P1: T4.4.1/T4.4.1a/T4.4.1b
    │
    ▼
Phase 5 (测试部署)
```

---

## Iteration Batches (P0/P1)

### 三批执行清单（用于排期与责任分配）

| 迭代 | 目标 | 关键任务 | 主要依赖 | 建议负责人 | 退出标准 |
|------|------|----------|----------|------------|----------|
| Iteration 1 (P0-A) | 基础骨架 + 迁移前置 + 情感基线 | T0.5, T0.6, T1.5, T1.6, T2.6.3, T2.6.4, T2.6.5, T2.4M.0, T2.4M.0a, T2.4M.1, T2.4M.2, T2.4M.2a, T2.4M.3 | 迁移先行（009/010/011） | Go 后端（memory + llm-agent） | memory-service 骨架与 proto 冻结，情感写入与冲突检测链路可跑通 |
| Iteration 2 (P0-B) | SoT 边界 + 账本充值 + MCP 闭环 | T2.1.7, T2.2.8, T2.2.9, T2.2.10, T2.4M.2b, T2.4M.4, T2.4M.4a, T2.4M.7, T2.4M.8, T2.4M.9, T2.4M.12, T2.4M.13, T2.5.4, T2.5.6, T2.5.7, T2.5.8, T3.5, T3.6 | Iteration 1 完成 | Go 后端（billing/memory/conversation/mcp） | 越权/冲突/死信可追溯，充值链路闭环，网关与 LiteLLM 可观测 |
| Iteration 3 (P1) | 前端真实接口完善与部署验收 | T4.0.4, T4.1.4, T4.4.1, T4.4.1a, T4.4.1b, T5.2.8, T5.5.1, T5.5.2, T5.5.3 | Iteration 2 API 稳定 | 前端 + QA + DevOps | 用户可查看编辑记忆、充值流程可验、完整部署拓扑可通过健康检查 |

### 建议并行分工

- **Backend Go Track**: T2.6.3/2.6.4/2.6.5 → T2.4M.0/0a/1/2/2a/2b/4/4a/7/8/9/10/11/12/13 → T2.5.4/5.6/5.7/5.8 → T2.2.8/2.2.9/2.2.10
- **Frontend Track**: Iteration 1 完成 T4.4.1 mock 骨架，Iteration 2 API 冻结后完成 T4.1.4 + T4.4.1 real + T4.4.1a/4.4.1b + T4.0.4
- **QA Track**: 每个迭代末完成对应 Gate，避免把记忆链路问题后移到 E2E 阶段

---

## Task Statistics

| Phase | 任务数 | 说明 |
|-------|--------|------|
| Phase 0 | 6 | P0 风险修复（含 Python llm-agent 退役治理） |
| Phase 1 | 6 | 基础设施（含 Redis Stream group + LiteLLM 基线） |
| Phase 2 | 58 | 核心服务 (user 7 + billing 12 + character 3 + llm-agent 5 + memory-core 18 + conversation 8 + migration 5) |
| Phase 3 | 6 | 网关集成（含 CORS + LiteLLM 接入） |
| Phase 4 | 17 | 前端集成 (架构升级 4 + API 4 + 实时通信 2 + 对话 3 + 记忆/积分 4) |
| Phase 5 | 17 | 测试部署（含充值集成与拆分部署验收） |
| **合计** | **110** | |

---

## Acceptance Checklist

### 构建验收
- [ ] `make setup` 编译通过
- [ ] 所有 proto 代码生成完整（含 billing/v1 + memory/v1 + mcp/v1）
- [ ] llm-agent（Go）服务编译与启动通过
- [ ] Python llm-agent 构建入口已从 CI/部署链路剔除
- [ ] billing-service 独立构建通过
- [ ] memory-service 独立构建通过

### 部署验收
- [ ] docker-compose 所有容器 healthy（含 mcp-service + memory-service + llm-agent + LiteLLM）
- [ ] APISIX upstream active check 通过
- [ ] APISIX CORS 白名单配置通过浏览器跨域验收
- [ ] 迁移脚本可重复执行（IF NOT EXISTS 幂等）

### 功能验收
- [ ] 用户注册→登录→获取 JWT
- [ ] 注册后 billing 赠送 100 积分
- [ ] 用户手动编辑 `user_profiles` 可触发 `USER_STATED` 画像 claim，同步到决议链路
- [ ] 选择角色→发送消息(client_message_id)→流式回复
- [ ] 积分预扣→LLM 调用→结算（多退少补）
- [ ] 预扣超时 5 分钟自动释放
- [ ] 相同 client_message_id 重复请求幂等
- [ ] `message_dedup_keys` 作为全局幂等真源生效（messages 分区表无全局唯一约束依赖）
- [ ] 余额为 0 时拒绝发送并提示充值
- [ ] 充值链路闭环可用（套餐查询→下单→支付回调入账，重复回调不重复加币）
- [ ] 多设备 WebSocket 同步
- [ ] WebSocket 不可用时自动降级 3-5 秒轮询，同步延迟口径仍满足 ≤5 秒（95%）
- [ ] 前端流式渲染（打字机效果）
- [ ] FR-012 流式口径映射通过（TTFT ≤ 3s + 10s/30s 分级提示 + 超时重试）
- [ ] 记忆列表展示
- [ ] 记忆基础编辑/删除可用
- [ ] 记忆安全审计最小闭环可查询
- [ ] 五层长期记忆写入链路生效（Layer1-5）
- [ ] 画像字段分级策略生效（Tier A 澄清门禁 / Tier B 共存排序 / Tier C 弱信号）
- [ ] 职业时态投影生效（current_occupation + history）
- [ ] SendMessage 主链路不因记忆异步写入失败而中断
- [ ] Outbox 可靠投递生效（消息已落库场景下 memory 事件最终可达）
- [ ] Outbox `max_retries` 与 `DEAD` 标记策略生效
- [ ] 矛盾信息触发人设化澄清流程
- [ ] claim 冲突状态机流转可追溯（含死信与重放）
- [ ] Owner-Private / Public / Visitor-Private 可见性隔离验证通过
- [ ] `visibility_filter/taboo_filter` 在 re-rank 前硬过滤生效
- [ ] 记忆压缩与遗忘维护任务按策略执行
- [ ] 非 billing-service 路径无法直写 `credit_accounts/credit_transactions`
- [ ] Agent 画像 CRUD 统一经 `mcp-service` 提供的 profile MCP tools
- [ ] `mcp-service` 不绕过 memory-service 直写数据库
- [ ] `mcp-service` 工具注册中心生效（`tools.list` 按 scope 过滤、未注册工具默认拒绝）

### 前置需求优先级 Gate
- [ ] P0-情感状态基础版完成（T2.4M.0/0a/1/2/2a/3）
- [ ] P0-安全审计最小闭环完成（T2.4M.2b/4/4a/7）
- [ ] P1-记忆基础编辑与审计展示完成（T4.4.1/4.4.1a/4.4.1b）

### PBT 验收
- [ ] 积分余额非负测试通过
- [ ] 扣费幂等测试通过
- [ ] 预扣结算一致测试通过
- [ ] 消息单调一致测试通过

### 测试覆盖
- [ ] Go 各服务单元测试 ≥70%
- [ ] llm-agent（Go）单元测试 ≥70%
- [ ] 前端关键组件测试通过
- [ ] E2E 核心流程测试通过


---

## Deferred Technical Backlog (Out of Current Scope)

以下项已明确延期，不计入当前 `110` 个任务与验收 Gate：

- 记忆高级评分参数：`memory_strength/boost_history/surprise_score/connectivity`（目标 v1.2）
- 事件高级结构：`title/is_recurring/participants/related_memory_ids/source_message_ids[]`（目标 v1.2）
- 图谱高级关系解释与热度：`relation_description/mention_count/last_mentioned_at`（目标 v1.2）
- 情感高级演化字段：`secondary_emotion` 与触发因果解释字段（目标 v1.2）

详见 `openspec/changes/ai-companion-platform/proposal.md` 的 Deferred Features 与 `openspec/changes/ai-companion-platform/data-model.md` 的 Deferred Schema Backlog。
