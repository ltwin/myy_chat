-- Migration: 003_create_characters_table
-- Description: 创建AI角色表
-- Dependencies: 001_create_users_table.sql

-- ================================================
-- AI角色表 (characters)
-- 存储预设角色和用户自定义角色
-- ================================================

CREATE TABLE IF NOT EXISTS characters (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT,  -- 关联users.id, NULL表示预设角色
    name VARCHAR(100) NOT NULL,
    avatar_url TEXT,
    description TEXT,
    personality JSONB NOT NULL,  -- {"mbti": "ENFP", "traits": ["friendly", "curious"]}
    background_story TEXT,
    speaking_style TEXT,
    system_prompt TEXT NOT NULL,
    world_view JSONB,  -- 世界观设定
    is_public BOOLEAN DEFAULT FALSE,
    is_preset BOOLEAN DEFAULT FALSE,
    version INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE,
    CONSTRAINT check_name_length CHECK (char_length(name) >= 2)
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_characters_user_id ON characters(user_id) WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_characters_public ON characters(is_public) WHERE is_public = TRUE AND NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_characters_preset ON characters(is_preset) WHERE is_preset = TRUE;
CREATE INDEX IF NOT EXISTS idx_characters_created_at ON characters(created_at);

-- 触发器: 自动更新 updated_at
CREATE TRIGGER update_characters_updated_at
    BEFORE UPDATE ON characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 表注释
COMMENT ON TABLE characters IS 'AI角色表 - 预设角色和用户自定义角色';
COMMENT ON COLUMN characters.id IS '雪花ID主键';
COMMENT ON COLUMN characters.user_id IS '关联users.id(应用层维护), NULL表示预设角色';
COMMENT ON COLUMN characters.name IS '角色名称,至少2个字符';
COMMENT ON COLUMN characters.avatar_url IS '角色头像URL';
COMMENT ON COLUMN characters.description IS '角色描述';
COMMENT ON COLUMN characters.personality IS '性格设定 JSON {"mbti": "...", "traits": [...]}';
COMMENT ON COLUMN characters.background_story IS '背景故事';
COMMENT ON COLUMN characters.speaking_style IS '说话风格';
COMMENT ON COLUMN characters.system_prompt IS '系统提示词';
COMMENT ON COLUMN characters.world_view IS '世界观设定 JSON';
COMMENT ON COLUMN characters.is_public IS '是否公开(其他用户可见)';
COMMENT ON COLUMN characters.is_preset IS '是否预设角色';
COMMENT ON COLUMN characters.version IS '版本号,用于乐观锁';
COMMENT ON COLUMN characters.is_deleted IS '软删除标记';
