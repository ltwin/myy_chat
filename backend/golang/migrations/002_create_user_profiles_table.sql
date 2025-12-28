-- Migration: 002_create_user_profiles_table
-- Description: 创建用户画像表
-- Dependencies: 001_create_users_table.sql

-- ================================================
-- 用户画像表 (user_profiles)
-- 存储用户主动填写的静态信息
-- 注意: AI 推断的 preferences/personality 存储在 user_portraits 表中
-- ================================================

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id BIGINT PRIMARY KEY,  -- 关联users.id,由应用层维护
    full_name VARCHAR(100),
    gender VARCHAR(20),  -- male, female, other
    birth_date DATE,
    location JSONB,  -- {"country": "China", "city": "Shanghai"}
    interests TEXT[],  -- 兴趣爱好数组 (用户手动填写)
    occupation VARCHAR(100),
    bio TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_gender CHECK (gender IS NULL OR gender IN ('male', 'female', 'other'))
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_profiles(user_id);

-- 触发器: 自动更新 updated_at
CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 表注释
COMMENT ON TABLE user_profiles IS '用户画像表 - 存储用户个人详细信息';
COMMENT ON COLUMN user_profiles.user_id IS '关联users.id（应用层维护外键关系）';
COMMENT ON COLUMN user_profiles.full_name IS '用户真实姓名';
COMMENT ON COLUMN user_profiles.gender IS '性别: male/female/other';
COMMENT ON COLUMN user_profiles.birth_date IS '出生日期';
COMMENT ON COLUMN user_profiles.location IS '位置信息 JSON {"country": "...", "city": "..."}';
COMMENT ON COLUMN user_profiles.interests IS '兴趣爱好数组(用户手动填写)';
COMMENT ON COLUMN user_profiles.occupation IS '职业';
COMMENT ON COLUMN user_profiles.bio IS '个人简介';
