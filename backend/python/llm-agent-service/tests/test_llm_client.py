"""Unit tests for LLM Client.

测试 OpenAI SDK 客户端与 LiteLLM Proxy 的集成。
"""
import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from openai import APIError, RateLimitError, APIConnectionError
from openai.types.chat import ChatCompletion, ChatCompletionMessage
from openai.types.chat.chat_completion import Choice
from openai.types import CompletionUsage

from app.core.llm_client import LLMClient, get_llm_client


@pytest.fixture
def llm_client():
    """创建 LLM 客户端实例."""
    return LLMClient(
        base_url="http://test-litellm:4000",
        api_key="test-key",
        timeout=10.0,
    )


@pytest.fixture
def mock_chat_completion():
    """创建 mock ChatCompletion 响应."""
    return ChatCompletion(
        id="chatcmpl-test",
        object="chat.completion",
        created=1234567890,
        model="gpt-4o-mini",
        choices=[
            Choice(
                index=0,
                message=ChatCompletionMessage(
                    role="assistant",
                    content="Hello! How can I help you?",
                ),
                finish_reason="stop",
            )
        ],
        usage=CompletionUsage(
            prompt_tokens=10,
            completion_tokens=15,
            total_tokens=25,
        ),
    )


@pytest.mark.asyncio
async def test_chat_completion_success(llm_client, mock_chat_completion):
    """测试成功的聊天补全请求."""
    with patch.object(
        llm_client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=mock_chat_completion,
    ):
        result = await llm_client.chat_completion(
            messages=[{"role": "user", "content": "Hello"}],
            model="gpt-4o-mini",
            user_id="user123",
        )

        assert result.id == "chatcmpl-test"
        assert result.choices[0].message.content == "Hello! How can I help you?"
        assert result.usage.total_tokens == 25


@pytest.mark.asyncio
async def test_chat_completion_with_params(llm_client, mock_chat_completion):
    """测试带参数的聊天补全请求."""
    with patch.object(
        llm_client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=mock_chat_completion,
    ) as mock_create:
        await llm_client.chat_completion(
            messages=[{"role": "user", "content": "Test"}],
            model="gpt-4",
            user_id="user123",
            temperature=0.9,
            max_tokens=100,
        )

        # 验证调用参数
        call_args = mock_create.call_args[1]
        assert call_args["model"] == "gpt-4"
        assert call_args["user"] == "user123"
        assert call_args["temperature"] == 0.9
        assert call_args["max_tokens"] == 100


@pytest.mark.asyncio
async def test_chat_completion_stream(llm_client):
    """测试流式聊天补全."""
    # Mock 流式响应 - 创建可迭代的 async iterator
    mock_chunks = [
        MagicMock(choices=[MagicMock(delta=MagicMock(content="Hello"))]),
        MagicMock(choices=[MagicMock(delta=MagicMock(content=" there"))]),
        MagicMock(choices=[MagicMock(delta=MagicMock(content="!"))]),
    ]

    class MockAsyncIterator:
        def __init__(self, items):
            self.items = items
            self.index = 0

        def __aiter__(self):
            return self

        async def __anext__(self):
            if self.index >= len(self.items):
                raise StopAsyncIteration
            item = self.items[self.index]
            self.index += 1
            return item

    # The OpenAI client.chat.completions.create returns an async iterator directly
    with patch.object(
        llm_client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=MockAsyncIterator(mock_chunks),
    ):
        # chat_completion is async, so await it first to get the async generator
        result = await llm_client.chat_completion(
            messages=[{"role": "user", "content": "Hi"}],
            stream=True,
        )

        chunks = []
        async for chunk in result:
            chunks.append(chunk)

        assert len(chunks) == 3
        assert chunks[0].choices[0].delta.content == "Hello"


@pytest.mark.asyncio
async def test_chat_completion_error(llm_client):
    """测试 LLM 请求错误处理."""
    with patch.object(
        llm_client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        side_effect=APIError("Test error", request=MagicMock(), body={}),
    ):
        with pytest.raises(APIError):
            await llm_client.chat_completion(
                messages=[{"role": "user", "content": "Test"}]
            )


@pytest.mark.asyncio
async def test_health_check_success(llm_client, mock_chat_completion):
    """测试健康检查成功."""
    with patch.object(
        llm_client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=mock_chat_completion,
    ):
        result = await llm_client.health_check()
        assert result is True


@pytest.mark.asyncio
async def test_health_check_failure(llm_client):
    """测试健康检查失败."""
    with patch.object(
        llm_client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        side_effect=APIConnectionError(message="Connection failed", request=MagicMock()),
    ):
        result = await llm_client.health_check()
        assert result is False


def test_get_llm_client_singleton():
    """测试单例模式."""
    client1 = get_llm_client()
    client2 = get_llm_client()
    assert client1 is client2
