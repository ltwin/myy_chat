-- MyY Chat AI角色对话平台 - 数据库初始化脚本
-- 创建数据库和基础扩展
-- PostgreSQL 16 + pgvector

-- 创建数据库 (如果不存在)
-- SELECT 'CREATE DATABASE myy_chat' WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'myy_chat')\gexec

-- 连接到 myy_chat 数据库后执行以下语句
-- \c myy_chat

-- ==================================================
-- 安装扩展
-- ==================================================

-- pgvector 向量扩展 (用于记忆系统的 Embedding 存储)
CREATE EXTENSION IF NOT EXISTS vector;

-- uuid-ossp 扩展 (备用,主要使用 Snowflake ID)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- pg_trgm 扩展 (用于模糊搜索)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ==================================================
-- 创建 Schema
-- ==================================================

-- 用户相关
CREATE SCHEMA IF NOT EXISTS users;

-- 角色相关
CREATE SCHEMA IF NOT EXISTS characters;

-- 对话相关
CREATE SCHEMA IF NOT EXISTS conversations;

-- 记忆相关
CREATE SCHEMA IF NOT EXISTS memories;

-- 计费相关
CREATE SCHEMA IF NOT EXISTS billing;

-- 系统相关
CREATE SCHEMA IF NOT EXISTS system;

-- ==================================================
-- 创建基础表 (MVP 阶段)
-- ==================================================

