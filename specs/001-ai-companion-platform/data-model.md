# Data Model Design: AI角色对话平台

**Version**: 1.2
**Date**: 2025-12-27
**Status**: Design Approved
**Database**: PostgreSQL 16 + pgvector + Neo4j 5.x (Graph Database)

**v1.2 更新**:
- 完善遗忘曲线字段 (memories: last_boosted_at, memory_strength, boost_history)
- 移除 user_profiles.preferences (AI 推断偏好只在 user_portraits)
- important_events 移除 embedding，改用 related_memory_ids 关联检索
- 添加同步增强字段 (sync_version, sync_failed_count, last_sync_error)
- 向量索引优先级标注

---

## Table of Contents

1. [Overview](#1-overview)
2. [Entity Relationship Diagram](#2-entity-relationship-diagram)
3. [Core Entities](#3-core-entities)
4. [Table Schemas](#4-table-schemas)
5. [Indexes & Optimization](#5-indexes--optimization)
6. [Data Migration](#6-data-migration)
7. [应用层关系维护](#7-应用层关系维护)
8. [Long-Term Memory System (长期记忆系统)](#8-long-term-memory-system-长期记忆系统)
9. [Neo4j Graph Database (图数据库)](#9-neo4j-graph-database-图数据库)
10. [PostgreSQL-Neo4j Sync Strategy (数据同步策略)](#10-postgresql-neo4j-sync-strategy-数据同步策略)

---

## 1. Overview

### Database Architecture

- **Primary Database**: PostgreSQL 16
- **Extensions**: pgvector (向量存储)
- **Graph Database**: Neo4j 5.x (知识图谱、关系网络)
- **Connection Pool**: pgBouncer (最大连接数200)
- **Replication**: 主从复制(1主2从,读写分离)
- **Backup**: 每日增量备份,每周全量备份
- **Sync Strategy**: PostgreSQL → Neo4j (事件驱动同步)

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

**Entities**: `credit_accounts`, `credit_transactions`, `orders`, `credit_packages`

**Key Requirements**:
- FR-031: 积分账户
- FR-031a: 新用户赠送100积分
- FR-032: 积分计费
- FR-032a: 管理员可配置 USD→积分转换公式
- FR-035a: 积分套餐管理
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
    interests TEXT[],  -- 兴趣爱好数组 (用户手动填写)
    occupation VARCHAR(100),
    bio TEXT,
    -- 注意: AI 推断的 preferences/personality 存储在 user_portraits 表中
    -- user_profiles 只存储用户主动填写的静态信息
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
    -- v1.1 新增: Mnemosyne 记忆评分字段
    surprise_score DECIMAL(3,2) CHECK (surprise_score BETWEEN 0 AND 1),  -- 惊讶度 0-1 (基于embedding距离计算)
    decay_factor DECIMAL(5,4) DEFAULT 1.0,  -- 遗忘衰减因子 (Ebbinghaus)
    boost_count INT DEFAULT 0,  -- 强化次数 (被回忆/引用次数)
    source_type VARCHAR(50) DEFAULT 'EXTRACTED',  -- EXTRACTED, INFERRED, USER_STATED
    connectivity INT DEFAULT 0,  -- 与其他记忆的关联数
    -- v1.2 新增: 完善遗忘曲线算法所需字段
    last_boosted_at TIMESTAMP WITH TIME ZONE,  -- 上次强化时间 (用于计算衰减)
    memory_strength DECIMAL(5,4) DEFAULT 1.0 CHECK (memory_strength BETWEEN 0 AND 1),  -- 当前记忆强度
    boost_history JSONB DEFAULT '[]',  -- 强化时间序列 [{"time": "2025-01-01T00:00:00Z", "trigger": "conversation"}]
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- v1.3 新增: 多态关联与统一检索支持
    target_table VARCHAR(50),  -- 关联表名: user_portraits, important_events, conversation_summaries, memory_graph_nodes
    target_id BIGINT,          -- 关联表主键ID

    -- v1.3.1 新增: 社区化记忆可见性与写入来源（AI虚拟社区）
    visibility_scope VARCHAR(20) DEFAULT 'OWNER_PRIVATE', -- OWNER_PRIVATE / PUBLIC / VISITOR_PRIVATE
    actor_user_id BIGINT,  -- 触发/写入该记忆的用户ID（用于审计与风控；可能为NULL表示系统任务）
    actor_role VARCHAR(20) DEFAULT 'owner', -- owner / visitor / moderator / system
    
    CONSTRAINT check_memory_type CHECK (memory_type IN ('USER_PROFILE', 'EVENT', 'EMOTION', 'CORE', 'TRIVIA', 'ENTITY', 'SUMMARY')),
    CONSTRAINT check_source_type CHECK (source_type IN ('EXTRACTED', 'INFERRED', 'USER_STATED')),
    CONSTRAINT check_visibility_scope CHECK (visibility_scope IN ('OWNER_PRIVATE', 'PUBLIC', 'VISITOR_PRIVATE')),
    CONSTRAINT check_actor_role CHECK (actor_role IN ('owner', 'visitor', 'moderator', 'system'))
);

-- 向量索引(HNSW,高性能) - 作为全局统一的向量入口
CREATE INDEX memories_embedding_hnsw_idx
ON memories USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- 其他索引
CREATE INDEX idx_memories_user_character ON memories(user_id, character_id);
CREATE INDEX idx_memories_type ON memories(memory_type);
CREATE INDEX idx_memories_importance ON memories(importance DESC);
CREATE INDEX idx_memories_created_at ON memories(created_at DESC);
CREATE INDEX idx_memories_access_count ON memories(access_count DESC);
CREATE INDEX idx_memories_target ON memories(target_table, target_id); -- 加速反向查询

COMMENT ON TABLE memories IS '记忆表 - 存储AI的长期记忆(全局向量索引，支持多态关联)';
COMMENT ON COLUMN memories.user_id IS '关联users.id(应用层维护)';
COMMENT ON COLUMN memories.character_id IS '关联characters.id(应用层维护)';
COMMENT ON COLUMN memories.embedding IS '1536维向量,由SiliconFlow API或本地模型生成';
COMMENT ON COLUMN memories.importance IS '重要性评分1-10,影响检索排序';
COMMENT ON COLUMN memories.last_boosted_at IS '上次被唤醒/强化时间,用于遗忘曲线计算';
COMMENT ON COLUMN memories.target_table IS '多态关联: 关联到具体业务表(如important_events)';
COMMENT ON COLUMN memories.visibility_scope IS '记忆可见性范围：Owner私密/公共/访客私密(社区化场景)';
COMMENT ON COLUMN memories.actor_user_id IS '写入来源用户，用于审计与风控';
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
    -- v1.1 新增: 情感历史和里程碑
    trust_score DECIMAL(5,2) DEFAULT 50.0 CHECK (trust_score BETWEEN 0 AND 100),  -- 信任度
    affection_score DECIMAL(5,2) DEFAULT 50.0 CHECK (affection_score BETWEEN 0 AND 100),  -- 喜爱度
    emotional_history JSONB DEFAULT '[]',  -- 情感变化历史 [{date, event, delta}]
    milestones JSONB DEFAULT '[]',  -- 关系里程碑 [{date, type, description}]
    communication_patterns JSONB,  -- 沟通模式分析 {preferred_time, avg_length, topics}
    first_interaction_at TIMESTAMP WITH TIME ZONE,
    -- v1.3.1 新增: 社区化与安全锁字段
    is_owner BOOLEAN DEFAULT FALSE,  -- 是否Owner
    permission_level VARCHAR(20) DEFAULT 'PUBLIC', -- OWNER / PUBLIC / BLOCKED
    loyalty_lock BOOLEAN DEFAULT TRUE,  -- 是否启用忠诚/安全锁（默认开启）
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
COMMENT ON COLUMN relationships.is_owner IS '是否为该AI伴侣的Owner';
COMMENT ON COLUMN relationships.permission_level IS '权限等级：OWNER/Public/Blocked（社区化场景）';
COMMENT ON COLUMN relationships.loyalty_lock IS '忠诚/安全锁开关：防越权与防拐跑';
```

---

### 4.7a character_policies (角色安全策略表) *(v1.3.1 新增)*

用于集中管理社区化场景下的权限策略（记忆分区、忠诚锁阈值、告警开关等）。

```sql
CREATE TABLE character_policies (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    character_id BIGINT NOT NULL,  -- 关联characters.id
    owner_user_id BIGINT NOT NULL,  -- Owner 用户ID（便于审计与策略判定）

    -- 忠诚/安全锁策略
    loyalty_lock_enabled BOOLEAN DEFAULT TRUE,
    owner_trust_floor DECIMAL(5,2) DEFAULT 80.0 CHECK (owner_trust_floor BETWEEN 0 AND 100),
    owner_affection_floor DECIMAL(5,2) DEFAULT 80.0 CHECK (owner_affection_floor BETWEEN 0 AND 100),
    visitor_closeness_cap DECIMAL(5,2) DEFAULT 60.0 CHECK (visitor_closeness_cap BETWEEN 0 AND 100),

    -- 记忆分区策略
    allow_public_memory_write BOOLEAN DEFAULT TRUE,
    allow_visitor_memory_write BOOLEAN DEFAULT TRUE,
    allow_owner_private_memory_read BOOLEAN DEFAULT TRUE,

    -- 告警策略
    enable_owner_notifications BOOLEAN DEFAULT TRUE,

    -- 其他可扩展策略
    policy_json JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(character_id)
);

CREATE INDEX idx_character_policies_owner ON character_policies(owner_user_id);

COMMENT ON TABLE character_policies IS '角色安全策略表 - 社区化权限/忠诚锁/告警等策略配置';
```

---

### 4.7b community_audit_events (社区安全审计事件表) *(v1.3.1 新增)*

记录越权请求、隐私拒答、提示词注入、拐跑诱导等安全事件，用于风控与后台治理。

```sql
CREATE TABLE community_audit_events (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    character_id BIGINT NOT NULL,  -- 关联characters.id
    owner_user_id BIGINT,  -- 关联users.id（可选）
    actor_user_id BIGINT,  -- 触发者用户ID（可选）
    conversation_id BIGINT, -- 关联conversations.id（可选）

    event_type VARCHAR(50) NOT NULL, -- PRIVACY_REQUEST / PROMPT_INJECTION / OWNERSHIP_TRANSFER_ATTEMPT / SEDUCTION / ABUSE
    severity VARCHAR(10) NOT NULL DEFAULT 'LOW', -- LOW / MEDIUM / HIGH / CRITICAL
    summary TEXT,  -- 事件摘要（避免存完整原文）
    metadata JSONB DEFAULT '{}',

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_severity CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL'))
);

CREATE INDEX idx_audit_events_character_time ON community_audit_events(character_id, created_at DESC);
CREATE INDEX idx_audit_events_actor_time ON community_audit_events(actor_user_id, created_at DESC);

COMMENT ON TABLE community_audit_events IS '社区安全审计事件表 - 记录高风险行为与策略触发';
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

### 4.11 credit_packages (积分套餐表)

```sql
CREATE TABLE credit_packages (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    package_name VARCHAR(100) NOT NULL,  -- 套餐名称 (如: "新手福利包")
    credits_amount INT NOT NULL,  -- 积分数量
    price_yuan DECIMAL(10,2) NOT NULL,  -- RMB 价格
    bonus_percentage DECIMAL(5,2) DEFAULT 0,  -- 额外赠送百分比 (如: 10.00 表示赠送10%)
    sort_order INT DEFAULT 0,  -- 排序权重
    is_active BOOLEAN DEFAULT TRUE,  -- 是否上架
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_credits_positive CHECK (credits_amount > 0),
    CONSTRAINT check_price_positive CHECK (price_yuan > 0),
    CONSTRAINT check_bonus_range CHECK (bonus_percentage >= 0 AND bonus_percentage <= 100)
);

-- 索引
CREATE INDEX idx_credit_packages_active ON credit_packages(is_active, sort_order);

CREATE TRIGGER update_credit_packages_updated_at BEFORE UPDATE ON credit_packages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 初始数据
INSERT INTO credit_packages (id, package_name, credits_amount, price_yuan, bonus_percentage, sort_order, is_active) VALUES
(1, '体验包', 100, 1.00, 0, 1, TRUE),
(2, '基础包', 1000, 9.90, 0, 2, TRUE),
(3, '标准包', 5000, 45.00, 10.00, 3, TRUE),  -- 赠送10%
(4, '豪华包', 10000, 88.00, 15.00, 4, TRUE),  -- 赠送15%
(5, '尊享包', 50000, 399.00, 25.00, 5, TRUE);  -- 赠送25%

COMMENT ON TABLE credit_packages IS '积分套餐表 - 充值套餐配置 (FR-035a)';
COMMENT ON COLUMN credit_packages.bonus_percentage IS '赠送百分比,如10.00表示额外赠送10%积分';
COMMENT ON COLUMN credit_packages.sort_order IS '排序权重,数字越小越靠前';
```

---

### 4.12 admin_users (管理员表)

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

### 4.13 system_configs (系统配置表)

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
('credit_formula', '{"base_rate": 0.001, "unit": "token"}', '积分计费公式(已废弃,请使用credit_price_mapping)'),
('credit_price_mapping', '{
    "usd_to_credit_rate": 1000,
    "platform_margin": 0.3,
    "model_multipliers": {
        "gpt-4o": 1.0,
        "gpt-4o-mini": 0.5,
        "claude-3-5-sonnet": 0.8,
        "claude-3-5-haiku": 0.4,
        "gemini-2.0-flash": 0.6
    },
    "description": "1 USD = 1000 积分, 30% 平台利润, 不同模型有定价倍数"
}', '积分与USD成本转换配置 (FR-032a)');

COMMENT ON TABLE system_configs IS '系统配置表 - Key-Value存储';
```

---

### 4.14 api_logs (API调用日志表)

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

## 8. Long-Term Memory System (长期记忆系统)

本节定义五层记忆架构所需的新增表结构，基于 Mnemosyne、PRIME、MemoryBank 等前沿研究设计。

### 8.1 user_portraits (用户画像表)

存储每个用户针对每个角色的动态画像信息 (PRIME 语义记忆)。

```sql
CREATE TABLE user_portraits (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 每个角色对用户有独立画像视角

    -- 静态信息 (人口统计学)
    demographics JSONB,  -- {"age": 28, "gender": "male", "occupation": "engineer", "location": "Shanghai"}

    -- 性格特征 (PRIME 语义记忆)
    personality JSONB,  -- {"mbti": "INTJ", "big_five": {"openness": 0.8, ...}, "values": ["family", "career"]}

    -- 动态偏好 (MemoryBank 强化机制)
    preferences JSONB,  -- {"topics": ["tech", "music"], "communication_style": "direct", "emotional_needs": ["support"]}

    -- 核心摘要 (Mnemosyne Core Summary)
    core_summary TEXT,  -- 固定长度用户人格概要 (~500字)
    -- v1.3: 为了支持全局模糊检索，需同步插入到 memories 表
    -- 建议应用层在写入本表时，同时写入 memories 表 (target_table='user_portraits')
    core_summary_embedding VECTOR(1536),  -- 用于快速检索

    -- 知识实体引用 (Neo4j 节点ID列表)
    known_entities JSONB DEFAULT '[]',  -- ["neo4j_node_id_1", "neo4j_node_id_2"]

    -- 元数据
    confidence_score DECIMAL(3,2) DEFAULT 0.5 CHECK (confidence_score BETWEEN 0 AND 1),  -- 画像置信度
    inference_version INT DEFAULT 1,  -- 画像推断版本
    last_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(user_id, character_id)
);

-- 索引
CREATE INDEX idx_user_portraits_user_character ON user_portraits(user_id, character_id);
CREATE INDEX user_portraits_core_embedding_idx
ON user_portraits USING hnsw (core_summary_embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

CREATE TRIGGER update_user_portraits_updated_at BEFORE UPDATE ON user_portraits
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE user_portraits IS '用户画像表 - 存储每个用户对每个角色的动态画像 (PRIME语义记忆)';
COMMENT ON COLUMN user_portraits.core_summary IS 'Mnemosyne核心摘要,固定长度用户人格概要';
COMMENT ON COLUMN user_portraits.confidence_score IS '画像置信度,0-1之间,随着交互增加而提升';
```

---

### 8.2 emotional_states (情感状态表)

记录用户情感状态的时序数据，基于 Russell 环形情感模型。

```sql
CREATE TABLE emotional_states (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 关联characters.id
    conversation_id BIGINT,  -- 关联conversations.id (可选)
    message_id BIGINT,  -- 触发情感变化的消息ID

    -- Russell 环形情感模型
    valence DECIMAL(3,2) NOT NULL CHECK (valence BETWEEN -1 AND 1),  -- 效价: -1(消极) to +1(积极)
    arousal DECIMAL(3,2) NOT NULL CHECK (arousal BETWEEN 0 AND 1),   -- 唤醒度: 0(平静) to 1(激动)

    -- 离散情感分类
    primary_emotion VARCHAR(50) NOT NULL,  -- joy, sadness, anger, fear, surprise, disgust, neutral
    secondary_emotion VARCHAR(50),  -- 次要情感

    -- 情感强度
    intensity DECIMAL(3,2) DEFAULT 0.5 CHECK (intensity BETWEEN 0 AND 1),

    -- 触发内容
    trigger_content TEXT,  -- 触发情感变化的对话内容
    trigger_type VARCHAR(50),  -- USER_MESSAGE, AI_RESPONSE, EVENT, EXTERNAL

    -- 置信度
    confidence DECIMAL(3,2) DEFAULT 0.7 CHECK (confidence BETWEEN 0 AND 1),

    -- 时间戳
    assessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_primary_emotion CHECK (primary_emotion IN (
        'joy', 'sadness', 'anger', 'fear', 'surprise', 'disgust', 'neutral',
        'love', 'trust', 'anticipation', 'anxiety', 'frustration', 'contentment'
    ))
);

-- 索引: 情感趋势分析
CREATE INDEX idx_emotional_states_user_char_time
ON emotional_states(user_id, character_id, assessed_at DESC);

CREATE INDEX idx_emotional_states_emotion
ON emotional_states(primary_emotion);

CREATE INDEX idx_emotional_states_valence_arousal
ON emotional_states(valence, arousal);

COMMENT ON TABLE emotional_states IS '情感状态表 - 记录用户情感时序数据 (Russell环形模型)';
COMMENT ON COLUMN emotional_states.valence IS '效价: -1(消极)到+1(积极)';
COMMENT ON COLUMN emotional_states.arousal IS '唤醒度: 0(平静)到1(激动)';
```

---

### 8.3 important_events (重要事件表)

存储用户生活中的重要事件，使用 Mnemosyne 惊讶度评分系统。

```sql
CREATE TABLE important_events (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 关联characters.id

    -- 事件内容
    event_type VARCHAR(50) NOT NULL,  -- 事件类型
    title VARCHAR(255) NOT NULL,  -- 事件标题
    content TEXT NOT NULL,  -- 事件详细描述
    summary TEXT,  -- 事件摘要 (用于上下文注入)

    -- Mnemosyne 评分系统
    surprise_score DECIMAL(3,2) DEFAULT 0.5 CHECK (surprise_score BETWEEN 0 AND 1),  -- 惊讶度 (基于embedding距离计算)
    emotional_intensity DECIMAL(3,2) DEFAULT 0.5 CHECK (emotional_intensity BETWEEN 0 AND 1),  -- 情感强度
    connectivity INT DEFAULT 0,  -- 关联记忆数
    boost_count INT DEFAULT 0,  -- 被回忆次数
    importance DECIMAL(5,2) DEFAULT 50.0,  -- 综合重要性分数 (0-100)

    -- 时间语义
    occurred_at TIMESTAMP WITH TIME ZONE,  -- 事件发生时间
    temporal_context VARCHAR(255),  -- 时间语义标注: "你28岁生日那天", "上周五"
    is_recurring BOOLEAN DEFAULT FALSE,  -- 是否周期性事件

    -- v1.2 优化: 不再存储独立 embedding，改为引用 memories 表
    -- 通过 related_memory_ids 关联检索，避免冗余存储
    related_memory_ids BIGINT[] DEFAULT '{}',  -- 关联的 memories.id 列表

    -- 关联对话
    source_conversation_id BIGINT,  -- 事件来源会话
    source_message_ids BIGINT[],  -- 事件来源消息

    -- 参与实体 (Neo4j 节点引用)
    participants JSONB DEFAULT '[]',  -- ["entity_id_1", "entity_id_2"]
    locations JSONB DEFAULT '[]',  -- ["location_entity_id"]

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_event_type CHECK (event_type IN (
        'MILESTONE',          -- 里程碑: 生日、纪念日、毕业
        'EMOTIONAL_PEAK',     -- 情感高峰: 极度开心/悲伤时刻
        'LIFE_CHANGE',        -- 人生变化: 搬家、换工作、分手
        'KEY_CONVERSATION',   -- 关键对话: 重要讨论、深度交流
        'ACHIEVEMENT',        -- 成就: 升职、获奖、完成目标
        'RELATIONSHIP_CHANGE', -- 关系变化: 新朋友、关系升级
        'HEALTH_EVENT',       -- 健康事件: 生病、康复
        'TRAVEL',             -- 旅行: 出差、度假
        'CUSTOM'              -- 自定义事件
    ))
);

-- 索引 (v1.2: 移除 embedding 索引，通过 related_memory_ids 关联检索)
CREATE INDEX idx_important_events_user_char ON important_events(user_id, character_id);
CREATE INDEX idx_important_events_type ON important_events(event_type);
CREATE INDEX idx_important_events_importance ON important_events(importance DESC);
CREATE INDEX idx_important_events_occurred_at ON important_events(occurred_at DESC);
CREATE INDEX idx_important_events_surprise ON important_events(surprise_score DESC);
CREATE INDEX idx_important_events_memory_refs ON important_events USING GIN (related_memory_ids);

COMMENT ON TABLE important_events IS '重要事件表 - 存储用户生活中的重要事件 (Mnemosyne惊讶度评分)';
COMMENT ON COLUMN important_events.surprise_score IS '惊讶度: 0-1, 基于embedding距离计算，越高越值得记忆';
COMMENT ON COLUMN important_events.temporal_context IS '时间语义标注,如"你28岁生日那天"';
COMMENT ON COLUMN important_events.related_memory_ids IS '关联的memories记录，用于检索相关向量';
```

---

### 8.4 memory_graph_nodes (图记忆节点表)

PostgreSQL 端存储图节点的元数据，与 Neo4j 保持同步。

```sql
CREATE TABLE memory_graph_nodes (
    id BIGINT PRIMARY KEY,  -- 雪花ID (同时作为 Neo4j 节点属性)
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 关联characters.id

    -- 节点信息
    node_type VARCHAR(50) NOT NULL,  -- 节点类型
    name VARCHAR(255) NOT NULL,  -- 实体名称
    normalized_name VARCHAR(255),  -- 规范化名称 (小写,去空格)
    description TEXT,  -- 实体描述
    aliases TEXT[],  -- 别名列表 (如: ["老妈", "母亲", "妈妈"])，用于实体消歧

    -- 属性
    attributes JSONB DEFAULT '{}',  -- 动态属性

    -- 向量
    embedding VECTOR(1536),

    -- 统计
    mention_count INT DEFAULT 1,  -- 被提及次数
    importance_score DECIMAL(5,2) DEFAULT 50.0,  -- 重要性分数
    last_mentioned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Neo4j 同步状态
    neo4j_synced BOOLEAN DEFAULT FALSE,
    neo4j_synced_at TIMESTAMP WITH TIME ZONE,
    -- v1.2 新增: 同步增强字段 (重试、幂等性、错误追踪)
    sync_version INT DEFAULT 0,  -- 同步版本号 (用于幂等性保证)
    sync_failed_count INT DEFAULT 0,  -- 同步失败次数 (超过阈值进入死信队列)
    last_sync_error TEXT,  -- 最后一次同步错误信息

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_node_type CHECK (node_type IN (
        'PERSON',        -- 人物: 家人、朋友、同事
        'PLACE',         -- 地点: 家、公司、城市
        'ORGANIZATION',  -- 组织: 公司、学校、社团
        'CONCEPT',       -- 概念: 兴趣、价值观、目标
        'EVENT',         -- 事件: 引用 important_events
        'OBJECT',        -- 物品: 宠物、收藏品
        'TIME_PERIOD',   -- 时间段: "大学时期", "去年夏天"
        'EMOTION'        -- 情感: 长期情感状态
    )),
    UNIQUE(user_id, character_id, node_type, normalized_name)
);

-- 索引
CREATE INDEX idx_graph_nodes_user_char ON memory_graph_nodes(user_id, character_id);
CREATE INDEX idx_graph_nodes_type ON memory_graph_nodes(node_type);
CREATE INDEX idx_graph_nodes_name ON memory_graph_nodes(normalized_name);
CREATE INDEX idx_graph_nodes_importance ON memory_graph_nodes(importance_score DESC);
CREATE INDEX memory_graph_nodes_embedding_idx
ON memory_graph_nodes USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- 需要同步到 Neo4j 的节点
CREATE INDEX idx_graph_nodes_sync_pending
ON memory_graph_nodes(neo4j_synced) WHERE neo4j_synced = FALSE;

CREATE TRIGGER update_memory_graph_nodes_updated_at BEFORE UPDATE ON memory_graph_nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE memory_graph_nodes IS '图记忆节点表 - 存储知识图谱节点 (与Neo4j同步)';
COMMENT ON COLUMN memory_graph_nodes.normalized_name IS '规范化名称,用于去重和匹配';
```

---

### 8.5 memory_graph_edges (图记忆边表)

PostgreSQL 端存储图边的元数据，与 Neo4j 保持同步。

```sql
CREATE TABLE memory_graph_edges (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    source_node_id BIGINT NOT NULL,  -- 源节点 memory_graph_nodes.id
    target_node_id BIGINT NOT NULL,  -- 目标节点 memory_graph_nodes.id

    -- 关系信息
    relation_type VARCHAR(100) NOT NULL,  -- 关系类型
    relation_description TEXT,  -- 关系描述

    -- 强度和置信度
    weight DECIMAL(5,2) DEFAULT 1.0 CHECK (weight >= 0),  -- 关系强度
    confidence DECIMAL(3,2) DEFAULT 0.7 CHECK (confidence BETWEEN 0 AND 1),

    -- 时效性
    valid_from TIMESTAMP WITH TIME ZONE,  -- 关系开始时间
    valid_until TIMESTAMP WITH TIME ZONE,  -- 关系结束时间 (NULL表示当前有效)

    -- 统计
    mention_count INT DEFAULT 1,  -- 被提及次数
    last_mentioned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Neo4j 同步状态
    neo4j_synced BOOLEAN DEFAULT FALSE,
    neo4j_synced_at TIMESTAMP WITH TIME ZONE,
    -- v1.2 新增: 同步增强字段 (重试、幂等性、错误追踪)
    sync_version INT DEFAULT 0,  -- 同步版本号 (用于幂等性保证)
    sync_failed_count INT DEFAULT 0,  -- 同步失败次数
    last_sync_error TEXT,  -- 最后一次同步错误信息

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_relation_type CHECK (relation_type IN (
        -- 人际关系
        'KNOWS',              -- 认识
        'FRIEND_OF',          -- 朋友
        'FAMILY_OF',          -- 家人
        'COLLEAGUE_OF',       -- 同事
        'PARTNER_OF',         -- 伴侣
        -- 归属关系
        'WORKS_AT',           -- 在...工作
        'LIVES_IN',           -- 住在
        'STUDIES_AT',         -- 在...学习
        'MEMBER_OF',          -- 是...成员
        -- 情感关系
        'LIKES',              -- 喜欢
        'DISLIKES',           -- 不喜欢
        'WORRIED_ABOUT',      -- 担心
        'EXCITED_ABOUT',      -- 期待
        -- 事件关系
        'PARTICIPATED_IN',    -- 参与了
        'HAPPENED_AT',        -- 发生在
        'CAUSED_BY',          -- 由...引起
        'RESULTED_IN',        -- 导致了
        -- 其他
        'RELATED_TO',         -- 相关
        'PART_OF',            -- 属于
        'OWNS',               -- 拥有
        'CUSTOM'              -- 自定义
    )),
    UNIQUE(source_node_id, target_node_id, relation_type)
);

-- 索引
CREATE INDEX idx_graph_edges_source ON memory_graph_edges(source_node_id);
CREATE INDEX idx_graph_edges_target ON memory_graph_edges(target_node_id);
CREATE INDEX idx_graph_edges_relation ON memory_graph_edges(relation_type);
CREATE INDEX idx_graph_edges_weight ON memory_graph_edges(weight DESC);

-- 需要同步到 Neo4j 的边
CREATE INDEX idx_graph_edges_sync_pending
ON memory_graph_edges(neo4j_synced) WHERE neo4j_synced = FALSE;

CREATE TRIGGER update_memory_graph_edges_updated_at BEFORE UPDATE ON memory_graph_edges
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE memory_graph_edges IS '图记忆边表 - 存储知识图谱关系 (与Neo4j同步)';
COMMENT ON COLUMN memory_graph_edges.valid_until IS 'NULL表示关系当前有效';
```

---

### 8.6 conversation_summaries (会话摘要表)

Layer 1 对话压缩层的数据存储。

```sql
CREATE TABLE conversation_summaries (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    conversation_id BIGINT NOT NULL,  -- 关联conversations.id
    user_id BIGINT NOT NULL,  -- 关联users.id
    character_id BIGINT NOT NULL,  -- 关联characters.id

    -- 摘要内容
    summary TEXT NOT NULL,  -- 会话摘要
    key_points JSONB DEFAULT '[]',  -- 关键点列表
    topics_discussed TEXT[],  -- 讨论话题

    -- 覆盖范围
    start_message_id BIGINT,  -- 摘要覆盖的起始消息
    end_message_id BIGINT,  -- 摘要覆盖的结束消息
    message_count INT NOT NULL,  -- 被摘要的消息数量
    original_token_count INT,  -- 原始token数
    summary_token_count INT,  -- 摘要token数

    -- 向量
    embedding VECTOR(1536),

    -- 元数据
    compression_ratio DECIMAL(5,2),  -- 压缩比
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_conv_summaries_conv ON conversation_summaries(conversation_id);
CREATE INDEX idx_conv_summaries_user_char ON conversation_summaries(user_id, character_id);
CREATE INDEX conversation_summaries_embedding_idx
ON conversation_summaries USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

COMMENT ON TABLE conversation_summaries IS '会话摘要表 - Layer 1 对话压缩层数据';
COMMENT ON COLUMN conversation_summaries.compression_ratio IS '压缩比 = summary_token_count / original_token_count';
```

---

## 9. Neo4j Graph Database (图数据库)

### 9.1 为什么选择 Neo4j

| 特性 | Neo4j | PostgreSQL 递归查询 |
|------|-------|-------------------|
| 多跳关系遍历 | O(k) 常数时间 | O(n) 线性扫描 |
| 图算法支持 | PageRank, 社区检测等内置 | 需手动实现 |
| 可视化 | Neo4j Browser 内置 | 需第三方工具 |
| 语义搜索 | 支持向量索引 | pgvector |
| 生态 | LangChain/Mem0 原生支持 | 需适配 |

### 9.2 节点 Schema (Cypher)

```cypher
// === 节点类型 ===

// 用户节点
CREATE CONSTRAINT user_id_unique IF NOT EXISTS
FOR (u:User) REQUIRE u.user_id IS UNIQUE;

// (:User) 节点属性:
// - user_id: BIGINT (PostgreSQL users.id)
// - username: STRING
// - core_summary: STRING (Mnemosyne 核心画像)
// - core_embedding: LIST<FLOAT> (1536 维向量)
// - created_at: DATETIME

// AI 角色节点
CREATE CONSTRAINT character_id_unique IF NOT EXISTS
FOR (c:Character) REQUIRE c.character_id IS UNIQUE;

// (:Character) 节点属性:
// - character_id: BIGINT (PostgreSQL characters.id)
// - name: STRING
// - created_at: DATETIME

// 实体节点 (用户提及的人/物/地点)
CREATE CONSTRAINT entity_id_unique IF NOT EXISTS
FOR (e:Entity) REQUIRE e.entity_id IS UNIQUE;

// (:Entity) 节点属性:
// - entity_id: BIGINT (PostgreSQL memory_graph_nodes.id)
// - name: STRING
// - normalized_name: STRING
// - type: STRING (PERSON, PLACE, ORGANIZATION, CONCEPT, OBJECT, TIME_PERIOD)
// - attributes: MAP
// - embedding: LIST<FLOAT>
// - mention_count: INT
// - importance_score: FLOAT
// - last_mentioned_at: DATETIME
// - created_at: DATETIME

// 事件节点
CREATE CONSTRAINT event_id_unique IF NOT EXISTS
FOR (ev:Event) REQUIRE ev.event_id IS UNIQUE;

// (:Event) 节点属性:
// - event_id: BIGINT (PostgreSQL important_events.id)
// - title: STRING
// - content: STRING
// - event_type: STRING
// - surprise_score: FLOAT
// - importance: FLOAT
// - occurred_at: DATETIME
// - temporal_context: STRING
// - embedding: LIST<FLOAT>
// - created_at: DATETIME

// 情感状态节点
CREATE CONSTRAINT emotion_state_id_unique IF NOT EXISTS
FOR (es:EmotionalState) REQUIRE es.state_id IS UNIQUE;

// (:EmotionalState) 节点属性:
// - state_id: BIGINT (PostgreSQL emotional_states.id)
// - valence: FLOAT (-1 to +1)
// - arousal: FLOAT (0 to 1)
// - primary_emotion: STRING
// - intensity: FLOAT
// - assessed_at: DATETIME
```

### 9.3 关系 Schema (Cypher)

```cypher
// === 关系类型 ===

// 用户-角色核心关系
// (:User)-[:INTERACTS_WITH]->(:Character)
// 属性:
// - closeness: FLOAT (0-100)
// - trust: FLOAT (0-100)
// - affection: FLOAT (0-100)
// - interaction_count: INT
// - total_time_minutes: INT
// - first_interaction_at: DATETIME
// - last_interaction_at: DATETIME
// - milestones: LIST<MAP>

// 用户-实体关系
// (:User)-[:KNOWS]->(:Entity:Person)
// 属性:
// - relation_type: STRING (friend, family, colleague, ...)
// - closeness: FLOAT
// - context: STRING
// - first_mentioned_at: DATETIME
// - last_mentioned_at: DATETIME

// 实体之间的关系 (任意关系类型)
// (:Entity)-[:RELATED_TO]->(:Entity)
// 属性:
// - relation_type: STRING
// - weight: FLOAT
// - description: STRING
// - valid_from: DATETIME
// - valid_until: DATETIME

// 事件参与关系
// (:User)-[:EXPERIENCED]->(:Event)
// (:Entity)-[:PARTICIPATED_IN]->(:Event)
// (:Event)-[:OCCURRED_AT]->(:Entity:Place)
// (:Event)-[:CAUSED_BY]->(:Event)
// (:Event)-[:RESULTED_IN]->(:Event)

// 情感状态关联
// (:User)-[:FELT {in_conversation: BIGINT}]->(:EmotionalState)
// (:Event)-[:TRIGGERED]->(:EmotionalState)

// 实体归属关系
// (:Entity:Person)-[:WORKS_AT]->(:Entity:Organization)
// (:Entity:Person)-[:LIVES_IN]->(:Entity:Place)
// (:Entity:Person)-[:MEMBER_OF]->(:Entity:Organization)
// (:Entity)-[:PART_OF]->(:Entity)

// 情感关系
// (:User)-[:LIKES]->(:Entity)
// (:User)-[:DISLIKES]->(:Entity)
// (:User)-[:WORRIED_ABOUT]->(:Entity)
```

### 9.4 向量索引 (Neo4j 5.x)

```cypher
// 创建实体向量索引
CREATE VECTOR INDEX entity_embeddings IF NOT EXISTS
FOR (e:Entity)
ON e.embedding
OPTIONS {indexConfig: {
    `vector.dimensions`: 1536,
    `vector.similarity_function`: 'cosine'
}};

// 创建事件向量索引
CREATE VECTOR INDEX event_embeddings IF NOT EXISTS
FOR (ev:Event)
ON ev.embedding
OPTIONS {indexConfig: {
    `vector.dimensions`: 1536,
    `vector.similarity_function`: 'cosine'
}};

// 创建用户画像向量索引
CREATE VECTOR INDEX user_core_embeddings IF NOT EXISTS
FOR (u:User)
ON u.core_embedding
OPTIONS {indexConfig: {
    `vector.dimensions`: 1536,
    `vector.similarity_function`: 'cosine'
}};
```

### 9.5 核心查询示例

```cypher
// 1. 获取用户完整知识图谱 (2跳以内)
MATCH (u:User {user_id: $user_id})-[r1*1..2]-(related)
RETURN u, r1, related
LIMIT 100;

// 2. 向量相似度检索 + 图遍历扩展上下文
CALL db.index.vector.queryNodes('entity_embeddings', 5, $query_embedding)
YIELD node, score
WHERE score > 0.7
MATCH (node)-[:RELATED_TO*1..2]-(context)
RETURN node, collect(DISTINCT context) AS related_context, score
ORDER BY score DESC;

// 3. 情感变化趋势 (最近30天)
MATCH (u:User {user_id: $user_id})-[:FELT]->(es:EmotionalState)
WHERE es.assessed_at > datetime() - duration('P30D')
RETURN es.primary_emotion, es.valence, es.arousal, es.assessed_at
ORDER BY es.assessed_at;

// 4. 关系里程碑查询
MATCH (u:User {user_id: $user_id})-[r:INTERACTS_WITH]->(c:Character {character_id: $char_id})
RETURN r.milestones, r.closeness, r.trust, r.affection;

// 5. 惊讶度最高的事件 (Mnemosyne)
MATCH (u:User {user_id: $user_id})-[:EXPERIENCED]->(e:Event)
RETURN e.title, e.surprise_score, e.importance, e.occurred_at, e.temporal_context
ORDER BY e.surprise_score DESC
LIMIT 10;

// 6. 查找用户社交网络
MATCH (u:User {user_id: $user_id})-[:KNOWS]->(p:Entity:Person)
OPTIONAL MATCH (p)-[r:RELATED_TO]->(other:Entity)
RETURN p.name, p.attributes, collect({relation: type(r), target: other.name}) AS connections;

// 7. 事件因果链追溯
MATCH path = (e1:Event)-[:CAUSED_BY|RESULTED_IN*1..3]-(e2:Event)
WHERE e1.event_id = $event_id
RETURN path;

// 8. 综合记忆检索 (混合评分)
MATCH (u:User {user_id: $user_id})-[:EXPERIENCED]->(e:Event)
WITH e,
     e.surprise_score * 0.3 +
     e.importance / 100 * 0.3 +
     1.0 / (1 + duration.inDays(datetime(), e.occurred_at).days * 0.01) * 0.2 +
     size((e)-[:PARTICIPATED_IN]-()) / 10.0 * 0.2 AS combined_score
RETURN e.title, e.content, combined_score
ORDER BY combined_score DESC
LIMIT 10;
```

---

## 10. PostgreSQL-Neo4j Sync Strategy (数据同步策略)

### 10.1 架构概述

```
PostgreSQL (Source of Truth)              Neo4j (Graph Query Optimizer)
├── users                   ─────────────→  (:User)
├── characters              ─────────────→  (:Character)
├── relationships           ←────────────→  [:INTERACTS_WITH]
├── memories               ─────────────→  (:Entity), (:Event) [部分]
├── emotional_states        ─────────────→  (:EmotionalState)
├── important_events        ─────────────→  (:Event)
├── memory_graph_nodes      ←────────────→  (:Entity)
├── memory_graph_edges      ←────────────→  [各种关系]
└── user_portraits          ─────────────→  (:User).core_*

同步方向:
  ─────────────→  单向同步 (PG → Neo4j)
  ←────────────→  双向同步 (写PG, 读Neo4j, 异步同步)
```

### 10.2 同步机制

**策略**: 事件驱动 + 批量同步

```python
# 同步服务伪代码
class Neo4jSyncService:
    """PostgreSQL → Neo4j 同步服务"""

    def __init__(self, pg_pool, neo4j_driver, redis_client):
        self.pg = pg_pool
        self.neo4j = neo4j_driver
        self.redis = redis_client  # 用于事件队列

    async def handle_node_created(self, node_data: dict):
        """处理节点创建事件"""
        cypher = """
        MERGE (e:Entity {entity_id: $entity_id})
        SET e.name = $name,
            e.normalized_name = $normalized_name,
            e.type = $type,
            e.attributes = $attributes,
            e.embedding = $embedding,
            e.mention_count = $mention_count,
            e.importance_score = $importance_score,
            e.created_at = datetime()
        """
        await self.neo4j.execute_query(cypher, node_data)

        # 更新 PostgreSQL 同步状态
        await self.pg.execute("""
            UPDATE memory_graph_nodes
            SET neo4j_synced = TRUE, neo4j_synced_at = NOW()
            WHERE id = $1
        """, node_data['entity_id'])

    async def handle_edge_created(self, edge_data: dict):
        """处理边创建事件"""
        cypher = """
        MATCH (source:Entity {entity_id: $source_id})
        MATCH (target:Entity {entity_id: $target_id})
        MERGE (source)-[r:RELATED_TO {relation_type: $relation_type}]->(target)
        SET r.weight = $weight,
            r.confidence = $confidence,
            r.description = $description,
            r.valid_from = $valid_from,
            r.valid_until = $valid_until,
            r.updated_at = datetime()
        """
        await self.neo4j.execute_query(cypher, edge_data)

    async def batch_sync_pending(self, batch_size: int = 100):
        """批量同步待处理数据"""
        # 同步待同步节点
        nodes = await self.pg.fetch("""
            SELECT * FROM memory_graph_nodes
            WHERE neo4j_synced = FALSE
            LIMIT $1
        """, batch_size)

        for node in nodes:
            await self.handle_node_created(dict(node))

        # 同步待同步边
        edges = await self.pg.fetch("""
            SELECT * FROM memory_graph_edges
            WHERE neo4j_synced = FALSE
            LIMIT $1
        """, batch_size)

        for edge in edges:
            await self.handle_edge_created(dict(edge))
```

### 10.3 读写分离策略

| 操作 | 数据源 | 说明 |
|------|--------|------|
| 创建记忆 | PostgreSQL | 写入 PG，异步同步到 Neo4j |
| 简单查询 | PostgreSQL | 单表查询直接走 PG |
| 向量检索 | PostgreSQL (pgvector) | Top-K 相似度查询 |
| 图遍历 | Neo4j | 多跳关系查询 |
| 混合检索 | PG → Neo4j | 先 PG 向量检索，再 Neo4j 扩展上下文 |
| 关系分析 | Neo4j | PageRank、社区检测等图算法 |

### 10.4 Go 同步服务实现

```go
package sync

import (
    "context"
    "encoding/json"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
    "github.com/redis/go-redis/v9"
)

type Neo4jSyncService struct {
    pg    *pgxpool.Pool
    neo4j neo4j.DriverWithContext
    redis *redis.Client
}

func NewNeo4jSyncService(pg *pgxpool.Pool, neo4jDriver neo4j.DriverWithContext, redis *redis.Client) *Neo4jSyncService {
    return &Neo4jSyncService{pg: pg, neo4j: neo4jDriver, redis: redis}
}

// SyncNode 同步单个节点到 Neo4j
func (s *Neo4jSyncService) SyncNode(ctx context.Context, nodeID int64) error {
    // 1. 从 PostgreSQL 读取节点数据
    var node struct {
        ID             int64
        UserID         int64
        CharacterID    int64
        NodeType       string
        Name           string
        NormalizedName string
        Description    *string
        Attributes     json.RawMessage
        Embedding      []float32
        MentionCount   int
        ImportanceScore float64
    }

    err := s.pg.QueryRow(ctx, `
        SELECT id, user_id, character_id, node_type, name, normalized_name,
               description, attributes, embedding, mention_count, importance_score
        FROM memory_graph_nodes WHERE id = $1
    `, nodeID).Scan(
        &node.ID, &node.UserID, &node.CharacterID, &node.NodeType,
        &node.Name, &node.NormalizedName, &node.Description,
        &node.Attributes, &node.Embedding, &node.MentionCount, &node.ImportanceScore,
    )
    if err != nil {
        return err
    }

    // 2. 写入 Neo4j
    session := s.neo4j.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
    defer session.Close(ctx)

    _, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
        _, err := tx.Run(ctx, `
            MERGE (e:Entity {entity_id: $entity_id})
            SET e.user_id = $user_id,
                e.character_id = $character_id,
                e.name = $name,
                e.normalized_name = $normalized_name,
                e.type = $type,
                e.description = $description,
                e.embedding = $embedding,
                e.mention_count = $mention_count,
                e.importance_score = $importance_score,
                e.updated_at = datetime()
        `, map[string]any{
            "entity_id":        node.ID,
            "user_id":          node.UserID,
            "character_id":     node.CharacterID,
            "name":             node.Name,
            "normalized_name":  node.NormalizedName,
            "type":             node.NodeType,
            "description":      node.Description,
            "embedding":        node.Embedding,
            "mention_count":    node.MentionCount,
            "importance_score": node.ImportanceScore,
        })
        return nil, err
    })
    if err != nil {
        return err
    }

    // 3. 更新 PostgreSQL 同步状态
    _, err = s.pg.Exec(ctx, `
        UPDATE memory_graph_nodes
        SET neo4j_synced = TRUE, neo4j_synced_at = NOW()
        WHERE id = $1
    `, nodeID)

    return err
}

// BatchSyncPending 批量同步待处理数据
func (s *Neo4jSyncService) BatchSyncPending(ctx context.Context, batchSize int) (int, error) {
    rows, err := s.pg.Query(ctx, `
        SELECT id FROM memory_graph_nodes
        WHERE neo4j_synced = FALSE
        LIMIT $1
    `, batchSize)
    if err != nil {
        return 0, err
    }
    defer rows.Close()

    synced := 0
    for rows.Next() {
        var nodeID int64
        if err := rows.Scan(&nodeID); err != nil {
            continue
        }
        if err := s.SyncNode(ctx, nodeID); err == nil {
            synced++
        }
    }

    return synced, nil
}
```

### 10.5 定时同步任务

```yaml
# docker-compose.yml 新增
services:
  neo4j-sync-worker:
    build:
      context: ./backend/golang
      dockerfile: Dockerfile.sync
    environment:
      - POSTGRES_DSN=postgres://...
      - NEO4J_URI=bolt://neo4j:7687
      - NEO4J_USER=neo4j
      - NEO4J_PASSWORD=xxx
      - REDIS_ADDR=redis:6379
      - SYNC_INTERVAL=30s
      - BATCH_SIZE=100
    depends_on:
      - postgres
      - neo4j
      - redis
```

---

## Summary

本数据模型设计涵盖了AI陪伴平台的所有核心业务:

- **19个核心表**: 用户、角色、对话、记忆、积分等 (v1.1 新增 6 个长期记忆表)
- **向量检索**: pgvector HNSW索引,支持<100ms查询
- **图数据库**: Neo4j 5.x 支持复杂关系遍历和图算法
- **五层记忆架构**: 对话压缩、情感追踪、重要事件、关系图谱、核心身份
- **分区策略**: 大表按月分区,支持数据归档
- **幂等性**: idempotency_key保证关键操作不重复
- **软删除**: 重要数据(用户、角色)支持30天恢复期
- **数据同步**: PostgreSQL-Neo4j 事件驱动同步

### v1.1 新增表

| 表名 | 用途 | 架构层 |
|------|------|--------|
| user_portraits | 用户画像 (PRIME 语义记忆) | Layer 5 |
| emotional_states | 情感状态时序数据 (Russell 模型) | Layer 2 |
| important_events | 重要事件 (Mnemosyne 惊讶度) | Layer 3 |
| memory_graph_nodes | 图记忆节点 (Neo4j 同步) | Layer 4 |
| memory_graph_edges | 图记忆边 (Neo4j 同步) | Layer 4 |
| conversation_summaries | 会话摘要 (对话压缩) | Layer 1 |

### v1.1 扩展字段

| 表名 | 新增字段 |
|------|----------|
| memories | surprise_score, decay_factor, boost_count, source_type, connectivity |
| relationships | trust_score, affection_score, emotional_history, milestones, communication_patterns, first_interaction_at |

**Next Steps**:
1. 生成 contracts/(Protobuf API定义)
2. 生成 quickstart.md(本地开发环境搭建)
3. 执行 `/speckit.tasks` 生成任务清单
4. 添加 Neo4j Docker 配置

---

**Document Version**: 1.1
**Last Updated**: 2025-12-27
**Status**: ✅ Ready for Implementation
