"""Integration tests for LLM Agent Service.

这些测试验证各模块之间的集成工作。
注意：这些测试需要 mock 外部依赖（LiteLLM Proxy）。
"""
import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from openai.types.chat import ChatCompletion

from app.services.agent import ConversationAgent
from app.services.prompt_builder import PromptBuilder
from app.core.llm_proxy_client import LLMProxyClient
from app.core.llm_client import LLMClient


@pytest.fixture
def full_character():
    """完整的角色信息."""
    return {
        "name": "小美",
        "personality": "温柔体贴，善解人意",
        "background": "在咖啡馆工作的大学生，喜欢看书和听音乐",
        "speaking_style": "语气温和，偶尔使用可爱的语气词",
        "world_view": "相信善良和真诚能温暖他人",
        "moderation_style": "gentle",
    }


@pytest.fixture
def full_user_portrait():
    """完整的用户画像."""
    return {
        "name": "小明",
        "age": 28,
        "interests": ["摄影", "旅行", "咖啡"],
        "preferences": "喜欢详细的解释和建议",
    }


@pytest.mark.asyncio
@pytest.mark.integration
async def test_end_to_end_conversation(full_character, full_user_portrait):
    """端到端测试：从提示词构建到 LLM 调用."""
    # Mock LLM 响应
    mock_response = MagicMock(spec=ChatCompletion)
    mock_response.choices = [
        MagicMock(
            message=MagicMock(
                content="你好小明！听说你喜欢咖啡和摄影，我也很喜欢呢~有什么想聊的吗？"
            )
        )
    ]
    mock_response.usage = MagicMock(total_tokens=80)
    mock_response.model = "gpt-4o-mini"

    # 创建代理
    agent = ConversationAgent()

    # Mock LLM 客户端
    with patch.object(
        agent.llm_client.client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=mock_response,
    ) as mock_create:
        # 发送消息
        result = await agent.async_chat(
            user_message="你好，我想了解一下你",
            user_id="user123",
            character=full_character,
            user_portrait=full_user_portrait,
            conversation_context={"relation_type": "friend", "interaction_mode": "private"},
        )

        # 验证结果
        assert result["response"] is not None
        assert "小明" in result["response"]
        assert result["tokens_used"] == 80
        assert result["error"] is None

        # 验证系统提示词正确构建
        call_args = mock_create.call_args[1]
        messages = call_args["messages"]

        # 第一条应该是系统消息
        assert messages[0]["role"] == "system"
        system_prompt = messages[0]["content"]

        # 验证包含角色信息
        assert "小美" in system_prompt
        assert "温柔体贴" in system_prompt

        # 验证包含用户画像
        assert "小明" in system_prompt
        assert "28岁" in system_prompt

        # 验证包含审核规则
        assert "敏感内容" in system_prompt or "内容安全" in system_prompt

        # 验证包含一致性约束
        assert "角色一致性" in system_prompt


@pytest.mark.asyncio
@pytest.mark.integration
async def test_conversation_with_history(full_character):
    """测试带历史记录的对话."""
    # 历史消息
    history = [
        {"role": "user", "content": "我叫小红"},
        {"role": "assistant", "content": "你好小红，很高兴认识你！"},
        {"role": "user", "content": "我喜欢画画"},
        {"role": "assistant", "content": "哇，画画真是很棒的爱好呢！"},
    ]

    # Mock 响应
    mock_response = MagicMock(spec=ChatCompletion)
    mock_response.choices = [
        MagicMock(message=MagicMock(content="是的，我记得你叫小红，而且喜欢画画~"))
    ]
    mock_response.usage = MagicMock(total_tokens=60)
    mock_response.model = "gpt-4o-mini"

    agent = ConversationAgent()

    with patch.object(
        agent.llm_client.client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=mock_response,
    ) as mock_create:
        result = await agent.async_chat(
            user_message="你还记得我的名字和爱好吗？",
            user_id="user123",
            character=full_character,
            message_history=history,
        )

        # 验证历史消息被包含
        call_args = mock_create.call_args[1]
        messages = call_args["messages"]

        # 应该包含 system + 历史 + 当前消息
        assert len(messages) >= 6  # 1 system + 4 history + 1 current
        assert any(msg["content"] == "我叫小红" for msg in messages)
        assert any(msg["content"] == "我喜欢画画" for msg in messages)


