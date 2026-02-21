-- 008_messages_client_message_id.sql
-- Schema drift hardening:
-- 1) messages/client_message_id + outbox 基础
-- 2) sessions/users/conversations/messages/characters 结构补齐
-- 3) credit 预扣字段预埋（兼容旧代码）
-- 4) 索引冗余清理

BEGIN;

-- ================================================================
-- 0. Extensions
-- ================================================================
CREATE EXTENSION IF NOT EXISTS vector;

-- ================================================================
-- 1. users / user_profiles 索引与约束整理
-- ================================================================
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_phone_key'
          AND conrelid = 'users'::regclass
    ) THEN
        ALTER TABLE users DROP CONSTRAINT users_phone_key;
    END IF;
END
$$;

CREATE UNIQUE INDEX IF NOT EXISTS uk_users_phone_active
    ON users(phone)
    WHERE phone IS NOT NULL AND NOT is_deleted;

DROP INDEX IF EXISTS idx_users_email_lower;
DROP INDEX IF EXISTS idx_user_profiles_user_id;

-- ================================================================
-- 2. sessions 补齐字段与索引
-- ================================================================
ALTER TABLE sessions
    ADD COLUMN IF NOT EXISTS device_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'sessions'
          AND column_name = 'ip_address'
          AND udt_name <> 'inet'
    ) THEN
        CREATE OR REPLACE FUNCTION __tmp_try_cast_inet(value TEXT)
        RETURNS INET
        LANGUAGE plpgsql
        AS $fn$
        BEGIN
            IF value IS NULL OR btrim(value) = '' THEN
                RETURN NULL;
            END IF;
            BEGIN
                RETURN value::INET;
            EXCEPTION WHEN OTHERS THEN
                RETURN NULL;
            END;
        END;
        $fn$;

        ALTER TABLE sessions
            ALTER COLUMN ip_address TYPE INET
            USING __tmp_try_cast_inet(ip_address);

        DROP FUNCTION __tmp_try_cast_inet(TEXT);
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_sessions_expires_active
    ON sessions(expires_at)
    WHERE revoked_at IS NULL;

-- ================================================================
-- 3. characters world_view 约束补齐
-- ================================================================
UPDATE characters
SET world_view = '{}'::jsonb
WHERE world_view IS NULL;

ALTER TABLE characters
    ALTER COLUMN world_view SET DEFAULT '{}'::jsonb,
    ALTER COLUMN world_view SET NOT NULL;

-- ================================================================
-- 4. conversations/messages 字段补齐（兼容现有实现）
-- ================================================================
ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

UPDATE conversations SET message_count = 0 WHERE message_count IS NULL;
UPDATE conversations SET token_count = 0 WHERE token_count IS NULL;
UPDATE conversations SET started_at = CURRENT_TIMESTAMP WHERE started_at IS NULL;
UPDATE conversations SET last_message_at = started_at WHERE last_message_at IS NULL;
UPDATE conversations SET is_archived = FALSE WHERE is_archived IS NULL;

ALTER TABLE conversations
    ALTER COLUMN message_count SET DEFAULT 0,
    ALTER COLUMN message_count SET NOT NULL,
    ALTER COLUMN token_count SET DEFAULT 0,
    ALTER COLUMN token_count SET NOT NULL,
    ALTER COLUMN started_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN started_at SET NOT NULL,
    ALTER COLUMN last_message_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN last_message_at SET NOT NULL,
    ALTER COLUMN is_archived SET DEFAULT FALSE,
    ALTER COLUMN is_archived SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_conversations_user_character_last
    ON conversations(user_id, character_id, last_message_at DESC);

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS client_message_id VARCHAR(64),
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

UPDATE messages
SET client_message_id = 'legacy-' || id::TEXT
WHERE client_message_id IS NULL;

UPDATE messages
SET created_at = CURRENT_TIMESTAMP
WHERE created_at IS NULL;

ALTER TABLE messages
    ALTER COLUMN client_message_id SET NOT NULL,
    ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
    ALTER COLUMN created_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_messages_client_message_id ON messages(client_message_id);

