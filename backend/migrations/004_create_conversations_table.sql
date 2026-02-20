-- Migration: 004_create_conversations_table
-- Description: 创建会话表(按月分区)
-- Dependencies: 001_create_users.sql, 002_create_characters.sql

-- ================================================
-- 会话表 (conversations)
-- 按月分区,用于存储用户与AI角色的对话会话
-- ================================================

CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT NOT NULL,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 关联characters.id
    title VARCHAR(255),
    message_count INT DEFAULT 0,
    token_count INT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_message_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_archived BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (id, started_at)  -- 分区表主键必须包含分区键
) PARTITION BY RANGE (started_at);

-- 创建2025年分区表
SELECT create_monthly_partition('conversations', 2025, 1);
SELECT create_monthly_partition('conversations', 2025, 2);
SELECT create_monthly_partition('conversations', 2025, 3);
SELECT create_monthly_partition('conversations', 2025, 4);
SELECT create_monthly_partition('conversations', 2025, 5);
SELECT create_monthly_partition('conversations', 2025, 6);
SELECT create_monthly_partition('conversations', 2025, 7);
SELECT create_monthly_partition('conversations', 2025, 8);
SELECT create_monthly_partition('conversations', 2025, 9);
SELECT create_monthly_partition('conversations', 2025, 10);
SELECT create_monthly_partition('conversations', 2025, 11);
SELECT create_monthly_partition('conversations', 2025, 12);

-- 创建2026年分区表(预创建)
SELECT create_monthly_partition('conversations', 2026, 1);
SELECT create_monthly_partition('conversations', 2026, 2);
SELECT create_monthly_partition('conversations', 2026, 3);
SELECT create_monthly_partition('conversations', 2026, 4);
SELECT create_monthly_partition('conversations', 2026, 5);
SELECT create_monthly_partition('conversations', 2026, 6);

-- 索引 (会自动创建在每个分区上)
CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_character_id ON conversations(character_id);
CREATE INDEX IF NOT EXISTS idx_conversations_last_message ON conversations(last_message_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversations_user_character ON conversations(user_id, character_id);

-- 表注释
COMMENT ON TABLE conversations IS '会话表 - 用户与AI角色的对话会话(按月分区)';
COMMENT ON COLUMN conversations.id IS '雪花ID主键';
COMMENT ON COLUMN conversations.user_id IS '关联users.id(应用层维护)';
COMMENT ON COLUMN conversations.character_id IS '关联characters.id(应用层维护)';
COMMENT ON COLUMN conversations.title IS '会话标题(可选,自动生成或用户设置)';
COMMENT ON COLUMN conversations.message_count IS '消息数量(由触发器或应用层更新)';
COMMENT ON COLUMN conversations.token_count IS '总Token消耗';
COMMENT ON COLUMN conversations.started_at IS '会话开始时间(分区键)';
COMMENT ON COLUMN conversations.last_message_at IS '最后消息时间';
COMMENT ON COLUMN conversations.is_archived IS '是否归档';
