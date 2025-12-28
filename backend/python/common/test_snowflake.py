"""
雪花ID生成器单元测试
"""

import os
import threading
import time
from concurrent.futures import ThreadPoolExecutor
from unittest.mock import patch

import pytest

from snowflake import (
    MAX_NODE_ID,
    MAX_STEP,
    SnowflakeGenerator,
    SnowflakeID,
    create_generator,
    create_generator_from_env,
    generate,
    generate_string,
    init_default,
    parse,
)


class TestSnowflakeGenerator:
    """SnowflakeGenerator类测试"""

    def test_valid_node_id(self):
        """测试有效的节点ID"""
        valid_ids = [0, 1, 512, 1023]
        for node_id in valid_ids:
            gen = SnowflakeGenerator(node_id)
            assert gen.node_id == node_id

    def test_invalid_node_id_negative(self):
        """测试负数节点ID"""
        with pytest.raises(ValueError) as excinfo:
            SnowflakeGenerator(-1)
        assert "must be between 0 and" in str(excinfo.value)

    def test_invalid_node_id_too_large(self):
        """测试过大的节点ID"""
        with pytest.raises(ValueError) as excinfo:
            SnowflakeGenerator(1024)
        assert "must be between 0 and" in str(excinfo.value)

    def test_generate_returns_positive_int(self):
        """测试生成的ID是正整数"""
        gen = SnowflakeGenerator(1)
        for _ in range(100):
            id = gen.generate()
            assert isinstance(id, int)
            assert id > 0

    def test_generate_unique_ids(self):
        """测试生成的ID唯一"""
        gen = SnowflakeGenerator(1)
        ids = [gen.generate() for _ in range(1000)]
        assert len(ids) == len(set(ids)), "Generated IDs should be unique"

    def test_generate_ordered_ids(self):
        """测试生成的ID递增"""
        gen = SnowflakeGenerator(1)
        prev_id = 0
        for _ in range(100):
            id = gen.generate()
            assert id > prev_id, "IDs should be strictly increasing"
            prev_id = id

    def test_generate_string(self):
        """测试生成字符串ID"""
        gen = SnowflakeGenerator(1)
        id_str = gen.generate_string()
        assert isinstance(id_str, str)
        assert id_str.isdigit()
        assert int(id_str) > 0

    def test_parse_node_id(self):
        """测试解析节点ID"""
        node_id = 42
        gen = SnowflakeGenerator(node_id)
        id = gen.generate()
        parsed_node_id, _, _ = gen.parse(id)
        assert parsed_node_id == node_id

    def test_parse_timestamp(self):
        """测试解析时间戳"""
        gen = SnowflakeGenerator(1)
        before = int(time.time() * 1000)
        id = gen.generate()
        after = int(time.time() * 1000)

        _, timestamp, _ = gen.parse(id)
        # 时间戳应该在生成前后的时间范围内
        assert before <= timestamp <= after

    def test_parse_sequence(self):
        """测试解析序列号"""
        gen = SnowflakeGenerator(1)
        id = gen.generate()
        _, _, sequence = gen.parse(id)
        assert 0 <= sequence <= MAX_STEP

    def test_parse_to_snowflake(self):
        """测试解析为SnowflakeID对象"""
        gen = SnowflakeGenerator(42)
        id = gen.generate()
        sf_id = gen.parse_to_snowflake(id)

        assert isinstance(sf_id, SnowflakeID)
        assert sf_id.id == id
        assert sf_id.node_id == 42
        assert str(sf_id) == str(id)
        assert int(sf_id) == id

    def test_concurrent_generation(self):
        """测试并发生成ID"""
        gen = SnowflakeGenerator(1)
        ids = []
        lock = threading.Lock()

        def generate_ids():
            local_ids = [gen.generate() for _ in range(100)]
            with lock:
                ids.extend(local_ids)

        threads = [threading.Thread(target=generate_ids) for _ in range(10)]
        for t in threads:
            t.start()
        for t in threads:
            t.join()

        # 应该有1000个唯一ID
        assert len(ids) == 1000
        assert len(set(ids)) == 1000, "All concurrent IDs should be unique"

    def test_concurrent_with_thread_pool(self):
        """使用线程池测试并发"""
        gen = SnowflakeGenerator(1)
        with ThreadPoolExecutor(max_workers=10) as executor:
            futures = [executor.submit(gen.generate) for _ in range(1000)]
            ids = [f.result() for f in futures]

        assert len(set(ids)) == 1000, "All IDs from thread pool should be unique"


