"""Unit tests for Conversation Agent.

测试 LangGraph 对话代理的功能。
"""
import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from openai.types.chat import ChatCompletion, ChatCompletionChunk

from app.services.agent import ConversationAgent, get_conversation_agent
from app.core.llm_proxy_client import BudgetRejectionError


@pytest.fixture
def agent():
    """创建 ConversationAgent 实例."""
    return ConversationAgent()


@pytest.fixture
def sample_character():
    """创建示例角色."""
    return {
        "name": "测试角色",
        "personality": "友善",
        "background": "测试背景",
        "speaking_style": "自然",
        "world_view": "积极",
        "moderation_style": "default",
    }


@pytest.fixture
def mock_completion_response():
    """创建 mock 聊天补全响应."""
    mock = MagicMock(spec=ChatCompletion)
    mock.choices = [
        MagicMock(message=MagicMock(content="你好！我很高兴与你交谈。"))
    ]
    mock.usage = MagicMock(total_tokens=50)
    mock.model = "gpt-4o-mini"
    return mock


@pytest.mark.asyncio
async def test_async_chat_success(agent, sample_character, mock_completion_response):
    """测试成功的异步对话."""
    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        return_value=mock_completion_response,
    ):
        result = await agent.async_chat(
            user_message="你好",
            user_id="user123",
            character=sample_character,
        )

        assert result["response"] == "你好！我很高兴与你交谈。"
        assert result["tokens_used"] == 50
        assert result["model"] == "gpt-4o-mini"
        assert result["error"] is None


@pytest.mark.asyncio
async def test_async_chat_with_history(
    agent, sample_character, mock_completion_response
):
    """测试带消息历史的对话."""
    message_history = [
        {"role": "user", "content": "我叫张三"},
        {"role": "assistant", "content": "你好，张三！"},
    ]

    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        return_value=mock_completion_response,
    ) as mock_call:
        result = await agent.async_chat(
            user_message="你还记得我的名字吗？",
            user_id="user123",
            character=sample_character,
            message_history=message_history,
        )

        # 验证调用时包含历史消息
        call_args = mock_call.call_args[1]
        messages = call_args["messages"]

        # 应该包含 system + 历史 + 当前消息
        assert len(messages) >= 4
        assert messages[0]["role"] == "system"
        assert any(msg["content"] == "我叫张三" for msg in messages)


@pytest.mark.asyncio
async def test_async_chat_with_user_portrait(
    agent, sample_character, mock_completion_response
):
    """测试带用户画像的对话."""
    user_portrait = {
        "name": "李四",
        "age": 30,
        "interests": ["音乐", "运动"],
    }

    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        return_value=mock_completion_response,
    ) as mock_call:
        result = await agent.async_chat(
            user_message="介绍一下自己",
            user_id="user123",
            character=sample_character,
            user_portrait=user_portrait,
        )

        # 验证系统提示词包含用户画像
        call_args = mock_call.call_args[1]
        system_message = call_args["messages"][0]["content"]
        assert "李四" in system_message


@pytest.mark.asyncio
async def test_async_chat_with_context(
    agent, sample_character, mock_completion_response
):
    """测试带对话上下文的对话."""
    conversation_context = {
        "relation_type": "friend",
        "interaction_mode": "private",
    }

    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        return_value=mock_completion_response,
    ) as mock_call:
        result = await agent.async_chat(
            user_message="测试",
            user_id="user123",
            character=sample_character,
            conversation_context=conversation_context,
        )

        # 验证系统提示词包含上下文
        call_args = mock_call.call_args[1]
        system_message = call_args["messages"][0]["content"]
        assert "朋友" in system_message


@pytest.mark.asyncio
async def test_async_chat_budget_rejection(agent, sample_character):
    """测试预算不足错误处理."""
    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        side_effect=BudgetRejectionError("Insufficient credits", user_id="user123"),
    ):
        result = await agent.async_chat(
            user_message="测试",
            user_id="user123",
            character=sample_character,
        )

        assert result["response"] is None
        assert result["tokens_used"] == 0
        assert result["error"] == "INSUFFICIENT_CREDITS"


@pytest.mark.asyncio
async def test_async_chat_general_error(agent, sample_character):
    """测试一般错误处理."""
    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        side_effect=Exception("Test error"),
    ):
        result = await agent.async_chat(
            user_message="测试",
            user_id="user123",
            character=sample_character,
        )

        assert result["response"] is None
        assert result["tokens_used"] == 0
        assert "Test error" in result["error"]


@pytest.mark.asyncio
async def test_async_chat_stream(agent, sample_character):
    """测试流式对话."""

    async def mock_stream():
        chunks = [
            MagicMock(choices=[MagicMock(delta=MagicMock(content="你"))]),
            MagicMock(choices=[MagicMock(delta=MagicMock(content="好"))]),
            MagicMock(choices=[MagicMock(delta=MagicMock(content="！"))]),
        ]
        for chunk in chunks:
            yield chunk

    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        return_value=mock_stream(),
    ):
        stream = agent.async_chat_stream(
            user_message="你好",
            user_id="user123",
            character=sample_character,
        )

        chunks = []
        async for chunk in stream:
            chunks.append(chunk)

        assert chunks == ["你", "好", "！"]


@pytest.mark.asyncio
async def test_async_chat_stream_error(agent, sample_character):
    """测试流式对话错误处理."""
    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        side_effect=Exception("Stream error"),
    ):
        stream = agent.async_chat_stream(
            user_message="测试",
            user_id="user123",
            character=sample_character,
        )

        chunks = []
        async for chunk in stream:
            chunks.append(chunk)

        assert len(chunks) == 1
        assert "ERROR" in chunks[0]


def test_messages_to_openai_format(agent):
    """测试消息格式转换."""
    from langchain_core.messages import HumanMessage, AIMessage, SystemMessage

    messages = [
        SystemMessage(content="System prompt"),
        HumanMessage(content="User message"),
        AIMessage(content="AI response"),
    ]

    openai_messages = agent._messages_to_openai_format(messages)

    assert len(openai_messages) == 3
    assert openai_messages[0] == {"role": "system", "content": "System prompt"}
    assert openai_messages[1] == {"role": "user", "content": "User message"}
    assert openai_messages[2] == {"role": "assistant", "content": "AI response"}


def test_get_conversation_agent_singleton():
    """测试单例模式."""
    agent1 = get_conversation_agent()
    agent2 = get_conversation_agent()
    assert agent1 is agent2


@pytest.mark.asyncio
async def test_async_chat_custom_model(
    agent, sample_character, mock_completion_response
):
    """测试自定义模型参数."""
    with patch.object(
        agent.llm_client,
        "chat_completion",
        new_callable=AsyncMock,
        return_value=mock_completion_response,
    ) as mock_call:
        await agent.async_chat(
            user_message="测试",
            user_id="user123",
            character=sample_character,
            model="gpt-4",
            temperature=0.9,
            max_tokens=100,
        )

        call_args = mock_call.call_args[1]
        assert call_args["model"] == "gpt-4"
        assert call_args["temperature"] == 0.9
        assert call_args["max_tokens"] == 100
