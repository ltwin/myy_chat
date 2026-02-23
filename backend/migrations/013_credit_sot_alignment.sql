-- 013_credit_sot_alignment.sql
-- Credit schema alignment (development phase, no backward compatibility):
-- 1) switch to canonical column names and semantics
-- 2) remove legacy compatibility columns/usages

BEGIN;

-- ================================================================
-- 1. credit_accounts alignment
-- ================================================================
DO $$
BEGIN
    -- If a previous dev migration created total_recharged as generated column,
    -- drop it first and then rename canonical storage column from total_charged.
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'credit_accounts'
          AND column_name = 'total_recharged'
          AND is_generated = 'ALWAYS'
    ) AND EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'credit_accounts'
          AND column_name = 'total_charged'
    ) THEN
        ALTER TABLE credit_accounts DROP COLUMN total_recharged;
    END IF;

    -- Rename legacy total_charged -> total_recharged.
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'credit_accounts'
          AND column_name = 'total_charged'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'credit_accounts'
          AND column_name = 'total_recharged'
    ) THEN
        ALTER TABLE credit_accounts
            RENAME COLUMN total_charged TO total_recharged;
    END IF;

    -- If total_recharged is missing entirely, create it.
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'credit_accounts'
          AND column_name = 'total_recharged'
    ) THEN
        ALTER TABLE credit_accounts
            ADD COLUMN total_recharged BIGINT NOT NULL DEFAULT 0;
    END IF;
END
$$;

UPDATE credit_accounts SET balance = 0 WHERE balance IS NULL;
UPDATE credit_accounts SET total_recharged = 0 WHERE total_recharged IS NULL;
UPDATE credit_accounts SET total_consumed = 0 WHERE total_consumed IS NULL;
UPDATE credit_accounts SET reserved_balance = 0 WHERE reserved_balance IS NULL;
UPDATE credit_accounts SET version = 0 WHERE version IS NULL;

ALTER TABLE credit_accounts
    ALTER COLUMN balance SET DEFAULT 0,
    ALTER COLUMN balance SET NOT NULL,
    ALTER COLUMN total_recharged SET DEFAULT 0,
    ALTER COLUMN total_recharged SET NOT NULL,
    ALTER COLUMN total_consumed SET DEFAULT 0,
    ALTER COLUMN total_consumed SET NOT NULL,
    ALTER COLUMN reserved_balance SET DEFAULT 0,
    ALTER COLUMN reserved_balance SET NOT NULL,
    ALTER COLUMN version SET DEFAULT 0,
    ALTER COLUMN version SET NOT NULL;

-- ================================================================
-- 2. credit_transactions alignment
-- ================================================================
-- Ensure transaction_type exists.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'credit_transactions'
          AND column_name = 'transaction_type'
    ) THEN
        ALTER TABLE credit_transactions
            ADD COLUMN transaction_type VARCHAR(20);
    END IF;
END
$$;

-- Fill canonical transaction_type from legacy type if needed.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'credit_transactions'
          AND column_name = 'type'
    ) THEN
        EXECUTE $sql$
            UPDATE credit_transactions
            SET transaction_type = CASE lower(type)
                WHEN 'bonus' THEN 'GRANT'
                WHEN 'consume' THEN 'SETTLE'
                WHEN 'charge' THEN 'RECHARGE'
                WHEN 'refund' THEN 'REFUND'
                ELSE 'ADJUST'
            END
            WHERE transaction_type IS NULL
              AND type IS NOT NULL
        $sql$;
    END IF;
END
$$;

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'credit_transactions'
          AND column_name = 'reference_id'
          AND udt_name <> 'varchar'
    ) THEN
        ALTER TABLE credit_transactions
            ALTER COLUMN reference_id TYPE VARCHAR(128)
            USING reference_id::TEXT;
    END IF;
END
$$;

UPDATE credit_transactions
SET idempotency_key = 'legacy-' || id::TEXT
WHERE idempotency_key IS NULL;

UPDATE credit_transactions
SET transaction_type = 'ADJUST'
WHERE transaction_type IS NULL;

ALTER TABLE credit_transactions
    ALTER COLUMN transaction_type SET NOT NULL,
    ALTER COLUMN idempotency_key SET NOT NULL;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'chk_credit_transactions_status'
          AND conrelid = 'credit_transactions'::regclass
    ) THEN
        ALTER TABLE credit_transactions
            ADD CONSTRAINT chk_credit_transactions_status
            CHECK (status IN ('SUCCESS', 'PENDING', 'FAILED', 'CANCELLED', 'REVERSED'));
    END IF;
END
$$;

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

DROP INDEX IF EXISTS idx_credit_transactions_type;
DROP INDEX IF EXISTS idx_credit_transactions_type_time;

CREATE INDEX IF NOT EXISTS idx_credit_transactions_type_time
    ON credit_transactions(transaction_type, created_at DESC);

-- Drop legacy type column in development phase.
ALTER TABLE credit_transactions
    DROP COLUMN IF EXISTS type;

COMMIT;