-- ================================================================
-- 5. 消息幂等键表 + outbox 表
-- ================================================================
CREATE TABLE IF NOT EXISTS conversation_partition_keys (
    conversation_id BIGINT PRIMARY KEY,
    started_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conversation_partition_keys_started_at
    ON conversation_partition_keys(started_at);

INSERT INTO conversation_partition_keys(conversation_id, started_at)
SELECT c.id, c.started_at
FROM conversations c
ON CONFLICT (conversation_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS message_partition_keys (
    message_id BIGINT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL,
    conversation_id BIGINT,
    created_on TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_message_partition_keys_created_at
    ON message_partition_keys(created_at);

INSERT INTO message_partition_keys(message_id, created_at, conversation_id)
SELECT m.id, m.created_at, m.conversation_id
FROM messages m
ON CONFLICT (message_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS message_dedup_keys (
    client_message_id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    conversation_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    message_created_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'RESERVED',
    reserved_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finalized_at TIMESTAMPTZ,
    ttl_at TIMESTAMPTZ,
    metadata JSONB,
    CONSTRAINT chk_message_dedup_status CHECK (status IN ('RESERVED', 'FINALIZED', 'FAILED', 'EXPIRED'))
);

CREATE INDEX IF NOT EXISTS idx_message_dedup_ttl
    ON message_dedup_keys(ttl_at)
    WHERE status IN ('FAILED', 'EXPIRED');

CREATE INDEX IF NOT EXISTS idx_message_dedup_message_partition_lookup
    ON message_dedup_keys(message_id, message_created_at)
    WHERE message_created_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS outbox_events (
    id BIGINT PRIMARY KEY,
    aggregate_type VARCHAR(64) NOT NULL,
    aggregate_id BIGINT NOT NULL,
    event_type VARCHAR(128) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    max_retries INT NOT NULL DEFAULT 16,
    next_retry_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_outbox_status CHECK (status IN ('PENDING', 'PUBLISHED', 'FAILED', 'DEAD'))
);

CREATE INDEX IF NOT EXISTS idx_outbox_pending
    ON outbox_events(status, next_retry_at, created_at);

CREATE INDEX IF NOT EXISTS idx_outbox_aggregate
    ON outbox_events(aggregate_type, aggregate_id);

-- ================================================================
-- 6. billing 结构预埋（与旧代码兼容）
-- ================================================================
ALTER TABLE credit_accounts
    ADD COLUMN IF NOT EXISTS reserved_balance BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 0;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_credit_accounts_reserved_non_negative'
          AND conrelid = 'credit_accounts'::regclass
    ) THEN
        ALTER TABLE credit_accounts
            ADD CONSTRAINT chk_credit_accounts_reserved_non_negative
            CHECK (reserved_balance >= 0);
    END IF;
END
$$;

ALTER TABLE credit_transactions
    ADD COLUMN IF NOT EXISTS transaction_type VARCHAR(20),
    ADD COLUMN IF NOT EXISTS message_id BIGINT,
    ADD COLUMN IF NOT EXISTS reserve_id BIGINT,
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'SUCCESS',
    ADD COLUMN IF NOT EXISTS metadata JSONB;

UPDATE credit_transactions
SET transaction_type = CASE lower(type)
    WHEN 'bonus' THEN 'GRANT'
    WHEN 'consume' THEN 'SETTLE'
    WHEN 'charge' THEN 'RECHARGE'
    WHEN 'refund' THEN 'REFUND'
    ELSE 'ADJUST'
END
WHERE transaction_type IS NULL;

ALTER TABLE credit_transactions
    ALTER COLUMN transaction_type SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_credit_transactions_transaction_type'
          AND conrelid = 'credit_transactions'::regclass
    ) THEN
        ALTER TABLE credit_transactions
            ADD CONSTRAINT chk_credit_transactions_transaction_type
            CHECK (transaction_type IN ('GRANT','RESERVE','SETTLE','RELEASE','RECHARGE','REFUND','ADJUST'));
    END IF;
END
$$;

CREATE INDEX IF NOT EXISTS idx_credit_transactions_reserve_id
    ON credit_transactions(reserve_id)
    WHERE reserve_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS credit_packages (
    id BIGINT PRIMARY KEY,
    package_code VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    credits BIGINT NOT NULL,
    bonus_credits BIGINT NOT NULL DEFAULT 0,
    price_usd NUMERIC(10,2) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    sort_order INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_credit_packages_credits CHECK (credits > 0),
    CONSTRAINT chk_credit_packages_bonus CHECK (bonus_credits >= 0),
    CONSTRAINT chk_credit_packages_price CHECK (price_usd >= 0)
);

CREATE TABLE IF NOT EXISTS recharge_orders (
    id BIGINT PRIMARY KEY,
    order_no VARCHAR(64) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    package_id BIGINT,
    credits BIGINT NOT NULL,
    amount_usd NUMERIC(10,2) NOT NULL,
    payment_provider VARCHAR(32),
    payment_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    idempotency_key VARCHAR(128) UNIQUE,
    paid_at TIMESTAMPTZ,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_recharge_orders_status CHECK (payment_status IN ('PENDING','PAID','FAILED','CANCELLED')),
    CONSTRAINT chk_recharge_orders_credits CHECK (credits > 0),
    CONSTRAINT chk_recharge_orders_amount CHECK (amount_usd >= 0)
);

CREATE INDEX IF NOT EXISTS idx_recharge_orders_user_time
    ON recharge_orders(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_recharge_orders_status
    ON recharge_orders(payment_status, created_at DESC);

-- ================================================================
-- 7. 分区维护辅助函数（需由外部调度周期执行）
-- ================================================================
CREATE OR REPLACE FUNCTION ensure_next_month_partitions(months_ahead INT DEFAULT 12)
RETURNS VOID AS $$
DECLARE
    base_date DATE := date_trunc('month', CURRENT_DATE);
    i INT;
    target DATE;
BEGIN
    IF months_ahead < 1 THEN
        months_ahead := 1;
    END IF;

    FOR i IN 0..(months_ahead - 1) LOOP
        target := (base_date + (i || ' month')::interval)::date;
        PERFORM create_monthly_partition('conversations', EXTRACT(YEAR FROM target)::INT, EXTRACT(MONTH FROM target)::INT);
        PERFORM create_monthly_partition('messages', EXTRACT(YEAR FROM target)::INT, EXTRACT(MONTH FROM target)::INT);
    END LOOP;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION ensure_next_month_partitions(INT) IS 'Ensure monthly partitions for conversations/messages; run via scheduler';

COMMIT;
