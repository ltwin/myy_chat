-- PostgreSQL 初始化脚本
-- 用于创建数据库、扩展和初始配置
-- 由 docker-compose 在首次启动时执行

-- ================================================
-- 1. 创建数据库扩展
-- ================================================

-- pgvector 扩展: 用于向量存储和相似度搜索
CREATE EXTENSION IF NOT EXISTS vector;

-- uuid-ossp 扩展: UUID 生成（备用，主要使用 Snowflake ID）
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- pg_trgm 扩展: 三元组索引，用于模糊搜索
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ================================================
-- 2. 创建通用触发器函数
-- ================================================

-- 自动更新 updated_at 字段的触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS '自动更新 updated_at 字段为当前时间';

-- ================================================
-- 3. 创建系统配置表
-- ================================================

-- 系统配置表（用于存储全局配置，如积分转换率等）
CREATE TABLE IF NOT EXISTS system_configs (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_system_configs_updated_at
    BEFORE UPDATE ON system_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE system_configs IS '系统配置表 - 存储全局配置项';

-- ================================================
-- 4. 初始化系统配置数据
-- ================================================

-- 积分转换配置 (USD to Credits)
-- rate: 1 USD = 100 积分基础
-- model_multipliers: 不同模型的成本倍率
-- platform_margin: 平台利润率 (1.2 = 20% 利润)
INSERT INTO system_configs (key, value, description) VALUES
(
    'credit_price_mapping',
    '{
        "base_rate": 100,
        "currency": "USD",
        "model_multipliers": {
            "gpt-4o": 1.0,
            "gpt-4o-mini": 0.3,
            "claude-3-5-sonnet": 1.2,
            "claude-3-opus": 3.0,
            "gemini-2.0-flash": 0.2
        },
        "platform_margin": 1.2,
        "updated_by": "system_init"
    }'::jsonb,
    'USD到积分的转换配置：base_rate=1USD对应积分数，model_multipliers=模型成本倍率，platform_margin=平台利润率'
)
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = CURRENT_TIMESTAMP;

-- ================================================
-- 5. 创建积分套餐表
-- ================================================

CREATE TABLE IF NOT EXISTS credit_packages (
    id BIGINT PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    credits DECIMAL(12,2) NOT NULL CHECK (credits > 0),
    price_cny DECIMAL(10,2) NOT NULL CHECK (price_cny > 0),
    bonus_credits DECIMAL(12,2) DEFAULT 0,  -- 赠送积分
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER update_credit_packages_updated_at
    BEFORE UPDATE ON credit_packages
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE credit_packages IS '积分套餐表 - 预设的充值套餐';

-- 初始化积分套餐数据 (5个预设套餐)
-- 使用固定的 Snowflake ID 便于引用
INSERT INTO credit_packages (id, name, description, credits, price_cny, bonus_credits, sort_order) VALUES
(1000000000000001, '体验包', '适合新用户体验', 100, 10.00, 0, 1),
(1000000000000002, '基础包', '日常对话使用', 500, 45.00, 50, 2),
(1000000000000003, '标准包', '高频用户推荐', 1000, 80.00, 150, 3),
(1000000000000004, '豪华包', '深度用户首选', 3000, 220.00, 600, 4),
(1000000000000005, '尊享包', 'VIP用户专属', 10000, 680.00, 2500, 5)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    credits = EXCLUDED.credits,
    price_cny = EXCLUDED.price_cny,
    bonus_credits = EXCLUDED.bonus_credits,
    sort_order = EXCLUDED.sort_order,
    updated_at = CURRENT_TIMESTAMP;

-- ================================================
-- 6. 创建月度分区自动创建函数
-- ================================================

-- 自动创建下一个月的分区表
CREATE OR REPLACE FUNCTION create_monthly_partition(
    p_table_name TEXT,
    p_year INT,
    p_month INT
)
RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    partition_name := p_table_name || '_' || p_year || '_' || LPAD(p_month::TEXT, 2, '0');
    start_date := make_date(p_year, p_month, 1);
    end_date := start_date + INTERVAL '1 month';

    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        p_table_name,
        start_date,
        end_date
    );

    RAISE NOTICE 'Created partition: %', partition_name;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_monthly_partition(TEXT, INT, INT) IS '自动创建指定月份的分区表';

-- ================================================
-- 7. 权限设置
-- ================================================

-- 创建应用用户（如果不存在）
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'myy_app') THEN
        CREATE ROLE myy_app WITH LOGIN PASSWORD 'myy_app_password';
    END IF;
END
$$;

-- 授予权限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO myy_app;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO myy_app;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO myy_app;

-- 设置默认权限，确保新表也能被 myy_app 访问
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON TABLES TO myy_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO myy_app;

-- ================================================
-- 8. 输出初始化完成信息
-- ================================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'PostgreSQL 初始化完成！';
    RAISE NOTICE '已创建扩展: vector, uuid-ossp, pg_trgm';
    RAISE NOTICE '已创建配置表和积分套餐表';
    RAISE NOTICE '已初始化积分转换配置和5个预设套餐';
    RAISE NOTICE '========================================';
END
$$;
