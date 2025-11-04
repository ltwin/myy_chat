-- ============================================
-- AI聊天服务器 - PostgreSQL数据库Schema
-- 支持: 分层记忆、多角色、向量检索
-- ============================================

-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";      -- UUID生成
CREATE EXTENSION IF NOT EXISTS "pgvector";       -- 向量检索
CREATE EXTENSION IF NOT EXISTS "pg_trgm";        -- 文本相似度

-- ============================================
-- 1. 用户系统
-- ============================================

-- 用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url TEXT,

    -- 用户画像（JSONB格式，灵活扩展）
    profile JSONB DEFAULT '{}'::jsonb,
    -- 示例: {"interests": ["编程", "AI"], "profession": "工程师", "interaction_style": "友好"}

    -- 元数据
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE
);

-- 用户画像索引
CREATE INDEX idx_users_profile ON users USING GIN(profile);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- ============================================
-- 2. 角色系统（核心设定层）
-- ============================================

-- 角色表
CREATE TABLE characters (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    creator_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- 核心设定
    personality JSONB NOT NULL,
    -- 示例: {"mbti": "ENFP", "traits": ["热情", "创造性"], "speaking_style": "活泼"}

    background_story TEXT,
    core_values JSONB DEFAULT '[]'::jsonb,
    -- 示例: ["诚实", "善良", "勇敢"]

    -- 系统提示词模板
    system_prompt TEXT NOT NULL,
    -- LLM的系统提示词，包含角色设定

    -- 外观设置
    avatar_url TEXT,

    -- 公开性
    is_public BOOLEAN DEFAULT FALSE,
    -- 公开角色可被其他用户使用

    -- 版本控制
    version INTEGER DEFAULT 1,

    -- 元数据
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_active BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_characters_creator ON characters(creator_id);
CREATE INDEX idx_characters_public ON characters(is_public) WHERE is_public = TRUE;

-- 角色版本历史表
CREATE TABLE character_versions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    character_id UUID REFERENCES characters(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,

    -- 变更内容快照
    personality JSONB,
    background_story TEXT,
    system_prompt TEXT,

    -- 变更原因
    change_reason TEXT,
    changed_by UUID REFERENCES users(id) ON DELETE SET NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(character_id, version)
);

-- ============================================
-- 3. 对话会话（临时记忆层）
-- ============================================

-- 会话表
CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    title VARCHAR(255),
    -- 自动生成的会话标题（由LLM总结前几轮对话）

    -- 会话摘要（定期由LLM更新）
    summary TEXT,

    -- 会话元数据
    message_count INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,

    -- 状态
    status VARCHAR(20) DEFAULT 'active',
    -- active, archived, deleted

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_message_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_conversations_user ON conversations(user_id);
CREATE INDEX idx_conversations_character ON conversations(character_id);
CREATE INDEX idx_conversations_status ON conversations(status);

-- 消息表（对话历史）
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,

    role VARCHAR(20) NOT NULL,
    -- 'user', 'assistant', 'system'

    content TEXT NOT NULL,

    -- Token统计
    token_count INTEGER,

    -- 元数据
    metadata JSONB DEFAULT '{}'::jsonb,
    -- 示例: {"model": "gpt-4", "temperature": 0.7, "finish_reason": "stop"}

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id);
CREATE INDEX idx_messages_created_at ON messages(created_at);

-- ============================================
-- 4. 长期记忆层（核心功能）
-- ============================================

-- 长期记忆表（带向量索引）
CREATE TABLE long_term_memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- 记忆类型
    memory_type VARCHAR(50) NOT NULL,
    -- 'user_profile', 'key_event', 'relationship', 'emotional_state'

    -- 记忆内容
    content TEXT NOT NULL,
    -- 原始文本内容

    -- 向量embedding（用于语义检索）
    embedding vector(1536),
    -- OpenAI text-embedding-3-small: 1536维
    -- 可根据使用的embedding模型调整维度

    -- 重要性评分（1-10）
    importance INTEGER CHECK (importance BETWEEN 1 AND 10),
    -- 由LLM评估或规则计算

    -- 关联信息
    related_conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,
    -- 记忆来源的对话

    -- 结构化数据（不同类型的记忆有不同的结构）
    structured_data JSONB DEFAULT '{}'::jsonb,
    -- 示例 (key_event): {"event_type": "decision", "outcome": "positive"}
    -- 示例 (emotional_state): {"emotion": "happy", "intensity": 8}
    -- 示例 (relationship): {"relation_type": "friend", "closeness": 7}

    -- 时间信息
    event_time TIMESTAMP WITH TIME ZONE,
    -- 记忆相关事件发生的时间（可能不同于created_at）

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- 访问统计
    access_count INTEGER DEFAULT 0,
    last_accessed_at TIMESTAMP WITH TIME ZONE
);

