-- 007_auth_and_credit_hardening.sql
-- 认证与积分模型加固：
-- 1) 邮箱统一 lower-case + 大小写不敏感唯一索引
-- 2) 积分字段统一为 BIGINT

BEGIN;

-- 统一邮箱归一化，避免大小写导致登录/唯一性漂移
UPDATE users
SET email = lower(email)
WHERE email <> lower(email);

-- 删除历史大小写敏感唯一约束，改为 lower(email) 唯一索引（活跃账号）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'users_email_key'
    ) THEN
        ALTER TABLE users DROP CONSTRAINT users_email_key;
    END IF;
END
$$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM users
        WHERE NOT is_deleted
        GROUP BY lower(email)
        HAVING count(*) > 1
    ) THEN
        RAISE EXCEPTION 'duplicate active emails detected after normalization';
    END IF;
END
$$;

DROP INDEX IF EXISTS idx_users_email;
CREATE UNIQUE INDEX IF NOT EXISTS uk_users_email_lower_active
    ON users(lower(email))
    WHERE NOT is_deleted;
CREATE INDEX IF NOT EXISTS idx_users_email_lower
    ON users(lower(email))
    WHERE NOT is_deleted;

-- 积分字段从 DECIMAL 收敛到 BIGINT（MVP 积分单位）
ALTER TABLE credit_accounts
    ALTER COLUMN balance TYPE BIGINT USING ROUND(balance)::BIGINT,
    ALTER COLUMN total_charged TYPE BIGINT USING ROUND(total_charged)::BIGINT,
    ALTER COLUMN total_consumed TYPE BIGINT USING ROUND(total_consumed)::BIGINT;

ALTER TABLE credit_transactions
    ALTER COLUMN amount TYPE BIGINT USING ROUND(amount)::BIGINT,
    ALTER COLUMN balance_after TYPE BIGINT USING ROUND(balance_after)::BIGINT;

COMMIT;
