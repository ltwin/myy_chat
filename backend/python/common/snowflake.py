"""
雪花ID生成器 - Python实现

使用Twitter的Snowflake算法生成64位分布式唯一ID。
与Go版本保持一致的接口和行为。
"""

import os
import threading
import time
from dataclasses import dataclass
from typing import Tuple

# Snowflake ID 结构 (64位):
# - 1位: 符号位 (始终为0)
# - 41位: 时间戳 (毫秒级，从epoch开始)
# - 10位: 节点ID (0-1023)
# - 12位: 序列号 (0-4095)

# 默认epoch: 2020-01-01 00:00:00 UTC (与bwmarrin/snowflake库一致)
DEFAULT_EPOCH = 1288834974657  # Twitter Snowflake 默认 epoch

# 位数定义
NODE_BITS = 10
STEP_BITS = 12

# 最大值
MAX_NODE_ID = (1 << NODE_BITS) - 1  # 1023
MAX_STEP = (1 << STEP_BITS) - 1  # 4095

# 位移量
NODE_SHIFT = STEP_BITS
TIME_SHIFT = NODE_BITS + STEP_BITS


@dataclass
class SnowflakeID:
    """解析后的雪花ID"""
    id: int
    node_id: int
    timestamp: int
    sequence: int

    def __str__(self) -> str:
        return str(self.id)

    def __int__(self) -> int:
        return self.id


class SnowflakeGenerator:
    """
    雪花ID生成器

    线程安全的ID生成器，每毫秒可生成4096个唯一ID。
    """

    def __init__(self, node_id: int, epoch: int = DEFAULT_EPOCH):
        """
        初始化生成器

        Args:
            node_id: 节点ID，范围 0-1023
            epoch: 起始时间戳（毫秒），默认使用Twitter Snowflake epoch

        Raises:
            ValueError: 当node_id超出有效范围时
        """
        if node_id < 0 or node_id > MAX_NODE_ID:
            raise ValueError(f"node_id must be between 0 and {MAX_NODE_ID}, got {node_id}")

        self._node_id = node_id
        self._epoch = epoch
        self._last_timestamp = -1
        self._step = 0
        self._lock = threading.Lock()

    @property
    def node_id(self) -> int:
        """获取节点ID"""
        return self._node_id

    def _current_timestamp(self) -> int:
        """获取当前时间戳（毫秒）"""
        return int(time.time() * 1000)

    def _wait_next_millis(self, last_timestamp: int) -> int:
        """等待下一毫秒"""
        timestamp = self._current_timestamp()
        while timestamp <= last_timestamp:
            timestamp = self._current_timestamp()
        return timestamp

    def generate(self) -> int:
        """
        生成一个新的雪花ID

        Returns:
            64位整数ID

        Raises:
            RuntimeError: 当时钟回拨时
        """
        with self._lock:
            timestamp = self._current_timestamp()

            # 检查时钟回拨
            if timestamp < self._last_timestamp:
                raise RuntimeError(
                    f"Clock moved backwards. Refusing to generate id for "
                    f"{self._last_timestamp - timestamp} milliseconds"
                )

            if timestamp == self._last_timestamp:
                # 同一毫秒内，递增序列号
                self._step = (self._step + 1) & MAX_STEP
                if self._step == 0:
                    # 序列号溢出，等待下一毫秒
                    timestamp = self._wait_next_millis(self._last_timestamp)
            else:
                # 新的毫秒，重置序列号
                self._step = 0

            self._last_timestamp = timestamp

            # 组装ID
            id = (
                ((timestamp - self._epoch) << TIME_SHIFT) |
                (self._node_id << NODE_SHIFT) |
                self._step
            )

            return id

    def generate_string(self) -> str:
        """生成一个新的雪花ID字符串"""
        return str(self.generate())

    def parse(self, id: int) -> Tuple[int, int, int]:
        """
        解析雪花ID

        Args:
            id: 要解析的雪花ID

        Returns:
            (node_id, timestamp, sequence) 元组
        """
        node_id = (id >> NODE_SHIFT) & MAX_NODE_ID
        timestamp = (id >> TIME_SHIFT) + self._epoch
        sequence = id & MAX_STEP
        return node_id, timestamp, sequence

    def parse_to_snowflake(self, id: int) -> SnowflakeID:
        """
        解析雪花ID并返回SnowflakeID对象

        Args:
            id: 要解析的雪花ID

        Returns:
            SnowflakeID对象
        """
        node_id, timestamp, sequence = self.parse(id)
        return SnowflakeID(
            id=id,
            node_id=node_id,
            timestamp=timestamp,
            sequence=sequence
        )


# 全局默认生成器
_default_generator: SnowflakeGenerator | None = None
_default_lock = threading.Lock()


def _get_node_id_from_env() -> int:
    """
    从环境变量计算NodeID

    NodeID = DATACENTER_ID * 32 + WORKER_ID
    与Go版本保持一致
    """
    datacenter_id_str = os.getenv("DATACENTER_ID", "0")
    worker_id_str = os.getenv("WORKER_ID", "0")

    try:
        datacenter_id = int(datacenter_id_str)
    except ValueError:
        raise ValueError(f"Invalid DATACENTER_ID: {datacenter_id_str}")

    try:
        worker_id = int(worker_id_str)
    except ValueError:
        raise ValueError(f"Invalid WORKER_ID: {worker_id_str}")

    if datacenter_id < 0 or datacenter_id > 31:
        raise ValueError(f"DATACENTER_ID must be between 0 and 31, got {datacenter_id}")

    if worker_id < 0 or worker_id > 31:
        raise ValueError(f"WORKER_ID must be between 0 and 31, got {worker_id}")

    return datacenter_id * 32 + worker_id


def init_default() -> SnowflakeGenerator:
    """
    初始化默认生成器

    从环境变量读取DATACENTER_ID和WORKER_ID

    Returns:
        默认生成器实例
    """
    global _default_generator
    with _default_lock:
        if _default_generator is None:
            node_id = _get_node_id_from_env()
            _default_generator = SnowflakeGenerator(node_id)
    return _default_generator


def generate() -> int:
    """
    使用默认生成器生成雪花ID

    如果默认生成器未初始化，会自动从环境变量初始化

    Returns:
        64位整数ID
    """
    if _default_generator is None:
        init_default()
    return _default_generator.generate()  # type: ignore


def generate_string() -> str:
    """
    使用默认生成器生成雪花ID字符串

    Returns:
        ID字符串
    """
    return str(generate())


def parse(id: int) -> Tuple[int, int, int]:
    """
    解析雪花ID

    Args:
        id: 要解析的雪花ID

    Returns:
        (node_id, timestamp, sequence) 元组
    """
    if _default_generator is None:
        init_default()
    return _default_generator.parse(id)  # type: ignore


def create_generator(node_id: int) -> SnowflakeGenerator:
    """
    创建新的雪花ID生成器

    Args:
        node_id: 节点ID，范围 0-1023

    Returns:
        新的生成器实例
    """
    return SnowflakeGenerator(node_id)


def create_generator_from_env() -> SnowflakeGenerator:
    """
    从环境变量创建雪花ID生成器

    读取 DATACENTER_ID (0-31) 和 WORKER_ID (0-31)
    NodeID = DATACENTER_ID * 32 + WORKER_ID

    Returns:
        新的生成器实例
    """
    node_id = _get_node_id_from_env()
    return SnowflakeGenerator(node_id)
