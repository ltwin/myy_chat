# Data Model Design: AI角色对话平台

**Version**: 1.0
**Date**: 2025-11-07
**Status**: Design Approved
**Database**: PostgreSQL 15 + pgvector extension

---

## Table of Contents

1. [Overview](#1-overview)
2. [Entity Relationship Diagram](#2-entity-relationship-diagram)
3. [Core Entities](#3-core-entities)
4. [Table Schemas](#4-table-schemas)
5. [Indexes & Optimization](#5-indexes--optimization)
6. [Data Migration](#6-data-migration)

---

## 1. Overview

### Database Architecture

- **Primary Database**: PostgreSQL 15
- **Extensions**: pgvector (向量存储)
- **Connection Pool**: pgBouncer (最大连接数200)
- **Replication**: 主从复制(1主2从,读写分离)
- **Backup**: 每日增量备份,每周全量备份

### Design Principles

1. **范式化**: 3NF设计,减少数据冗余
2. **索引优化**: 根据查询模式设计索引
3. **分区策略**: 大表(conversations, messages)按时间分区
4. **无外键约束**: 应用层维护引用完整性,支持未来分库分表
5. **雪花ID主键**: 使用Snowflake算法生成单调递增的64位ID
6. **软删除**: 关键数据(users, characters)使用`is_deleted`标记

### Snowflake ID Structure

本项目所有主键使用**雪花算法ID**(Snowflake),64位整数组成:

```
0 - 0000000000 0000000000 0000000000 0000000000 0 - 0000000000 - 000000000000
|   |-------------------------------------------|   |----------|   |----------|
符号  时间戳(41位,毫秒级,可用69年)                    机器ID(10位)   序列号(12位)
1位   支持到2089年                                  1024个节点     每毫秒4096个ID
```

**优势**:
- ✅ 单调递增,B-tree索引友好,无页分裂
- ✅ 包含时间信息,天然可排序
- ✅ 分布式生成,无需中心协调
- ✅ 性能优于UUID: 插入TPS提升5倍
- ✅ 存储空间: 8字节 vs UUID的16字节

**实现库**:
- Golang: `github.com/bwmarrin/snowflake`
- Python: `pysnowflake`

---

## 2. Entity Relationship Diagram

```
┌──────────────┐           ┌──────────────┐
│    users     │1        * │ characters   │
│──────────────│───────────│──────────────│
│ id (PK)      │           │ id (PK)      │
│ username     │           │ user_id (FK) │
│ email        │           │ name         │
│ password_hash│           │ personality  │
│ is_deleted   │           │ system_prompt│
└──────────────┘           └──────────────┘
       │                          │
       │1                         │1
       │                          │
       │*                         │*
┌──────────────┐           ┌──────────────┐
│credit_accounts│          │conversations │
│──────────────│           │──────────────│
│ user_id (PK) │           │ id (PK)      │
│ balance      │           │ user_id (FK) │
│ total_consumed│          │ character_id │
└──────────────┘           └──────────────┘
       │                          │
       │1                         │1
       │                          │
       │*                         │*
┌──────────────┐           ┌──────────────┐
│credit_transactions│      │   messages   │
│──────────────│           │──────────────│
│ id (PK)      │           │ id (PK)      │
│ user_id (FK) │           │ conversation_id│
│ amount       │           │ role         │
│ idempotency_key│         │ content      │
└──────────────┘           │ token_count  │
                           └──────────────┘
                                  │
                                  │1
                                  │
                                  │*
                           ┌──────────────┐
                           │   memories   │
                           │──────────────│
                           │ id (PK)      │
                           │ user_id (FK) │
                           │ character_id │
                           │ content      │
                           │ embedding    │
                           │ importance   │
                           └──────────────┘
```

---

## 3. Core Entities

### 3.1 User Domain

**Entities**: `users`, `user_profiles`, `sessions`

**Key Requirements**:
- FR-001: 支持邮箱/手机号注册
- FR-002: 邮箱/手机号验证
- FR-003: 登录、登出、密码管理
- FR-004a-d: 账号删除(30天冷静期)

---

### 3.2 Character Domain

**Entities**: `characters`, `character_templates`

**Key Requirements**:
- FR-006: 至少5个预设角色
- FR-007: 用户自定义角色
- FR-008: 编辑、删除角色
- FR-010: 每个角色独立记忆

---

### 3.3 Conversation Domain

**Entities**: `conversations`, `messages`

**Key Requirements**:
- FR-011: 实时文本对话
- FR-012: 响应时间<3秒
- FR-013: 完整对话历史
- FR-014: 多会话管理
- FR-016: 跨会话记忆共享

---

### 3.4 Memory Domain

**Entities**: `memories`, `relationships`, `events`

**Key Requirements**:
- FR-017: 自动提取个人信息
- FR-017a: 矛盾信息处理
- FR-018: 情感状态追踪
- FR-019: 关系等级
- FR-020: 重大事件记录
- FR-023: 记忆优先级自动调整

---

### 3.5 Billing Domain

**Entities**: `credit_accounts`, `credit_transactions`, `orders`

**Key Requirements**:
- FR-031: 积分账户
- FR-031a: 新用户赠送100积分
- FR-032: 积分计费
- FR-037: 余额为0时阻止发送
- 幂等性保证(idempotency_key)

---

## 4. Table Schemas

### 4.1 users (用户表)

```sql
CREATE TABLE users (
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
    deletion_scheduled_at TIMESTAMP WITH TIME ZONE,  -- 冷静期开始时间
    CONSTRAINT check_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$')
);

-- 索引
CREATE INDEX idx_users_email ON users(email) WHERE NOT is_deleted;
CREATE INDEX idx_users_username ON users(username) WHERE NOT is_deleted;
CREATE INDEX idx_users_deletion_scheduled ON users(deletion_scheduled_at) WHERE deletion_scheduled_at IS NOT NULL;
CREATE INDEX idx_users_created_at ON users(created_at);  -- 雪花ID已包含时间,但仍建索引以优化范围查询

-- 触发器: 自动更新updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE users IS '用户表 - 存储用户基本信息';
COMMENT ON COLUMN users.id IS '雪花ID主键,由应用层生成';
COMMENT ON COLUMN users.deletion_scheduled_at IS '账号删除计划时间,设置后30天自动删除';
```

---

### 4.2 user_profiles (用户画像表)

```sql
CREATE TABLE user_profiles (
    user_id BIGINT PRIMARY KEY,  -- 关联users.id,由应用层维护
    full_name VARCHAR(100),
    gender VARCHAR(20),  -- male, female, other
    birth_date DATE,
    location JSONB,  -- {"country": "China", "city": "Shanghai"}
    interests TEXT[],  -- 兴趣爱好数组
    occupation VARCHAR(100),
    bio TEXT,
    preferences JSONB,  -- 用户偏好设置
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);

CREATE TRIGGER update_user_profiles_updated_at BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE user_profiles IS '用户画像表 - 存储用户个人详细信息';
COMMENT ON COLUMN user_profiles.user_id IS '关联users.id（应用层维护外键关系）';
```

---

### 4.3 characters (AI角色表)

```sql
CREATE TABLE characters (
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
CREATE INDEX idx_characters_user_id ON characters(user_id) WHERE NOT is_deleted;
CREATE INDEX idx_characters_public ON characters(is_public) WHERE is_public = TRUE AND NOT is_deleted;
CREATE INDEX idx_characters_preset ON characters(is_preset) WHERE is_preset = TRUE;

CREATE TRIGGER update_characters_updated_at BEFORE UPDATE ON characters
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE characters IS 'AI角色表 - 预设角色和用户自定义角色';
COMMENT ON COLUMN characters.id IS '雪花ID主键';
COMMENT ON COLUMN characters.user_id IS '关联users.id(应用层维护), NULL表示预设角色';
```

---

### 4.4 conversations (会话表)

```sql
CREATE TABLE conversations (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 关联characters.id
    title VARCHAR(255),
    message_count INT DEFAULT 0,
    token_count INT DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_message_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_archived BOOLEAN DEFAULT FALSE
) PARTITION BY RANGE (started_at);  -- 按月分区

-- 创建分区表(示例:2025年各月)
CREATE TABLE conversations_2025_01 PARTITION OF conversations
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE conversations_2025_02 PARTITION OF conversations
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
-- ... 其他月份

-- 索引
CREATE INDEX idx_conversations_user_id ON conversations(user_id);
CREATE INDEX idx_conversations_character_id ON conversations(character_id);
CREATE INDEX idx_conversations_last_message ON conversations(last_message_at DESC);

COMMENT ON TABLE conversations IS '会话表 - 用户与AI角色的对话会话';
COMMENT ON COLUMN conversations.user_id IS '关联users.id(应用层维护)';
COMMENT ON COLUMN conversations.character_id IS '关联characters.id(应用层维护)';
```

---

### 4.5 messages (消息表)

```sql
CREATE TABLE messages (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    conversation_id BIGINT NOT NULL,  -- 关联conversations.id
    role VARCHAR(20) NOT NULL,  -- user, assistant, system
    content TEXT NOT NULL,
    token_count INT NOT NULL DEFAULT 0,
    metadata JSONB,  -- {"latency_ms": 2500, "model": "gpt-4", "cost": 0.05}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_role CHECK (role IN ('user', 'assistant', 'system'))
) PARTITION BY RANGE (created_at);  -- 按月分区

-- 创建分区表
CREATE TABLE messages_2025_01 PARTITION OF messages
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
CREATE TABLE messages_2025_02 PARTITION OF messages
    FOR VALUES FROM ('2025-02-01') TO ('2025-03-01');
-- ... 其他月份

-- 索引
CREATE INDEX idx_messages_conversation_id ON messages(conversation_id, created_at DESC);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);

COMMENT ON TABLE messages IS '消息表 - 对话消息记录(按月分区)';
COMMENT ON COLUMN messages.conversation_id IS '关联conversations.id(应用层维护)';
```

---

### 4.6 memories (记忆表) - 核心

```sql
CREATE EXTENSION IF NOT EXISTS vector;  -- pgvector扩展

CREATE TABLE memories (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 关联characters.id
    memory_type VARCHAR(20) NOT NULL,  -- USER_PROFILE, EVENT, EMOTION, CORE, TRIVIA
    content TEXT NOT NULL,
    embedding VECTOR(1536),  -- 向量维度1536 (SiliconFlow BGE-M3或OpenAI)
    importance INT NOT NULL CHECK (importance BETWEEN 1 AND 10),
    structured_data JSONB,  -- 结构化信息
    access_count INT DEFAULT 0,  -- 被检索次数
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_memory_type CHECK (memory_type IN ('USER_PROFILE', 'EVENT', 'EMOTION', 'CORE', 'TRIVIA'))
);

-- 向量索引(HNSW,高性能)
CREATE INDEX memories_embedding_hnsw_idx
ON memories USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- 其他索引
CREATE INDEX idx_memories_user_character ON memories(user_id, character_id);
CREATE INDEX idx_memories_type ON memories(memory_type);
CREATE INDEX idx_memories_importance ON memories(importance DESC);
CREATE INDEX idx_memories_created_at ON memories(created_at DESC);
CREATE INDEX idx_memories_access_count ON memories(access_count DESC);

COMMENT ON TABLE memories IS '记忆表 - 存储AI的长期记忆(包含向量Embedding)';
COMMENT ON COLUMN memories.user_id IS '关联users.id(应用层维护)';
COMMENT ON COLUMN memories.character_id IS '关联characters.id(应用层维护)';
COMMENT ON COLUMN memories.embedding IS '1536维向量,由SiliconFlow API或本地模型生成';
COMMENT ON COLUMN memories.importance IS '重要性评分1-10,影响检索排序';
```

---

### 4.7 relationships (关系表)

```sql
CREATE TABLE relationships (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 关联characters.id
    relation_type VARCHAR(20) DEFAULT 'friend',  -- friend, lover, mentor, etc.
    closeness DECIMAL(5,2) DEFAULT 50.0 CHECK (closeness BETWEEN 0 AND 100),
    emotional_bond JSONB,  -- {"trust": 80, "affection": 70, "respect": 90}
    interaction_count INT DEFAULT 0,
    total_conversation_time INT DEFAULT 0,  -- 秒
    last_interaction_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, character_id)
);

CREATE INDEX idx_relationships_user_id ON relationships(user_id);
CREATE INDEX idx_relationships_character_id ON relationships(character_id);
CREATE INDEX idx_relationships_closeness ON relationships(closeness DESC);

CREATE TRIGGER update_relationships_updated_at BEFORE UPDATE ON relationships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE relationships IS '关系表 - 用户与AI角色的关系强度';
COMMENT ON COLUMN relationships.user_id IS '关联users.id(应用层维护)';
COMMENT ON COLUMN relationships.character_id IS '关联characters.id(应用层维护)';
```

---

### 4.8 credit_accounts (积分账户表)

```sql
CREATE TABLE credit_accounts (
    user_id BIGINT PRIMARY KEY,  -- 关联users.id, 一对一关系
    balance DECIMAL(12,2) DEFAULT 100.0 CHECK (balance >= 0),  -- 新用户赠送100积分
    total_recharged DECIMAL(12,2) DEFAULT 0.0,
    total_consumed DECIMAL(12,2) DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_credit_accounts_updated_at BEFORE UPDATE ON credit_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE credit_accounts IS '积分账户表 - 用户积分余额';
COMMENT ON COLUMN credit_accounts.user_id IS '关联users.id(应用层维护), 一对一关系';
```

---

### 4.9 credit_transactions (积分交易表)

```sql
CREATE TABLE credit_transactions (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    transaction_type VARCHAR(20) NOT NULL,  -- RECHARGE, CONSUME, REFUND
    amount DECIMAL(12,2) NOT NULL,
    balance_after DECIMAL(12,2) NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE NOT NULL,  -- 幂等性保证
    metadata JSONB,  -- {"conversation_id": "xxx", "tokens": 1500}
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_transaction_type CHECK (transaction_type IN ('RECHARGE', 'CONSUME', 'REFUND'))
);

-- 索引
CREATE INDEX idx_transactions_user_id ON credit_transactions(user_id, created_at DESC);
CREATE INDEX idx_transactions_idempotency ON credit_transactions(idempotency_key);
CREATE INDEX idx_transactions_created_at ON credit_transactions(created_at DESC);

COMMENT ON TABLE credit_transactions IS '积分交易表 - 所有积分变动记录(幂等性)';
COMMENT ON COLUMN credit_transactions.user_id IS '关联users.id(应用层维护)';
COMMENT ON COLUMN credit_transactions.idempotency_key IS '幂等性Key,防止重复扣费';
```

---

### 4.10 orders (订单表)

```sql
CREATE TABLE orders (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    order_no VARCHAR(64) UNIQUE NOT NULL,  -- 订单号
    product_type VARCHAR(50) NOT NULL,  -- CREDIT_PACK
    product_name VARCHAR(255) NOT NULL,
    credits_amount DECIMAL(12,2) NOT NULL,
    price_yuan DECIMAL(10,2) NOT NULL,
    payment_method VARCHAR(50),  -- WECHAT_PAY, ALIPAY, STRIPE
    payment_status VARCHAR(20) DEFAULT 'PENDING',  -- PENDING, PAID, FAILED, REFUNDED
    paid_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_payment_status CHECK (payment_status IN ('PENDING', 'PAID', 'FAILED', 'REFUNDED'))
);

-- 索引
CREATE INDEX idx_orders_user_id ON orders(user_id, created_at DESC);
CREATE INDEX idx_orders_payment_status ON orders(payment_status);
CREATE INDEX idx_orders_order_no ON orders(order_no);

CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE orders IS '订单表 - 积分充值订单';
COMMENT ON COLUMN orders.user_id IS '关联users.id(应用层维护)';
```

---

### 4.11 admin_users (管理员表)

```sql
CREATE TABLE admin_users (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin',  -- admin, super_admin
    permissions JSONB,  -- ["user_management", "llm_config"]
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_role CHECK (role IN ('admin', 'super_admin'))
);

CREATE INDEX idx_admin_users_email ON admin_users(email);
CREATE INDEX idx_admin_users_is_active ON admin_users(is_active);

COMMENT ON TABLE admin_users IS '管理员表 - 后台管理系统用户';
```

---

### 4.12 system_configs (系统配置表)

```sql
CREATE TABLE system_configs (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 初始数据
INSERT INTO system_configs (key, value, description) VALUES
('llm_providers', '[
    {"name": "azure-openai", "api_key": "***", "priority": 1, "enabled": true},
    {"name": "claude", "api_key": "***", "priority": 2, "enabled": true}
]', 'LLM供应商配置'),
('credit_formula', '{"base_rate": 0.001, "unit": "token"}', '积分计费公式');

COMMENT ON TABLE system_configs IS '系统配置表 - Key-Value存储';
```

---

### 4.13 api_logs (API调用日志表)

```sql
CREATE TABLE api_logs (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    service_name VARCHAR(50) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    user_id BIGINT,  -- 关联users.id, 可为NULL(未认证请求)
    status_code INT NOT NULL,
    latency_ms INT NOT NULL,
    request_body JSONB,
    response_body JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (created_at);

-- 创建分区表
CREATE TABLE api_logs_2025_01 PARTITION OF api_logs
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
-- ... 其他月份

-- 索引
CREATE INDEX idx_api_logs_service ON api_logs(service_name, created_at DESC);
CREATE INDEX idx_api_logs_user_id ON api_logs(user_id, created_at DESC);
CREATE INDEX idx_api_logs_status ON api_logs(status_code) WHERE status_code >= 400;

COMMENT ON TABLE api_logs IS 'API调用日志表 - 用于审计和监控';
COMMENT ON COLUMN api_logs.user_id IS '关联users.id(应用层维护), 可为NULL';
```

---

## 5. Indexes & Optimization

### 5.1 Query Optimization

**常见查询及索引**:

1. **检索记忆**(向量相似度)
   ```sql
   -- 查询: 检索用户最相关的5条记忆
   SELECT id, content, 1 - (embedding <=> $1::vector) AS similarity
   FROM memories
   WHERE user_id = $2 AND character_id = $3
   ORDER BY embedding <=> $1::vector
   LIMIT 5;

   -- 索引: memories_embedding_hnsw_idx (已创建)
   -- 优化: SET hnsw.ef_search = 40;
   ```

2. **获取会话历史**
   ```sql
   -- 查询: 获取会话最近50条消息
   SELECT id, role, content, created_at
   FROM messages
   WHERE conversation_id = $1
   ORDER BY created_at DESC
   LIMIT 50;

   -- 索引: idx_messages_conversation_id
   ```

3. **积分扣除**(幂等性)
   ```sql
   -- 查询: 检查幂等性Key
   SELECT id FROM credit_transactions
   WHERE idempotency_key = $1;

   -- 索引: idx_transactions_idempotency (UNIQUE)
   ```

---

### 5.2 Connection Pooling

```python
# Python (asyncpg)
pool = await asyncpg.create_pool(
    dsn='postgresql://user:pass@localhost/myy_chat',
    min_size=10,
    max_size=50,
    command_timeout=60,
)
```

```go
// Go (pgx)
config, _ := pgxpool.ParseConfig("postgres://user:pass@localhost/myy_chat")
config.MaxConns = 50
config.MinConns = 10
config.MaxConnLifetime = time.Hour
pool, _ := pgxpool.NewWithConfig(context.Background(), config)
```

---

### 5.3 Partitioning Strategy

**分区表**:
- `conversations`: 按`started_at`月度分区
- `messages`: 按`created_at`月度分区
- `api_logs`: 按`created_at`月度分区

**维护**:
```sql
-- 自动创建下月分区(通过cron job)
CREATE TABLE conversations_2025_12 PARTITION OF conversations
    FOR VALUES FROM ('2025-12-01') TO ('2026-01-01');

-- 归档/删除旧分区(保留12个月数据)
DROP TABLE messages_2024_01;
```

---

## 6. Data Migration

### 6.1 Initial Migration

```sql
-- scripts/init-db.sql
\c myy_chat;

-- 1. 创建扩展
CREATE EXTENSION IF NOT EXISTS "vector";  -- pgvector用于向量检索

-- 2. 创建表(按依赖顺序)
\i migrations/001_create_users.sql
\i migrations/002_create_characters.sql
\i migrations/003_create_conversations.sql
\i migrations/004_create_messages.sql
\i migrations/005_create_memories.sql
\i migrations/006_create_relationships.sql
\i migrations/007_create_billing.sql
\i migrations/008_create_admin.sql

-- 3. 种子数据
\i migrations/seed_preset_characters.sql
```

---

### 6.2 Seed Data

```sql
-- migrations/seed_preset_characters.sql
-- 注意: 雪花ID需要由应用层生成,这里使用固定的示例ID
INSERT INTO characters (id, name, personality, system_prompt, is_preset, created_at) VALUES
(
    1000000000000001,  -- 雪花ID示例(由应用层ID生成器生成)
    '小雪',
    '{"mbti": "ENFP", "traits": ["活泼", "好奇", "善解人意"]}',
    '你是小雪,一个20岁的大学生,性格活泼开朗,对世界充满好奇。',
    TRUE,
    CURRENT_TIMESTAMP
),
(
    1000000000000002,
    '阿明',
    '{"mbti": "INTJ", "traits": ["理性", "深思", "博学"]}',
    '你是阿明,一个30岁的程序员,理性思考,知识渊博。',
    TRUE,
    CURRENT_TIMESTAMP
);
-- ... 其他预设角色
```

---

### 6.3 Migration Tools

**推荐工具**: golang-migrate

```bash
# 安装
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 创建迁移
migrate create -ext sql -dir migrations -seq create_users_table

# 执行迁移
migrate -path migrations -database "postgres://localhost:5432/myy_chat?sslmode=disable" up

# 回滚
migrate -path migrations -database "..." down 1
```

---

## 7. 应用层关系维护

### 7.1 设计理念

本数据模型**不使用数据库外键约束**,而是在应用层维护表间关系,原因如下:

1. **支持分库分表**: 外键约束会限制水平分片能力
2. **提升性能**: 避免外键检查的锁开销,提高高并发场景吞吐量
3. **灵活性**: 业务逻辑变更时无需修改数据库约束
4. **微服务友好**: 跨服务边界的数据关系无法使用外键

**代价**: 需要在业务代码中显式处理级联删除、级联更新等逻辑。

---

### 7.2 级联删除示例

#### Go 实现 (Transaction-Based)

```go
package data

import (
    "context"
    "database/sql"
)

// UserRepo 用户仓储
type UserRepo struct {
    db *sql.DB
}

// DeleteUser 删除用户及所有关联数据(级联删除)
func (r *UserRepo) DeleteUser(ctx context.Context, userID int64) error {
    // 开启事务
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. 删除用户画像
    _, err = tx.ExecContext(ctx, "DELETE FROM user_profiles WHERE user_id = $1", userID)
    if err != nil {
        return err
    }

    // 2. 删除用户创建的角色
    _, err = tx.ExecContext(ctx, "DELETE FROM characters WHERE user_id = $1", userID)
    if err != nil {
        return err
    }

    // 3. 删除会话及其消息(先删除子表)
    _, err = tx.ExecContext(ctx, `
        DELETE FROM messages
        WHERE conversation_id IN (
            SELECT id FROM conversations WHERE user_id = $1
        )
    `, userID)
    if err != nil {
        return err
    }

    _, err = tx.ExecContext(ctx, "DELETE FROM conversations WHERE user_id = $1", userID)
    if err != nil {
        return err
    }

    // 4. 删除记忆
    _, err = tx.ExecContext(ctx, "DELETE FROM memories WHERE user_id = $1", userID)
    if err != nil {
        return err
    }

    // 5. 删除关系
    _, err = tx.ExecContext(ctx, "DELETE FROM relationships WHERE user_id = $1", userID)
    if err != nil {
        return err
    }

    // 6. 删除积分交易记录
    _, err = tx.ExecContext(ctx, "DELETE FROM credit_transactions WHERE user_id = $1", userID)
    if err != nil {
        return err
    }

    // 7. 删除积分账户
    _, err = tx.ExecContext(ctx, "DELETE FROM credit_accounts WHERE user_id = $1", userID)
    if err != nil {
        return err
    }

    // 8. 删除订单
    _, err = tx.ExecContext(ctx, "DELETE FROM orders WHERE user_id = $1", userID)
    if err != nil {
        return err
    }

    // 9. 最后删除用户本身
    result, err := tx.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
    if err != nil {
        return err
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return sql.ErrNoRows
    }

    // 提交事务
    return tx.Commit()
}

// DeleteCharacter 删除角色及其关联数据
func (r *UserRepo) DeleteCharacter(ctx context.Context, characterID int64) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // 1. 删除与该角色的所有会话消息
    _, err = tx.ExecContext(ctx, `
        DELETE FROM messages
        WHERE conversation_id IN (
            SELECT id FROM conversations WHERE character_id = $1
        )
    `, characterID)
    if err != nil {
        return err
    }

    // 2. 删除与该角色的所有会话
    _, err = tx.ExecContext(ctx, "DELETE FROM conversations WHERE character_id = $1", characterID)
    if err != nil {
        return err
    }

    // 3. 删除与该角色的记忆
    _, err = tx.ExecContext(ctx, "DELETE FROM memories WHERE character_id = $1", characterID)
    if err != nil {
        return err
    }

    // 4. 删除与该角色的关系
    _, err = tx.ExecContext(ctx, "DELETE FROM relationships WHERE character_id = $1", characterID)
    if err != nil {
        return err
    }

    // 5. 删除角色本身
    _, err = tx.ExecContext(ctx, "DELETE FROM characters WHERE id = $1", characterID)
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

#### Python 实现 (Async Transaction)

```python
from typing import Optional
import asyncpg

class UserRepository:
    def __init__(self, pool: asyncpg.Pool):
        self.pool = pool

    async def delete_user(self, user_id: int) -> bool:
        """删除用户及所有关联数据(级联删除)"""
        async with self.pool.acquire() as conn:
            async with conn.transaction():
                # 1. 删除用户画像
                await conn.execute(
                    "DELETE FROM user_profiles WHERE user_id = $1",
                    user_id
                )

                # 2. 删除用户创建的角色
                await conn.execute(
                    "DELETE FROM characters WHERE user_id = $1",
                    user_id
                )

                # 3. 删除会话及其消息
                await conn.execute(
                    """
                    DELETE FROM messages
                    WHERE conversation_id IN (
                        SELECT id FROM conversations WHERE user_id = $1
                    )
                    """,
                    user_id
                )
                await conn.execute(
                    "DELETE FROM conversations WHERE user_id = $1",
                    user_id
                )

                # 4. 删除记忆
                await conn.execute(
                    "DELETE FROM memories WHERE user_id = $1",
                    user_id
                )

                # 5. 删除关系
                await conn.execute(
                    "DELETE FROM relationships WHERE user_id = $1",
                    user_id
                )

                # 6. 删除积分交易记录
                await conn.execute(
                    "DELETE FROM credit_transactions WHERE user_id = $1",
                    user_id
                )

                # 7. 删除积分账户
                await conn.execute(
                    "DELETE FROM credit_accounts WHERE user_id = $1",
                    user_id
                )

                # 8. 删除订单
                await conn.execute(
                    "DELETE FROM orders WHERE user_id = $1",
                    user_id
                )

                # 9. 最后删除用户
                result = await conn.execute(
                    "DELETE FROM users WHERE id = $1",
                    user_id
                )

                # 检查是否删除成功
                return "DELETE 1" in result

    async def delete_character(self, character_id: int) -> bool:
        """删除角色及其关联数据"""
        async with self.pool.acquire() as conn:
            async with conn.transaction():
                # 1. 删除会话消息
                await conn.execute(
                    """
                    DELETE FROM messages
                    WHERE conversation_id IN (
                        SELECT id FROM conversations WHERE character_id = $1
                    )
                    """,
                    character_id
                )

                # 2. 删除会话
                await conn.execute(
                    "DELETE FROM conversations WHERE character_id = $1",
                    character_id
                )

                # 3. 删除记忆
                await conn.execute(
                    "DELETE FROM memories WHERE character_id = $1",
                    character_id
                )

                # 4. 删除关系
                await conn.execute(
                    "DELETE FROM relationships WHERE character_id = $1",
                    character_id
                )

                # 5. 删除角色
                result = await conn.execute(
                    "DELETE FROM characters WHERE id = $1",
                    character_id
                )

                return "DELETE 1" in result
```

---

### 7.3 关系完整性检查

虽然没有外键约束,但应在应用层添加引用完整性检查:

```go
// 创建会话前检查用户和角色是否存在
func (r *ConversationRepo) CreateConversation(ctx context.Context, userID, characterID int64) error {
    // 1. 检查用户是否存在
    var exists bool
    err := r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&exists)
    if err != nil || !exists {
        return fmt.Errorf("user %d not found", userID)
    }

    // 2. 检查角色是否存在
    err = r.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM characters WHERE id = $1)", characterID).Scan(&exists)
    if err != nil || !exists {
        return fmt.Errorf("character %d not found", characterID)
    }

    // 3. 创建会话
    // ...
}
```

---

### 7.4 最佳实践

1. **事务保证**: 所有级联操作必须在事务中执行,保证原子性
2. **删除顺序**: 先删除子表,再删除父表,避免孤儿数据
3. **软删除优先**: 对用户、角色等重要数据使用`deleted_at`,而非物理删除
4. **异步清理**: 大量关联数据的删除可使用消息队列异步处理
5. **监控告警**: 记录所有级联删除操作,监控异常情况

---

## Summary

本数据模型设计涵盖了AI陪伴平台的所有核心业务:

- **13个核心表**: 用户、角色、对话、记忆、积分等
- **向量检索**: pgvector HNSW索引,支持<100ms查询
- **分区策略**: 大表按月分区,支持数据归档
- **幂等性**: idempotency_key保证关键操作不重复
- **软删除**: 重要数据(用户、角色)支持30天恢复期

**Next Steps**:
1. 生成contracts/(Protobuf API定义)
2. 生成quickstart.md(本地开发环境搭建)
3. 执行`/speckit.tasks`生成任务清单

---

**Document Version**: 1.0
**Last Updated**: 2025-11-07
**Status**: ✅ Ready for Implementation
