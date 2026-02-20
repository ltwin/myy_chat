-- Migration: 005_create_messages_table
-- Description: 创建消息表(按月分区)
-- Dependencies: 004_create_conversations_table.sql

-- ================================================
-- 消息表 (messages)
-- 按月分区,用于存储对话消息记录
-- ================================================

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT NOT NULL,  -- 雪花ID
    conversation_id BIGINT NOT NULL,  -- 关联conversations.id
    role VARCHAR(20) NOT NULL,  -- user, assistant, system
    content TEXT NOT NULL,
    token_count INT NOT NULL DEFAULT 0,
    metadata JSONB,  -- {"latency_ms": 2500, "model": "gpt-4o", "cost": 0.05}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id, created_at),  -- 分区表主键必须包含分区键
    CONSTRAINT check_role CHECK (role IN ('user', 'assistant', 'system'))
) PARTITION BY RANGE (created_at);

-- 创建2025年分区表
SELECT create_monthly_partition('messages', 2025, 1);
SELECT create_monthly_partition('messages', 2025, 2);
SELECT create_monthly_partition('messages', 2025, 3);
SELECT create_monthly_partition('messages', 2025, 4);
SELECT create_monthly_partition('messages', 2025, 5);
SELECT create_monthly_partition('messages', 2025, 6);
SELECT create_monthly_partition('messages', 2025, 7);
SELECT create_monthly_partition('messages', 2025, 8);
SELECT create_monthly_partition('messages', 2025, 9);
SELECT create_monthly_partition('messages', 2025, 10);
SELECT create_monthly_partition('messages', 2025, 11);
SELECT create_monthly_partition('messages', 2025, 12);

-- 创建2026年分区表(预创建)
SELECT create_monthly_partition('messages', 2026, 1);
SELECT create_monthly_partition('messages', 2026, 2);
SELECT create_monthly_partition('messages', 2026, 3);
SELECT create_monthly_partition('messages', 2026, 4);
SELECT create_monthly_partition('messages', 2026, 5);
SELECT create_monthly_partition('messages', 2026, 6);

-- 索引 (会自动创建在每个分区上)
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at DESC);

-- 表注释
COMMENT ON TABLE messages IS '消息表 - 对话消息记录(按月分区)';
COMMENT ON COLUMN messages.id IS '雪花ID主键';
COMMENT ON COLUMN messages.conversation_id IS '关联conversations.id(应用层维护)';
COMMENT ON COLUMN messages.role IS '消息角色: user/assistant/system';
COMMENT ON COLUMN messages.content IS '消息内容';
COMMENT ON COLUMN messages.token_count IS 'Token数量';
COMMENT ON COLUMN messages.metadata IS '元数据 JSON {"latency_ms": ..., "model": ..., "cost": ...}';
COMMENT ON COLUMN messages.created_at IS '创建时间(分区键)';
