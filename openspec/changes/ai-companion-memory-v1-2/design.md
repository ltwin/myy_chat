## Context

`ai-companion-platform` 在 v1.0 已完成长期记忆基线：

- 五层记忆写入链路（Layer1~5）
- Claim 冲突状态机与人设化澄清
- Recall → Hard Gate → Re-rank 的权限硬过滤
- 时态关系边（`valid_from/valid_until`）与基础可见性隔离

但当前记忆质量仍依赖简化参数（`importance_score/decay_factor/boost_count/access_count`），在“长期使用可解释性”“事件表达能力”“图谱关系热度”与“情感演化深度”上存在上限。v1.2 目标是在不破坏 v1.0 行为的前提下，做可回滚的前向增强。

约束与干系人：

- **约束**：必须保持对现有 API、迁移链路与在线数据兼容；不得影响 v1.0 主链路稳定性。
- **后端干系人**：memory-service、conversation-service、migration 维护者。
- **前端干系人**：记忆管理/审计页面（展示更多解释字段）。
- **质量干系人**：QA/PBT（新增不变量与回归场景）。

## Goals / Non-Goals

**Goals:**
- 在 `memories` 增加高级评分字段，使记忆优先级从“线性规则”升级为“可解释混合评分”。
- 在 `important_events` 增加结构化字段，支持标题化展示、周期事件与多来源追溯。
- 在 `memory_graph_edges` 增加关系解释与热度字段，支持“关系如何变化”的叙事与衰减。
- 在 `emotional_states` 增加次级情感与触发因果字段，支持更细粒度演化分析。
- 保持 v1.0 兼容：所有新增字段默认值安全、旧逻辑可继续运行。

**Non-Goals:**
- 不引入全新的图数据库/向量数据库基础设施。
- 不在本变更内交付后台可视化配置 UI。
- 不重写现有 claim 流程与权限模型（仅扩展其输入信号）。
- 不对历史数据做强制全量回填（采用增量可选回填策略）。

## Decisions

### 1) Schema 采用“加法变更”而非替换

**Decision**
- 所有能力通过 `ADD COLUMN/ADD INDEX` 引入，旧字段保留。
- 新字段要求提供默认值或允许 `NULL`，避免迁移时锁表与逻辑中断。

**Rationale**
- 兼容现有线上数据与调用方；可分批灰度启用新评分逻辑。

**Alternatives considered**
- 直接重构为新表并迁移：一致性更干净，但迁移风险与切换复杂度高，放弃。

### 2) 高级评分采用“分层组合评分”

**Decision**
- 评分模型采用分层组合：
  - Base: `importance_score`
  - Time-decay: `decay_factor`
  - Reinforcement: `memory_strength + boost_history + access_count`
  - Novelty/Topology: `surprise_score + connectivity`
- 检索排序保留 Hard Gate 前置，不允许用评分替代权限过滤。

**Rationale**
- 在保证安全的前提下提升召回质量，并可解释“为什么被召回”。

**Alternatives considered**
- 单一黑盒模型打分：效果潜力高，但难以审计与调参，暂不采用。

### 3) 事件结构采用“核心字段 + 可扩展数组”

**Decision**
- `important_events` 增补：`title`、`is_recurring`、`participants`、`related_memory_ids`、`source_message_ids[]`。
- 结构化字段优先用于查询和展示，复杂扩展继续放在 `metadata`。

**Rationale**
- 兼顾查询效率与演化灵活性，避免过度 JSON 化导致查询困难。

**Alternatives considered**
- 全部放入 JSONB：灵活但索引与约束弱，不利于长期维护。

### 4) 图谱边增强保持“时态语义为主、热度为辅”

**Decision**
- 在 `memory_graph_edges` 增补 `relation_description`、`mention_count`、`last_mentioned_at`。
- 继续以 `valid_from/valid_until` 作为关系演化主语义，热度用于排序与解释。

**Rationale**
- 保持与 v1.0 时态设计一致，同时支持“关系变化依据”的可解释输出。

**Alternatives considered**
- 仅做热度不做描述字段：计算简单，但解释能力不足。

### 5) 情感演化采用“主信号 + 次信号 + 触发因果”

**Decision**
- 在 `emotional_states` 增补 `secondary_emotion`、`trigger_type`、`trigger_content`（可脱敏）。
- 延续 PAD 主信号校验约束，次级信号可选。

**Rationale**
- 让情感记录从“状态快照”升级为“状态 + 触发原因”，提高分析可用性。

**Alternatives considered**
- 仅在 metadata 里记录触发信息：实现快，但语义不可控、统计复杂。

### 6) Proto 演进采用“向后兼容扩展”

**Decision**
- 优先在 reply/message 尾部追加字段（protobuf 向后兼容规则），不复用已有 tag。
- 旧客户端忽略新字段，服务端按能力开关回填。

**Rationale**
- 控制跨端升级成本，避免同步升级风险。

**Alternatives considered**
- 新建 v2 proto 包：隔离清晰，但系统改造面过大，暂不采用。

## Risks / Trade-offs

- **[风险] 高级评分参数引入后排序抖动** → **缓解**：灰度开关 + 双写观测 + A/B 对照。
- **[风险] 新增字段导致迁移耗时增加** → **缓解**：分批迁移、先加列后加索引、离峰执行。
- **[风险] 事件/图谱字段增加导致写入放大** → **缓解**：默认惰性填充，异步补全非关键字段。
- **[风险] 情感触发内容涉及隐私** → **缓解**：`trigger_content` 支持脱敏或摘要化存储。
- **[风险] 前端一次性接收过多新字段** → **缓解**：API 兼容模式与渐进展示。

## Migration Plan

1. **Schema 扩展**
   - 新增迁移（建议 `011_*`）：`memories` 高级评分字段与索引。
   - 新增迁移（建议 `012_*`）：`important_events` 结构化字段。
   - 新增迁移（建议 `013_*`）：`memory_graph_edges` 热度/解释字段。
   - 新增迁移（建议 `014_*`）：`emotional_states` 次级情感与触发字段。
2. **服务层适配**
   - memory-service 仓储层读取/写入新增字段。
   - 检索层引入新评分组合逻辑（受 feature flag 控制）。
3. **灰度与验证**
   - 先影子计算（不影响线上排序），对比召回与点击/编辑反馈。
   - 再小流量启用新排序，观察回归指标。
4. **回滚策略**
   - 回滚优先关闭 feature flag；
   - 保留新增列但停用读取逻辑，避免紧急 DDL 回滚。

## Open Questions

- `boost_history` 是采用 JSONB（灵活）还是独立表（查询友好）？
- `participants` 应仅存 ID 还是存“ID + type”结构？
- `trigger_content` 的脱敏规则由 memory-service 还是上游统一处理？
- 关系热度衰减作业频率（小时/天）与成本阈值如何设定？
- 新评分上线后的主指标采用“召回相关性”还是“用户编辑率下降”作为第一目标？