-- 向量相似度索引（使用HNSW算法，性能更好）
CREATE INDEX idx_memories_embedding ON long_term_memories
    USING hnsw (embedding vector_cosine_ops);

-- 其他索引
CREATE INDEX idx_memories_user ON long_term_memories(user_id);
CREATE INDEX idx_memories_character ON long_term_memories(character_id);
CREATE INDEX idx_memories_type ON long_term_memories(memory_type);
CREATE INDEX idx_memories_importance ON long_term_memories(importance DESC);
CREATE INDEX idx_memories_event_time ON long_term_memories(event_time DESC);
CREATE INDEX idx_memories_structured ON long_term_memories USING GIN(structured_data);

-- ============================================
-- 5. 人际关系图谱
-- ============================================

-- 关系表（用户-角色关系）
CREATE TABLE user_character_relationships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,

    -- 关系类型
    relation_type VARCHAR(50) DEFAULT 'acquaintance',
    -- 'stranger', 'acquaintance', 'friend', 'close_friend', 'custom'

    -- 亲密度（0-100）
    closeness INTEGER DEFAULT 0 CHECK (closeness BETWEEN 0 AND 100),

    -- 情感连接
    emotional_bond JSONB DEFAULT '{}'::jsonb,
    -- 示例: {"trust": 70, "affection": 60, "respect": 80}

    -- 交互统计
    interaction_count INTEGER DEFAULT 0,
    total_conversation_time INTEGER DEFAULT 0,
    -- 单位：秒

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE(user_id, character_id)
);

CREATE INDEX idx_relationships_user ON user_character_relationships(user_id);
CREATE INDEX idx_relationships_closeness ON user_character_relationships(closeness DESC);

-- ============================================
-- 6. 情感状态追踪（时间序列）
-- ============================================

