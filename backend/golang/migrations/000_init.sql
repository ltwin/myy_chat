-- Migration: 000_init
-- Description: 初始化数据库：创建必要的函数和扩展
-- This must run before all other migrations

-- ================================================
-- 创建 PostgreSQL 扩展
-- ================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ================================================
-- 自动更新 updated_at 列的触发器函数
-- ================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION update_updated_at_column() IS '自动更新 updated_at 字段的触发器函数';

-- ================================================
-- 创建按月分区表的辅助函数
-- ================================================

CREATE OR REPLACE FUNCTION create_monthly_partition(
    table_name TEXT,
    year INT,
    month INT
) RETURNS VOID AS $$
DECLARE
    partition_name TEXT;
    start_date DATE;
    end_date DATE;
BEGIN
    -- 生成分区表名：表名_YYYY_MM
    partition_name := table_name || '_' || year || '_' || LPAD(month::TEXT, 2, '0');

    -- 计算分区的开始和结束日期
    start_date := DATE(year || '-' || month || '-01');
    end_date := start_date + INTERVAL '1 month';

    -- 创建分区表（如果不存在）
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF %I FOR VALUES FROM (%L) TO (%L)',
        partition_name,
        table_name,
        start_date,
        end_date
    );

    RAISE NOTICE 'Created partition: %', partition_name;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION create_monthly_partition(TEXT, INT, INT) IS '创建按月分区表的辅助函数';

-- ================================================
-- 输出完成信息
-- ================================================

DO $$
BEGIN
    RAISE NOTICE '======================================';
    RAISE NOTICE '数据库初始化完成';
    RAISE NOTICE '- update_updated_at_column(): 自动更新 updated_at';
    RAISE NOTICE '- create_monthly_partition(): 创建月度分区';
    RAISE NOTICE '======================================';
END
$$;
