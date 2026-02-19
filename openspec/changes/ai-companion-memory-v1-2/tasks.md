## 1. 方案冻结与前置决策

- [ ] 1.1 确认 `boost_history` 存储形态（JSONB 或独立表）并在设计文档落锤
- [ ] 1.2 确认 `participants` 字段结构（纯 ID 或 ID+type）并固定校验规则
- [ ] 1.3 确认 `trigger_content` 脱敏策略与敏感词处理边界
- [ ] 1.4 定义 v1.2 feature flag 清单（advanced_scoring、event_rich_mode、graph_edge_heat、emotion_cause）

## 2. 数据库迁移与回滚预案

- [ ] 2.1 新增 `011_memory_advanced_scoring.sql` 扩展 `memories` 高级评分字段与索引
- [ ] 2.2 新增 `012_important_events_rich_fields.sql` 扩展事件结构字段
- [ ] 2.3 新增 `013_graph_edge_heat_fields.sql` 扩展图谱边解释/热度字段
- [ ] 2.4 新增 `014_emotional_state_evolution.sql` 扩展情感次级与触发字段
- [ ] 2.5 为新增字段补充默认值与约束，确保迁移期间旧逻辑可运行
- [ ] 2.6 编写迁移回滚策略文档（优先关开关，避免紧急 DDL 回退）

## 3. Memory Domain 模型扩展

- [ ] 3.1 更新 `memories` 领域对象与 repository 映射高级评分字段
- [ ] 3.2 更新 `important_events` 领域对象与 repository 映射富结构字段
- [ ] 3.3 更新 `memory_graph_edges` 领域对象与 repository 映射热度字段
- [ ] 3.4 更新 `emotional_states` 领域对象与 repository 映射次级情感与触发字段
- [ ] 3.5 补齐新增字段的 DTO 转换与空值兼容逻辑

## 4. 记忆评分与检索策略

- [ ] 4.1 实现组合评分计算器（base + decay + reinforcement + novelty + topology）
- [ ] 4.2 在检索排序管线接入组合评分并保留 Hard Gate 前置
- [ ] 4.3 实现 v1.0 回退逻辑（高级字段缺失时自动降级）
- [ ] 4.4 接入 `advanced_scoring` feature flag，实现灰度可控切换

## 5. 事件与图谱增强逻辑

- [ ] 5.1 实现事件提取写入 `title/is_recurring/participants` 字段
- [ ] 5.2 实现多来源追溯写入 `source_message_ids` 与 `related_memory_ids`
- [ ] 5.3 实现周期事件 upsert 语义，避免重复插入
- [ ] 5.4 实现图谱边 `mention_count/last_mentioned_at` 更新逻辑
- [ ] 5.5 实现图谱边热度衰减维护作业与调度参数

## 6. 情感演化扩展

- [ ] 6.1 实现 `secondary_emotion` 提取与写入链路
- [ ] 6.2 实现 `trigger_type/trigger_content` 提取与脱敏存储
- [ ] 6.3 为情感记录新增一致性约束校验（主次情感与 PAD 信号）
- [ ] 6.4 接入 `emotion_cause` feature flag，支持分阶段上线

## 7. API / Proto / 前端适配

- [ ] 7.1 评估并扩展 memory proto 返回字段（采用向后兼容追加）
- [ ] 7.2 更新 memory-service API 序列化层，返回新增解释字段
- [ ] 7.3 更新前端记忆管理视图，展示事件标题、参与者与关系说明
- [ ] 7.4 更新前端审计/详情视图，展示触发因果与评分解释信息

## 8. 测试与质量门禁

- [ ] 8.1 为迁移脚本新增幂等执行测试与回滚演练脚本
- [ ] 8.2 为组合评分新增单元测试（含缺字段回退场景）
- [ ] 8.3 为周期事件 upsert 与多来源追溯新增集成测试
- [ ] 8.4 为图谱热度衰减新增作业测试（含边界时间窗口）
- [ ] 8.5 为情感触发脱敏新增安全测试
- [ ] 8.6 补充 PBT 不变量（排序稳定性、热度衰减单调、回退一致性）

## 9. 灰度发布与验收

- [ ] 9.1 在灰度环境启用影子评分，采集对照指标（召回相关性/用户编辑率）
- [ ] 9.2 小流量开启 `advanced_scoring` 并监控回归指标
- [ ] 9.3 分批开启 `event_rich_mode`、`graph_edge_heat`、`emotion_cause`
- [ ] 9.4 完成 v1.2 验收报告并冻结默认开关策略
