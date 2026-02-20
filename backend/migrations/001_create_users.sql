-- 001_create_users.sql
-- 用户表和用户画像表创建脚本
-- 执行时机: Phase 3 开始前

-- ================================================
-- 1. 创建用户表
-- ================================================

CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,  -- 雪花ID,由应用层生成
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    phone VARCHAR(20) UNIQUE,
    avatar_url TEXT,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE,
    deletion_scheduled_at TIMESTAMP WITH TIME ZONE,  -- 冷静期开始时间
    CONSTRAINT check_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT check_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$')
);

-- 索引
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_email_lower_active ON users(lower(email)) WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_users_email_lower ON users(lower(email)) WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_users_deletion_scheduled ON users(deletion_scheduled_at)
    WHERE deletion_scheduled_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- 触发器: 自动更新updated_at（幂等）
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE users IS '用户表 - 存储用户基本信息';
COMMENT ON COLUMN users.id IS '雪花ID主键,由应用层生成';
COMMENT ON COLUMN users.deletion_scheduled_at IS '账号删除计划时间,设置后30天自动删除';

-- ================================================
-- 2. 创建用户画像表
-- ================================================

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id BIGINT PRIMARY KEY,  -- 关联users.id,由应用层维护
    full_name VARCHAR(100),
    gender VARCHAR(20),  -- male, female, other
    birth_date DATE,
    location JSONB,  -- {"country": "China", "city": "Shanghai", "province": "Shanghai"}
    interests TEXT[],  -- 兴趣爱好数组 (用户手动填写)
    occupation VARCHAR(100),
    bio TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_profiles(user_id);

DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON user_profiles;
CREATE TRIGGER update_user_profiles_updated_at
    BEFORE UPDATE ON user_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE user_profiles IS '用户画像表 - 存储用户个人详细信息';
COMMENT ON COLUMN user_profiles.user_id IS '关联users.id（应用层维护外键关系）';

-- ================================================
-- 3. 创建积分账户表
-- ================================================

CREATE TABLE IF NOT EXISTS credit_accounts (
    user_id BIGINT PRIMARY KEY,  -- 关联users.id
    balance BIGINT DEFAULT 0 CHECK (balance >= 0),
    total_charged BIGINT DEFAULT 0,  -- 累计充值
    total_consumed BIGINT DEFAULT 0,  -- 累计消费
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

DROP TRIGGER IF EXISTS update_credit_accounts_updated_at ON credit_accounts;
CREATE TRIGGER update_credit_accounts_updated_at
    BEFORE UPDATE ON credit_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE credit_accounts IS '积分账户表 - 存储用户积分余额';

-- ================================================
-- 4. 创建积分交易记录表
-- ================================================

CREATE TABLE IF NOT EXISTS credit_transactions (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,  -- charge, consume, refund, bonus
    amount BIGINT NOT NULL,  -- 正数为增加，负数为减少
    balance_after BIGINT NOT NULL,  -- 交易后余额
    description TEXT,
    reference_type VARCHAR(50),  -- order, llm_call, admin
    reference_id BIGINT,  -- 关联的订单ID或其他
    idempotency_key VARCHAR(100) UNIQUE,  -- 幂等键
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_credit_transactions_user_id ON credit_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_type ON credit_transactions(type);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_created_at ON credit_transactions(created_at);

COMMENT ON TABLE credit_transactions IS '积分交易记录表 - 记录所有积分变动';
COMMENT ON COLUMN credit_transactions.idempotency_key IS '幂等键，防止重复扣费';

-- ================================================
-- 5. 创建 Session 表 (用于刷新令牌管理)
-- ================================================

CREATE TABLE IF NOT EXISTS sessions (
    id BIGINT PRIMARY KEY,  -- 雪花ID
    user_id BIGINT NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL,
    user_agent TEXT,
    ip_address VARCHAR(45),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_token ON sessions(refresh_token_hash)
    WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

COMMENT ON TABLE sessions IS '会话表 - 存储用户刷新令牌';

-- ================================================
-- 输出完成信息
-- ================================================

DO $$
BEGIN
    RAISE NOTICE '======================================';
    RAISE NOTICE '用户相关表创建完成';
    RAISE NOTICE '- users: 用户基本信息';
    RAISE NOTICE '- user_profiles: 用户画像';
    RAISE NOTICE '- credit_accounts: 积分账户';
    RAISE NOTICE '- credit_transactions: 积分交易记录';
    RAISE NOTICE '- sessions: 会话管理';
    RAISE NOTICE '======================================';
END
$$;