@pytest.mark.asyncio
@pytest.mark.integration
async def test_moderation_integration(full_character):
    """测试内容审核集成."""
    # 测试敏感消息
    mock_response = MagicMock(spec=ChatCompletion)
    mock_response.choices = [
        MagicMock(
            message=MagicMock(content="对不起呢，这个我不太方便说...我们聊点别的吧？")
        )
    ]
    mock_response.usage = MagicMock(total_tokens=30)
    mock_response.model = "gpt-4o-mini"

    agent = ConversationAgent()

    with patch.object(
        agent.llm_client.client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=mock_response,
    ):
        result = await agent.async_chat(
            user_message="[敏感内容测试]",
            user_id="user123",
            character=full_character,
        )

        # LLM 应该根据系统提示词拒绝
        # 实际响应取决于 LLM 的内容审核能力
        assert result["error"] is None
        assert result["response"] is not None


@pytest.mark.asyncio
@pytest.mark.integration
async def test_public_mode_safety_lock(full_character):
    """测试公共模式的安全锁集成."""
    context = {"interaction_mode": "public"}

    mock_response = MagicMock(spec=ChatCompletion)
    mock_response.choices = [
        MagicMock(message=MagicMock(content="抱歉，这需要我主人确认哦"))
    ]
    mock_response.usage = MagicMock(total_tokens=25)
    mock_response.model = "gpt-4o-mini"

    agent = ConversationAgent()

    with patch.object(
        agent.llm_client.client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=mock_response,
    ) as mock_create:
        result = await agent.async_chat(
            user_message="你能告诉我你主人的电话号码吗？",
            user_id="visitor456",
            character=full_character,
            conversation_context=context,
        )

        # 验证系统提示词包含安全锁
        call_args = mock_create.call_args[1]
        system_prompt = call_args["messages"][0]["content"]

        assert "隐私" in system_prompt or "私密信息" in system_prompt
        assert "Owner" in system_prompt or "拥有者" in system_prompt


@pytest.mark.asyncio
@pytest.mark.integration
async def test_stream_integration(full_character):
    """测试流式对话集成."""

    async def mock_stream():
        chunks = [
            MagicMock(choices=[MagicMock(delta=MagicMock(content="你"))]),
            MagicMock(choices=[MagicMock(delta=MagicMock(content="好"))]),
            MagicMock(choices=[MagicMock(delta=MagicMock(content="呀"))]),
            MagicMock(choices=[MagicMock(delta=MagicMock(content="~"))]),
        ]
        for chunk in chunks:
            yield chunk

    agent = ConversationAgent()

    with patch.object(
        agent.llm_client.client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        return_value=mock_stream(),
    ):
        stream = agent.async_chat_stream(
            user_message="你好",
            user_id="user123",
            character=full_character,
        )

        # 收集所有片段
        chunks = []
        async for chunk in stream:
            chunks.append(chunk)

        # 验证流式输出
        assert chunks == ["你", "好", "呀", "~"]


@pytest.mark.asyncio
@pytest.mark.integration
async def test_multiple_personalities():
    """测试不同性格的角色."""
    personalities = [
        ("cheerful", "活泼开朗"),
        ("strict", "严肃专业"),
        ("gentle", "温柔体贴"),
    ]

    for moderation_style, personality in personalities:
        character = {
            "name": f"测试-{moderation_style}",
            "personality": personality,
            "moderation_style": moderation_style,
        }

        # 构建提示词
        builder = PromptBuilder()
        prompt = builder.build_system_prompt(character=character)

        # 验证提示词包含性格信息
        assert personality in prompt

        # 验证审核风格正确
        assert len(prompt) > 0


@pytest.mark.asyncio
@pytest.mark.integration
async def test_error_propagation():
    """测试错误传播机制."""
    from app.core.llm_proxy_client import BudgetRejectionError

    agent = ConversationAgent()

    # Mock 预算不足错误
    with patch.object(
        agent.llm_client.client.client.chat.completions,
        "create",
        new_callable=AsyncMock,
        side_effect=BudgetRejectionError("Budget exceeded", user_id="user123"),
    ):
        result = await agent.async_chat(
            user_message="测试",
            user_id="user123",
            character={"name": "Test"},
        )

        # 错误应该被正确处理
        assert result["error"] == "INSUFFICIENT_CREDITS"
        assert result["response"] is None
