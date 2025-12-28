-- Migration: 001_create_users_table
-- Description: 创建用户表
-- Dependencies: init.sql (must run first for update_updated_at_column function)

-- ================================================
-- 用户表 (users)
-- ================================================

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,  -- 雪花ID,由应用层生成
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20) UNIQUE,
    avatar_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE,
    deletion_scheduled_at TIMESTAMP WITH TIME ZONE,  -- 账号删除计划时间(30天冷静期)
    CONSTRAINT check_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$')
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_users_deletion_scheduled ON users(deletion_scheduled_at) WHERE deletion_scheduled_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- 触发器: 自动更新 updated_at
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 表注释
COMMENT ON TABLE users IS '用户表 - 存储用户基本信息';
COMMENT ON COLUMN users.id IS '雪花ID主键,由应用层生成';
COMMENT ON COLUMN users.username IS '用户名,唯一且至少3个字符';
COMMENT ON COLUMN users.email IS '邮箱地址,唯一';
COMMENT ON COLUMN users.password_hash IS '密码哈希(bcrypt)';
COMMENT ON COLUMN users.phone IS '手机号,唯一';
COMMENT ON COLUMN users.avatar_url IS '头像URL';
COMMENT ON COLUMN users.last_login_at IS '最后登录时间';
COMMENT ON COLUMN users.is_deleted IS '软删除标记';
COMMENT ON COLUMN users.deletion_scheduled_at IS '账号删除计划时间,设置后30天自动删除';