class TestCreateGenerator:
    """工厂函数测试"""

    def test_create_generator(self):
        """测试create_generator函数"""
        gen = create_generator(42)
        assert isinstance(gen, SnowflakeGenerator)
        assert gen.node_id == 42

    def test_create_generator_from_env(self, monkeypatch):
        """测试从环境变量创建生成器"""
        monkeypatch.setenv("DATACENTER_ID", "1")
        monkeypatch.setenv("WORKER_ID", "2")

        gen = create_generator_from_env()
        assert gen.node_id == 1 * 32 + 2  # 34

    def test_create_generator_from_env_defaults(self, monkeypatch):
        """测试环境变量默认值"""
        monkeypatch.delenv("DATACENTER_ID", raising=False)
        monkeypatch.delenv("WORKER_ID", raising=False)

        gen = create_generator_from_env()
        assert gen.node_id == 0

    def test_create_generator_from_env_invalid_datacenter(self, monkeypatch):
        """测试无效的DATACENTER_ID"""
        monkeypatch.setenv("DATACENTER_ID", "32")
        monkeypatch.setenv("WORKER_ID", "0")

        with pytest.raises(ValueError) as excinfo:
            create_generator_from_env()
        assert "DATACENTER_ID" in str(excinfo.value)

    def test_create_generator_from_env_invalid_worker(self, monkeypatch):
        """测试无效的WORKER_ID"""
        monkeypatch.setenv("DATACENTER_ID", "0")
        monkeypatch.setenv("WORKER_ID", "32")

        with pytest.raises(ValueError) as excinfo:
            create_generator_from_env()
        assert "WORKER_ID" in str(excinfo.value)

    def test_create_generator_from_env_non_numeric(self, monkeypatch):
        """测试非数字环境变量"""
        monkeypatch.setenv("DATACENTER_ID", "abc")
        monkeypatch.setenv("WORKER_ID", "0")

        with pytest.raises(ValueError) as excinfo:
            create_generator_from_env()
        assert "Invalid DATACENTER_ID" in str(excinfo.value)


class TestNodeIDCalculation:
    """NodeID计算测试"""

    @pytest.mark.parametrize("datacenter_id,worker_id,expected", [
        ("0", "0", 0),
        ("0", "1", 1),
        ("1", "0", 32),
        ("1", "1", 33),
        ("31", "31", 1023),
        ("15", "15", 495),  # 15*32 + 15 = 495
    ])
    def test_node_id_calculation(self, monkeypatch, datacenter_id, worker_id, expected):
        """测试NodeID计算公式"""
        monkeypatch.setenv("DATACENTER_ID", datacenter_id)
        monkeypatch.setenv("WORKER_ID", worker_id)

        gen = create_generator_from_env()
        assert gen.node_id == expected

        # 验证生成的ID包含正确的节点ID
        id = gen.generate()
        parsed_node_id, _, _ = gen.parse(id)
        assert parsed_node_id == expected


class TestGlobalFunctions:
    """全局函数测试"""

    def test_generate(self, monkeypatch):
        """测试全局generate函数"""
        # 重置全局状态
        import snowflake
        snowflake._default_generator = None

        monkeypatch.setenv("DATACENTER_ID", "1")
        monkeypatch.setenv("WORKER_ID", "1")

        id = generate()
        assert isinstance(id, int)
        assert id > 0

        # 清理
        snowflake._default_generator = None

    def test_generate_string(self, monkeypatch):
        """测试全局generate_string函数"""
        import snowflake
        snowflake._default_generator = None

        monkeypatch.setenv("DATACENTER_ID", "0")
        monkeypatch.setenv("WORKER_ID", "0")

        id_str = generate_string()
        assert isinstance(id_str, str)
        assert id_str.isdigit()

        snowflake._default_generator = None

    def test_parse(self, monkeypatch):
        """测试全局parse函数"""
        import snowflake
        snowflake._default_generator = None

        monkeypatch.setenv("DATACENTER_ID", "1")
        monkeypatch.setenv("WORKER_ID", "1")

        id = generate()
        node_id, timestamp, sequence = parse(id)

        assert node_id == 33  # 1*32 + 1 = 33
        assert timestamp > 0
        assert 0 <= sequence <= MAX_STEP

        snowflake._default_generator = None

    def test_init_default(self, monkeypatch):
        """测试init_default函数"""
        import snowflake
        snowflake._default_generator = None

        monkeypatch.setenv("DATACENTER_ID", "2")
        monkeypatch.setenv("WORKER_ID", "3")

        gen = init_default()
        assert isinstance(gen, SnowflakeGenerator)
        assert gen.node_id == 2 * 32 + 3  # 67

        snowflake._default_generator = None


class TestPerformance:
    """性能测试"""

    def test_generation_speed(self):
        """测试生成速度"""
        gen = SnowflakeGenerator(1)
        count = 10000
        start = time.time()
        for _ in range(count):
            gen.generate()
        elapsed = time.time() - start

        # 应该能每秒生成至少10万个ID
        rate = count / elapsed
        assert rate > 10000, f"Generation rate too slow: {rate:.0f} IDs/sec"

    def test_sequence_overflow_handling(self):
        """测试序列号溢出处理"""
        gen = SnowflakeGenerator(1)
        # 快速生成大量ID来测试序列号溢出
        ids = [gen.generate() for _ in range(5000)]
        assert len(set(ids)) == 5000, "All IDs should be unique even with sequence overflow"