-- 情感状态历史表
CREATE TABLE emotional_states (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES characters(id) ON DELETE CASCADE,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,

    -- 情感维度（可扩展）
    emotions JSONB NOT NULL,
    -- 示例: {"joy": 0.8, "sadness": 0.1, "anger": 0.0, "surprise": 0.3}
    -- 值范围: 0.0-1.0

    -- 整体情感倾向
    overall_sentiment VARCHAR(20),
    -- 'very_positive', 'positive', 'neutral', 'negative', 'very_negative'

    sentiment_score DECIMAL(3, 2),
    -- -1.0 到 1.0

    -- 触发原因
    trigger_event TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_emotional_states_user ON emotional_states(user_id);
CREATE INDEX idx_emotional_states_time ON emotional_states(created_at DESC);
CREATE INDEX idx_emotional_states_conversation ON emotional_states(conversation_id);

-- ============================================
-- 7. 上下文压缩历史
-- ============================================

-- 压缩记录表
CREATE TABLE context_compressions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,

    -- 压缩前的消息ID范围
    start_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    end_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,

    -- 压缩前后的Token数
    original_tokens INTEGER NOT NULL,
    compressed_tokens INTEGER NOT NULL,
    compression_ratio DECIMAL(4, 2),
    -- 压缩率 = compressed/original

    -- 压缩摘要
    summary TEXT NOT NULL,

    -- 提取的关键信息
    key_points JSONB DEFAULT '[]'::jsonb,
    -- 示例: [{"point": "用户询问了Python问题", "importance": 8}]

    -- 压缩使用的模型
    model_used VARCHAR(100),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_compressions_conversation ON context_compressions(conversation_id);

-- ============================================
-- 8. API调用统计（成本控制）
-- ============================================

-- LLM调用日志
CREATE TABLE llm_api_calls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    conversation_id UUID REFERENCES conversations(id) ON DELETE SET NULL,

    -- API信息
    provider VARCHAR(50) NOT NULL,
    -- 'openai', 'azure', 'anthropic', 'local'

    model VARCHAR(100) NOT NULL,

    -- Token使用
    prompt_tokens INTEGER NOT NULL,
    completion_tokens INTEGER NOT NULL,
    total_tokens INTEGER NOT NULL,

    -- 成本（美元）
    estimated_cost DECIMAL(10, 6),

    -- 响应时间（毫秒）
    latency_ms INTEGER,

    -- 状态
    status VARCHAR(20),
    -- 'success', 'error', 'timeout'

    error_message TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_api_calls_user ON llm_api_calls(user_id);
CREATE INDEX idx_api_calls_created_at ON llm_api_calls(created_at);
CREATE INDEX idx_api_calls_provider ON llm_api_calls(provider);

-- ============================================
-- 9. 系统配置
-- ============================================

-- 系统配置表
CREATE TABLE system_configs (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 插入默认配置
INSERT INTO system_configs (key, value, description) VALUES
('context_window_limit', '{"gpt-4": 8000, "gpt-3.5-turbo": 4000}', 'Model context window limits'),
('compression_threshold', '0.8', 'Trigger compression when context reaches 80% of limit'),
('embedding_model', '"text-embedding-3-small"', 'Default embedding model'),
('memory_retention_days', '90', 'Long-term memory retention period');

-- ============================================
-- 10. 视图（方便查询）
-- ============================================

-- 用户活跃度视图
CREATE VIEW user_activity_stats AS
SELECT
    u.id AS user_id,
    u.username,
    COUNT(DISTINCT c.id) AS total_conversations,
    COUNT(m.id) AS total_messages,
    COALESCE(SUM(m.token_count), 0) AS total_tokens,
    MAX(c.last_message_at) AS last_active_at
FROM users u
LEFT JOIN conversations c ON u.id = c.user_id
LEFT JOIN messages m ON c.id = m.conversation_id
GROUP BY u.id, u.username;

-- 角色使用统计视图
CREATE VIEW character_usage_stats AS
SELECT
    ch.id AS character_id,
    ch.name AS character_name,
    COUNT(DISTINCT c.id) AS conversation_count,
    COUNT(DISTINCT c.user_id) AS unique_users,
    AVG(ucr.closeness) AS avg_closeness,
    MAX(c.last_message_at) AS last_used_at
FROM characters ch
LEFT JOIN conversations c ON ch.id = c.character_id
LEFT JOIN user_character_relationships ucr ON ch.id = ucr.character_id
GROUP BY ch.id, ch.name;

-- ============================================
-- 11. 触发器（自动更新）
-- ============================================

-- 更新updated_at字段的触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 应用触发器到各表
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversations_updated_at BEFORE UPDATE ON conversations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 更新会话消息计数的触发器
CREATE OR REPLACE FUNCTION update_conversation_stats()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE conversations
    SET
        message_count = message_count + 1,
        total_tokens = total_tokens + COALESCE(NEW.token_count, 0),
        last_message_at = NEW.created_at
    WHERE id = NEW.conversation_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_conversation_on_message_insert
    AFTER INSERT ON messages
    FOR EACH ROW EXECUTE FUNCTION update_conversation_stats();

-- ============================================
-- 12. 常用查询函数
-- ============================================

-- 向量相似度搜索函数（带时间衰减）
CREATE OR REPLACE FUNCTION search_memories_with_decay(
    p_user_id UUID,
    p_character_id UUID,
    p_query_embedding vector(1536),
    p_limit INTEGER DEFAULT 5,
    p_decay_days INTEGER DEFAULT 30
)
RETURNS TABLE (
    memory_id UUID,
    content TEXT,
    similarity FLOAT,
    importance INTEGER,
    time_weight FLOAT,
    final_score FLOAT,
    created_at TIMESTAMP WITH TIME ZONE
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        m.id,
        m.content,
        1 - (m.embedding <=> p_query_embedding) AS similarity,
        m.importance,
        EXP(-EXTRACT(EPOCH FROM (NOW() - m.created_at)) / (p_decay_days * 86400.0)) AS time_weight,
        (
            (1 - (m.embedding <=> p_query_embedding)) * 0.5 +
            EXP(-EXTRACT(EPOCH FROM (NOW() - m.created_at)) / (p_decay_days * 86400.0)) * 0.3 +
            (m.importance / 10.0) * 0.2
        ) AS final_score,
        m.created_at
    FROM long_term_memories m
    WHERE m.user_id = p_user_id
      AND m.character_id = p_character_id
    ORDER BY final_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 完成！
-- ============================================

-- 创建数据库时执行:
-- psql -U postgres -d ai_chat_db -f database_schema.sql