-- 用户表 (users.users)
CREATE TABLE IF NOT EXISTS users.users (
    id BIGINT PRIMARY KEY,                    -- Snowflake ID
    email VARCHAR(255) UNIQUE NOT NULL,       -- 邮箱
    username VARCHAR(100),                    -- 用户名
    password_hash VARCHAR(255) NOT NULL,      -- 密码哈希(bcrypt)
    avatar_url TEXT,                          -- 头像URL
    status VARCHAR(20) DEFAULT 'active',      -- 状态: active, suspended, deleted
    deletion_scheduled_at TIMESTAMP,          -- 预定删除时间(30天冷静期)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 用户画像表 (users.profiles)
CREATE TABLE IF NOT EXISTS users.profiles (
    user_id BIGINT PRIMARY KEY,               -- 用户ID (无外键)
    bio TEXT,                                  -- 简介
    preferences JSONB DEFAULT '{}',            -- 用户偏好(JSON)
    metadata JSONB DEFAULT '{}',               -- 其他元数据
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 角色表 (characters.characters)
CREATE TABLE IF NOT EXISTS characters.characters (
    id BIGINT PRIMARY KEY,                     -- Snowflake ID
    name VARCHAR(100) NOT NULL,                -- 角色名称
    avatar_url TEXT,                           -- 头像URL
    personality TEXT NOT NULL,                 -- 性格描述
    background TEXT,                           -- 背景故事
    speaking_style TEXT,                       -- 说话风格
    is_preset BOOLEAN DEFAULT FALSE,           -- 是否预设角色
    creator_id BIGINT,                         -- 创建者ID (用户自定义角色)
    status VARCHAR(20) DEFAULT 'active',       -- 状态
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 对话表 (conversations.conversations)
CREATE TABLE IF NOT EXISTS conversations.conversations (
    id BIGINT PRIMARY KEY,                     -- Snowflake ID
    user_id BIGINT NOT NULL,                   -- 用户ID
    character_id BIGINT NOT NULL,              -- 角色ID
    title VARCHAR(200),                        -- 对话标题
    last_message_at TIMESTAMP,                 -- 最后消息时间
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 消息表 (conversations.messages)
CREATE TABLE IF NOT EXISTS conversations.messages (
    id BIGINT PRIMARY KEY,                     -- Snowflake ID
    conversation_id BIGINT NOT NULL,           -- 对话ID
    role VARCHAR(20) NOT NULL,                 -- 角色: user, assistant
    content TEXT NOT NULL,                     -- 消息内容
    tokens INTEGER DEFAULT 0,                  -- Token数量
    created_at TIMESTAMP DEFAULT NOW()
);

-- 记忆表 (memories.memories)
CREATE TABLE IF NOT EXISTS memories.memories (
    id BIGINT PRIMARY KEY,                     -- Snowflake ID
    user_id BIGINT NOT NULL,                   -- 用户ID
    character_id BIGINT NOT NULL,              -- 角色ID
    content TEXT NOT NULL,                     -- 记忆内容
    type VARCHAR(50) NOT NULL,                 -- 记忆类型: temporary, user_portrait, emotional_state, etc.
    importance DECIMAL(3,2) DEFAULT 0.5,       -- 重要性 (0-1)
    embedding vector(1536),                    -- Embedding向量 (OpenAI ada-002 / BGE-M3)
    expires_at TIMESTAMP,                      -- 过期时间 (临时记忆)
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 关系表 (memories.relationships)
CREATE TABLE IF NOT EXISTS memories.relationships (
    user_id BIGINT NOT NULL,                   -- 用户ID
    character_id BIGINT NOT NULL,              -- 角色ID
    closeness DECIMAL(5,2) DEFAULT 0,          -- 亲密度 (0-100)
    emotional_bond DECIMAL(5,2) DEFAULT 0,     -- 情感纽带 (0-100)
    metadata JSONB DEFAULT '{}',               -- 其他元数据
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, character_id)
);

-- 积分账户表 (billing.credit_accounts)
CREATE TABLE IF NOT EXISTS billing.credit_accounts (
    user_id BIGINT PRIMARY KEY,                -- 用户ID
    balance BIGINT DEFAULT 100,                -- 积分余额 (新用户100积分)
    total_earned BIGINT DEFAULT 100,           -- 累计获得
    total_spent BIGINT DEFAULT 0,              -- 累计消费
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 积分交易记录表 (billing.credit_transactions)
CREATE TABLE IF NOT EXISTS billing.credit_transactions (
    id BIGINT PRIMARY KEY,                     -- Snowflake ID
    user_id BIGINT NOT NULL,                   -- 用户ID
    amount BIGINT NOT NULL,                    -- 金额 (正数=充值,负数=消费)
    type VARCHAR(50) NOT NULL,                 -- 类型: initial, recharge, conversation, tool_use
    description TEXT,                          -- 描述
    idempotency_key VARCHAR(255) UNIQUE,       -- 幂等键
    metadata JSONB DEFAULT '{}',               -- 元数据
    created_at TIMESTAMP DEFAULT NOW()
);

-- ==================================================
-- 创建索引
-- ==================================================

-- 用户表索引
CREATE INDEX IF NOT EXISTS idx_users_email ON users.users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users.users(status);

-- 对话表索引
CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations.conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_conversations_character_id ON conversations.conversations(character_id);
CREATE INDEX IF NOT EXISTS idx_conversations_updated_at ON conversations.conversations(updated_at DESC);

-- 消息表索引
CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON conversations.messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON conversations.messages(created_at DESC);

-- 记忆表索引
CREATE INDEX IF NOT EXISTS idx_memories_user_character ON memories.memories(user_id, character_id);
CREATE INDEX IF NOT EXISTS idx_memories_type ON memories.memories(type);
CREATE INDEX IF NOT EXISTS idx_memories_expires_at ON memories.memories(expires_at) WHERE expires_at IS NOT NULL;

-- pgvector HNSW 索引 (向量检索)
CREATE INDEX IF NOT EXISTS idx_memories_embedding ON memories.memories
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- 积分交易索引
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON billing.credit_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON billing.credit_transactions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency_key ON billing.credit_transactions(idempotency_key);

-- ==================================================
-- 插入预设数据
-- ==================================================

-- 插入5个预设AI角色 (id 使用固定值,便于引用)
INSERT INTO characters.characters (id, name, avatar_url, personality, background, speaking_style, is_preset, status)
VALUES
    (1000000000000000001, '智能助手小艾', 'https://example.com/avatars/xiaoai.png', '友善、耐心、专业', '一位经验丰富的AI助手,擅长解答各类问题', '礼貌、清晰、逻辑性强', TRUE, 'active'),
    (1000000000000000002, '心理导师小心', 'https://example.com/avatars/xiaoxin.png', '温暖、善解人意、同理心强', '专业的心理咨询师,倾听者和陪伴者', '温柔、鼓励、支持性', TRUE, 'active'),
    (1000000000000000003, '编程导师小码', 'https://example.com/avatars/xiaoma.png', '严谨、逻辑清晰、富有经验', '资深软件工程师,精通多种编程语言', '技术性强、条理清晰、善于解释', TRUE, 'active'),
    (1000000000000000004, '创意伙伴小思', 'https://example.com/avatars/xiaosi.png', '活泼、富有想象力、乐观向上', '创意工作者,擅长头脑风暴和创新思维', '生动、幽默、启发性', TRUE, 'active'),
    (1000000000000000005, '学习伙伴小博', 'https://example.com/avatars/xiaobo.png', '博学、严谨、善于教学', '教育工作者,帮助用户学习和成长', '耐心、循序渐进、鼓励式', TRUE, 'active')
ON CONFLICT (id) DO NOTHING;

-- ==================================================
-- 创建函数和触发器
-- ==================================================

-- 自动更新 updated_at 字段的函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为各表添加触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users.users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_profiles_updated_at BEFORE UPDATE ON users.profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters.characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations.conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_memories_updated_at BEFORE UPDATE ON memories.memories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_relationships_updated_at BEFORE UPDATE ON memories.relationships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_credit_accounts_updated_at BEFORE UPDATE ON billing.credit_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ==================================================
-- 权限设置 (可选)
-- ==================================================

-- 授予应用用户权限 (如果需要)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA users TO myy_chat_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA characters TO myy_chat_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA conversations TO myy_chat_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA memories TO myy_chat_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA billing TO myy_chat_user;

-- ==================================================
-- 完成提示
-- ==================================================

DO $$
BEGIN
    RAISE NOTICE '✅ MyY Chat 数据库初始化完成!';
    RAISE NOTICE '📊 已创建 6 个 Schema';
    RAISE NOTICE '📋 已创建 10 个基础表';
    RAISE NOTICE '🔍 已创建 11 个索引 (包括 pgvector HNSW)';
    RAISE NOTICE '🎭 已插入 5 个预设AI角色';
    RAISE NOTICE '⚡ 已创建 8 个 updated_at 触发器';
END $$;
