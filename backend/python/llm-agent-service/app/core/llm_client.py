"""OpenAI-compatible LLM Client for LiteLLM Proxy.

此模块提供与 LiteLLM Proxy 的集成，使用 OpenAI SDK 进行通信。
主要功能：
- 支持流式和非流式响应
- 用户级别预算追踪（通过 user 参数）
- 自动重试和错误处理
"""
import logging
from typing import AsyncIterator, Optional, Dict, Any, List

from openai import AsyncOpenAI, OpenAIError, APIError, RateLimitError, APIConnectionError
from openai.types.chat import ChatCompletion, ChatCompletionChunk

from app.config.settings import get_settings

logger = logging.getLogger(__name__)
settings = get_settings()


class LLMClient:
    """OpenAI-compatible client for LiteLLM Proxy.

    使用 OpenAI SDK 与 LiteLLM Proxy 通信，支持：
    - 多模型调用（通过 LiteLLM 路由）
    - 用户级别计费追踪
    - 流式和非流式响应
    - 自动错误处理
    """

    def __init__(
        self,
        base_url: Optional[str] = None,
        api_key: Optional[str] = None,
        timeout: float = 60.0,
    ):
        """初始化 LLM 客户端.

        Args:
            base_url: LiteLLM Proxy URL (默认从配置读取)
            api_key: API 密钥 (默认从配置读取)
            timeout: 请求超时时间（秒）
        """
        self.base_url = base_url or settings.litellm_proxy_url
        self.api_key = api_key or settings.litellm_api_key
        self.timeout = timeout

        # 初始化 OpenAI 客户端，指向 LiteLLM Proxy
        self.client = AsyncOpenAI(
            base_url=self.base_url,
            api_key=self.api_key or "dummy-key",  # LiteLLM 可能不需要真实 key
            timeout=timeout,
        )

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
        """发送聊天补全请求.

        Args:
            messages: 消息列表 [{"role": "user", "content": "..."}]
            model: 模型标识符（通过 LiteLLM 路由）
            user_id: 用户ID（用于预算追踪和限流）
            temperature: 采样温度 (0.0 到 2.0)
            max_tokens: 最大生成 token 数
            stream: 是否流式返回
            **kwargs: 其他参数传递给 OpenAI API

        Returns:
            ChatCompletion 对象或 ChatCompletionChunk 流式迭代器

        Raises:
            OpenAIError: OpenAI SDK 相关错误
        """
        params: Dict[str, Any] = {
            "model": model,
            "messages": messages,
            "temperature": temperature,
            "stream": stream,
            **kwargs,
        }

        # 添加 user_id 用于预算追踪（LiteLLM 支持）
        if user_id:
            params["user"] = user_id

        if max_tokens:
            params["max_tokens"] = max_tokens

        logger.info(
            f"LLM request: model={model}, user={user_id}, "
            f"messages_count={len(messages)}, stream={stream}"
        )

        try:
            if stream:
                # Async generator - don't await, just return
                return self._stream_completion(**params)
            else:
                return await self._sync_completion(**params)
        except Exception as e:
            logger.error(f"LLM request failed: {type(e).__name__}: {e}")
            raise

    async def _sync_completion(self, **params: Any) -> ChatCompletion:
        """非流式聊天补全.

        Args:
            **params: 传递给 OpenAI API 的参数

        Returns:
            ChatCompletion 对象
        """
        response = await self.client.chat.completions.create(**params)

        logger.info(
            f"LLM response: model={response.model}, "
            f"tokens={response.usage.total_tokens if response.usage else 'N/A'}"
        )

        return response

    async def _stream_completion(
        self, **params: Any
    ) -> AsyncIterator[ChatCompletionChunk]:
        """流式聊天补全.

        Args:
            **params: 传递给 OpenAI API 的参数

        Yields:
            ChatCompletionChunk 对象
        """
        stream = await self.client.chat.completions.create(**params)

        async for chunk in stream:
            yield chunk

    async def health_check(self) -> bool:
        """检查 LiteLLM Proxy 健康状态.

        Returns:
            True 表示健康，False 表示不健康
        """
        try:
            # 发送简单请求验证连接
            response = await self.client.chat.completions.create(
                model="gpt-4o-mini",
                messages=[{"role": "user", "content": "hi"}],
                max_tokens=1,
            )
            return response is not None
        except Exception as e:
            logger.warning(f"LiteLLM health check failed: {e}")
            return False


# 全局单例实例
_llm_client: Optional[LLMClient] = None


def get_llm_client() -> LLMClient:
    """获取 LLM 客户端单例.

    Returns:
        LLMClient 实例
    """
    global _llm_client
    if _llm_client is None:
        _llm_client = LLMClient()
    return _llm_client
