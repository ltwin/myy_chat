## Why

当前 `ai-companion-platform` 已完成长期记忆的可用基线（五层写入、冲突澄清、可见性硬过滤），但多项高价值记忆能力仍被明确延期到 v1.2。为了避免“延期项长期悬置”与后续实现口径漂移，需要单独创建 v1.2 变更，将高级记忆评分、事件可解释性、图谱热度与情感演化能力纳入可执行计划。

## What Changes

- 新增记忆高级评分能力：在现有 `importance_score/decay_factor/boost_count/access_count` 基线之上，引入 `memory_strength`、`boost_history`、`surprise_score`、`connectivity` 等扩展参数。
- 新增事件高级结构能力：为 `important_events` 补充 `title`、`is_recurring`、`participants`、`related_memory_ids`、`source_message_ids[]`，增强可解释性与 UI 展示能力。
- 新增图谱关系增强能力：为图谱边引入 `relation_description`、`mention_count`、`last_mentioned_at`，支持关系解释与热度衰减。
- 新增情感状态高级演化能力：在 PAD 基线外扩展 `secondary_emotion` 与触发因果字段，支持更细粒度情感追踪。
- 明确本次变更与 v1.0 的边界：v1.0 基线字段保持兼容，不做破坏式删除，仅执行前向增量。

## Capabilities

### New Capabilities
- `memory-advanced-scoring`: 扩展长期记忆评分模型，支持强度、强化历史、惊讶度与连接度等高阶信号。
- `memory-event-rich-structure`: 扩展重要事件的数据结构，支持周期事件、参与者与多来源追溯。
- `memory-graph-advanced-edges`: 扩展图谱边的解释与热度字段，支持关系变化叙事与运营分析。
- `emotional-state-evolution`: 扩展情感状态模型，支持次级情感与触发因果解释。

### Modified Capabilities
- （空）

## Impact

- **Data Model / Migration**: 影响 `backend/migrations/` 中记忆域迁移，新增/扩展 `memories`、`important_events`、`emotional_states`、`memory_graph_edges` 字段与索引。
- **Memory Service**: 影响 `backend/app/memory/` 领域模型、仓储层与检索/维护作业（尤其评分与衰减策略）。
- **Proto / API**: 需要评估并可能扩展 `backend/api/memory/v1/memory.proto` 的响应字段，保持向后兼容。
- **Frontend**: 影响记忆管理与审计展示的信息密度与解释能力（事件标题、参与者、关系说明等）。
- **QA / PBT**: 需新增针对高级评分与关系热度衰减的不变量与回归测试。
