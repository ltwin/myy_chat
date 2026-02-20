"""LLM Proxy Client with enhanced error handling and retry logic.

此模块是 LLMClient 的高级包装器，提供：
- 智能重试机制（指数退避）
- 预算不足检测
- 连接池管理
- 详细错误分类
"""
import asyncio
import logging
from typing import AsyncIterator, Optional, Dict, Any, List
from functools import wraps

from openai import (
    OpenAIError,
    APIError,
    RateLimitError,
    APIConnectionError,
    APITimeoutError,
    AuthenticationError,
)
from openai.types.chat import ChatCompletion, ChatCompletionChunk

from app.core.llm_client import LLMClient, get_llm_client

logger = logging.getLogger(__name__)


class BudgetRejectionError(Exception):
    """预算不足错误.

    当用户积分余额不足时抛出此异常。
    """

    def __init__(self, message: str, user_id: Optional[str] = None):
        super().__init__(message)
        self.user_id = user_id


class LLMProxyClient:
    """LLM Proxy 客户端包装器.

    提供高级功能：
    - 自动重试（指数退避）
    - 错误分类和处理
    - 预算检测
    - 降级处理
    """

    def __init__(
        self,
        client: Optional[LLMClient] = None,
        max_retries: int = 3,
        retry_delay: float = 1.0,
    ):
        """初始化 LLM Proxy 客户端.

        Args:
            client: LLMClient 实例（默认使用全局单例）
            max_retries: 最大重试次数
            retry_delay: 初始重试延迟（秒，指数退避）
        """
        self.client = client or get_llm_client()
        self.max_retries = max_retries
        self.retry_delay = retry_delay

    async def chat_completion(
        self,
        messages: List[Dict[str, str]],
        model: str = "gpt-4o-mini",
        user_id: Optional[str] = None,
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        stream: bool = False,
        **kwargs: Any,
    ) -> ChatCompletion | AsyncIterator[ChatCompletionChunk]:
        """发送聊天补全请求（带重试）.

        Args:
            messages: 消息列表
            model: 模型标识符
            user_id: 用户ID（用于预算追踪）
            temperature: 采样温度
            max_tokens: 最大生成 token 数
            stream: 是否流式返回
            **kwargs: 其他参数

        Returns:
            ChatCompletion 对象或流式迭代器

        Raises:
            BudgetRejectionError: 预算不足
            APIError: API 错误
            APIConnectionError: 连接错误
            OpenAIError: 其他 OpenAI 错误
        """
        last_exception: Optional[Exception] = None

        for attempt in range(self.max_retries):
            try:
                return await self.client.chat_completion(
                    messages=messages,
                    model=model,
                    user_id=user_id,
                    temperature=temperature,
                    max_tokens=max_tokens,
                    stream=stream,
                    **kwargs,
                )
            except AuthenticationError as e:
                # 检测预算不足（LiteLLM 可能返回 401）
                error_msg = str(e).lower()
                if "budget" in error_msg or "quota" in error_msg or "insufficient" in error_msg:
                    logger.warning(f"Budget rejection for user {user_id}: {e}")
                    raise BudgetRejectionError(
                        f"User {user_id} has insufficient credits", user_id=user_id
                    ) from e
                raise
            except RateLimitError as e:
                # 限流错误：重试
                logger.warning(
                    f"Rate limit hit (attempt {attempt + 1}/{self.max_retries}): {e}"
                )
                last_exception = e
                if attempt < self.max_retries - 1:
                    await self._exponential_backoff(attempt)
                else:
                    raise
            except (APIConnectionError, APITimeoutError) as e:
                # 连接/超时错误：重试
                logger.warning(
                    f"Connection error (attempt {attempt + 1}/{self.max_retries}): {e}"
                )
                last_exception = e
                if attempt < self.max_retries - 1:
                    await self._exponential_backoff(attempt)
                else:
                    raise
            except APIError as e:
                # 服务器错误：5xx 重试，4xx 不重试
                if e.status_code and 500 <= e.status_code < 600:
                    logger.warning(
                        f"Server error {e.status_code} (attempt {attempt + 1}/{self.max_retries}): {e}"
                    )
                    last_exception = e
                    if attempt < self.max_retries - 1:
                        await self._exponential_backoff(attempt)
                    else:
                        raise
                else:
                    # 4xx 客户端错误，不重试
                    logger.error(f"Client error {e.status_code}: {e}")
                    raise
            except OpenAIError as e:
                # 其他 OpenAI 错误：不重试
                logger.error(f"OpenAI error: {type(e).__name__}: {e}")
                raise
            except Exception as e:
                # 未知错误：记录并抛出
                logger.error(f"Unexpected error: {type(e).__name__}: {e}")
                raise

        # 所有重试失败
        if last_exception:
            raise last_exception
        raise RuntimeError("All retries failed without exception")

    async def _exponential_backoff(self, attempt: int) -> None:
        """指数退避延迟.

        Args:
            attempt: 当前尝试次数（从0开始）
        """
        delay = self.retry_delay * (2 ** attempt)
        logger.info(f"Retrying after {delay:.2f}s...")
        await asyncio.sleep(delay)

    async def health_check(self) -> bool:
        """检查 LiteLLM Proxy 健康状态.

        Returns:
            True 表示健康，False 表示不健康
        """
        return await self.client.health_check()

    async def test_budget(self, user_id: str) -> bool:
        """测试用户预算是否充足.

        Args:
            user_id: 用户ID

        Returns:
            True 表示预算充足，False 表示不足
        """
        try:
            # 发送最小请求测试预算
            await self.chat_completion(
                messages=[{"role": "user", "content": "hi"}],
                model="gpt-4o-mini",
                user_id=user_id,
                max_tokens=1,
            )
            return True
        except BudgetRejectionError:
            return False
        except Exception as e:
            logger.warning(f"Budget test failed with unexpected error: {e}")
            return False


# 全局单例实例
_llm_proxy_client: Optional[LLMProxyClient] = None


def get_llm_proxy_client() -> LLMProxyClient:
    """获取 LLM Proxy 客户端单例.

    Returns:
        LLMProxyClient 实例
    """
    global _llm_proxy_client
    if _llm_proxy_client is None:
        _llm_proxy_client = LLMProxyClient()
    return _llm_proxy_client
