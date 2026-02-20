"""Unit tests for LLM Proxy Client.

测试高级包装器的错误处理和重试逻辑。
"""
import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from openai import (
    APIError,
    RateLimitError,
    APIConnectionError,
    APITimeoutError,
    AuthenticationError,
)
from openai.types.chat import ChatCompletion

from app.core.llm_proxy_client import (
    LLMProxyClient,
    BudgetRejectionError,
    get_llm_proxy_client,
)
from app.core.llm_client import LLMClient


@pytest.fixture
def mock_llm_client():
    """创建 mock LLMClient."""
    return MagicMock(spec=LLMClient)


@pytest.fixture
def proxy_client(mock_llm_client):
    """创建 LLMProxyClient 实例."""
    return LLMProxyClient(
        client=mock_llm_client,
        max_retries=3,
        retry_delay=0.1,  # 快速重试用于测试
    )


@pytest.mark.asyncio
async def test_chat_completion_success(proxy_client, mock_llm_client):
    """测试成功的请求."""
    mock_response = MagicMock(spec=ChatCompletion)
    mock_llm_client.chat_completion = AsyncMock(return_value=mock_response)

    result = await proxy_client.chat_completion(
        messages=[{"role": "user", "content": "Hello"}],
        user_id="user123",
    )

    assert result == mock_response
    mock_llm_client.chat_completion.assert_called_once()


@pytest.mark.asyncio
async def test_budget_rejection(proxy_client, mock_llm_client):
    """测试预算不足错误."""
    mock_llm_client.chat_completion = AsyncMock(
        side_effect=AuthenticationError(
            "User budget exceeded", response=MagicMock(), body={}
        )
    )

    with pytest.raises(BudgetRejectionError) as exc_info:
        await proxy_client.chat_completion(
            messages=[{"role": "user", "content": "Test"}],
            user_id="user123",
        )

    assert exc_info.value.user_id == "user123"


@pytest.mark.asyncio
async def test_rate_limit_retry(proxy_client, mock_llm_client):
    """测试限流错误重试."""
    mock_response = MagicMock(spec=ChatCompletion)

    # 第一次限流，第二次成功
    mock_llm_client.chat_completion = AsyncMock(
        side_effect=[
            RateLimitError("Rate limit", response=MagicMock(), body={}),
            mock_response,
        ]
    )

    result = await proxy_client.chat_completion(
        messages=[{"role": "user", "content": "Test"}]
    )

    assert result == mock_response
    assert mock_llm_client.chat_completion.call_count == 2


@pytest.mark.asyncio
async def test_connection_error_retry(proxy_client, mock_llm_client):
    """测试连接错误重试."""
    mock_response = MagicMock(spec=ChatCompletion)

    # 前两次连接错误，第三次成功
    mock_llm_client.chat_completion = AsyncMock(
        side_effect=[
            APIConnectionError(message="Connection failed", request=MagicMock()),
            APITimeoutError(request=MagicMock()),
            mock_response,
        ]
    )

    result = await proxy_client.chat_completion(
        messages=[{"role": "user", "content": "Test"}]
    )

    assert result == mock_response
    assert mock_llm_client.chat_completion.call_count == 3


@pytest.mark.asyncio
async def test_server_error_retry(proxy_client, mock_llm_client):
    """测试服务器错误重试."""
    mock_response = MagicMock(spec=ChatCompletion)

    # 5xx 错误重试
    api_error = APIError("Server error", request=MagicMock(), body={})
    api_error.status_code = 500

    mock_llm_client.chat_completion = AsyncMock(
        side_effect=[api_error, mock_response]
    )

    result = await proxy_client.chat_completion(
        messages=[{"role": "user", "content": "Test"}]
    )

    assert result == mock_response
    assert mock_llm_client.chat_completion.call_count == 2


@pytest.mark.asyncio
async def test_client_error_no_retry(proxy_client, mock_llm_client):
    """测试客户端错误不重试."""
    api_error = APIError("Bad request", request=MagicMock(), body={})
    api_error.status_code = 400

    mock_llm_client.chat_completion = AsyncMock(side_effect=api_error)

    with pytest.raises(APIError):
        await proxy_client.chat_completion(
            messages=[{"role": "user", "content": "Test"}]
        )

    # 4xx 错误不重试
    assert mock_llm_client.chat_completion.call_count == 1


@pytest.mark.asyncio
async def test_max_retries_exhausted(proxy_client, mock_llm_client):
    """测试重试次数耗尽."""
    mock_llm_client.chat_completion = AsyncMock(
        side_effect=RateLimitError("Rate limit", response=MagicMock(), body={})
    )

    with pytest.raises(RateLimitError):
        await proxy_client.chat_completion(
            messages=[{"role": "user", "content": "Test"}]
        )

    # 应该尝试 max_retries 次
    assert mock_llm_client.chat_completion.call_count == 3


@pytest.mark.asyncio
async def test_health_check(proxy_client, mock_llm_client):
    """测试健康检查."""
    mock_llm_client.health_check = AsyncMock(return_value=True)

    result = await proxy_client.health_check()
    assert result is True


@pytest.mark.asyncio
async def test_test_budget_success(proxy_client, mock_llm_client):
    """测试预算测试成功."""
    mock_llm_client.chat_completion = AsyncMock(return_value=MagicMock())

    result = await proxy_client.test_budget(user_id="user123")
    assert result is True


@pytest.mark.asyncio
async def test_test_budget_failure(proxy_client, mock_llm_client):
    """测试预算测试失败."""
    mock_llm_client.chat_completion = AsyncMock(
        side_effect=AuthenticationError(
            "Budget exceeded", response=MagicMock(), body={}
        )
    )

    # BudgetRejectionError 会在内部被捕获并转换为 False
    with patch.object(
        proxy_client, "chat_completion", side_effect=BudgetRejectionError("Test")
    ):
        result = await proxy_client.test_budget(user_id="user123")
        assert result is False


def test_get_llm_proxy_client_singleton():
    """测试单例模式."""
    client1 = get_llm_proxy_client()
    client2 = get_llm_proxy_client()
    assert client1 is client2
