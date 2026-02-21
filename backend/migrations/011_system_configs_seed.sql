-- 011_system_configs_seed.sql
-- Runtime config table + bootstrap seed.

BEGIN;

CREATE TABLE IF NOT EXISTS system_configs (
    config_key VARCHAR(128) PRIMARY KEY,
    config_value JSONB NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    updated_by BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO system_configs(config_key, config_value, description, is_public)
VALUES
('billing.credit_formula', '{"usd_to_credit": 100}'::jsonb, 'USD 到积分换算', false),
('memory.compression.threshold', '{"token_threshold": 6000}'::jsonb, '上下文压缩阈值', false),
('memory.decay.default_factor', '{"daily_decay": 0.98}'::jsonb, '默认遗忘衰减系数', false)
ON CONFLICT (config_key) DO UPDATE
SET config_value = EXCLUDED.config_value,
    description = EXCLUDED.description,
    updated_at = CURRENT_TIMESTAMP;

COMMIT;
