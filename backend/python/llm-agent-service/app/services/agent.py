"""LangGraph Agent Workflow for AI Character Conversations.

此模块实现基于 LangGraph 的 AI 对话流程，包括：
- 简单对话工作流（无工具调用，Phase 1）
- 系统提示词注入
- 消息历史管理
- 流式和非流式响应
"""
import logging
from typing import Dict, Any, List, Optional, AsyncIterator, TypedDict
from datetime import datetime

from langchain_core.messages import HumanMessage, AIMessage, SystemMessage, BaseMessage
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.state import CompiledStateGraph

from app.core.llm_proxy_client import get_llm_proxy_client, BudgetRejectionError
from app.services.prompt_builder import get_prompt_builder

logger = logging.getLogger(__name__)


class AgentState(TypedDict):
    """LangGraph 状态定义.

    存储对话过程中的所有必要信息。
    """

    messages: List[BaseMessage]  # 消息历史（包含 System/Human/AI）
    user_id: str  # 用户ID
    character: Dict[str, Any]  # 角色信息
    user_portrait: Optional[Dict[str, Any]]  # 用户画像
    conversation_context: Optional[Dict[str, Any]]  # 对话上下文
    model: str  # 使用的模型
    temperature: float  # 温度参数
    max_tokens: Optional[int]  # 最大 token 数
    error: Optional[str]  # 错误信息


class ConversationAgent:
    """AI 对话代理（基于 LangGraph）.

    实现基础对话工作流：
    1. 构建系统提示词
    2. 注入消息历史
    3. 调用 LLM
    4. 返回响应
    """

    def __init__(self):
        """初始化对话代理."""
        self.llm_client = get_llm_proxy_client()
        self.prompt_builder = get_prompt_builder()
        self.graph = self._build_graph()

    def _build_graph(self) -> CompiledStateGraph:
        """构建 LangGraph 工作流图.

        Returns:
            编译后的状态图
        """
        workflow = StateGraph(AgentState)

        # 添加节点
        workflow.add_node("prepare_prompt", self._prepare_prompt_node)
        workflow.add_node("call_llm", self._call_llm_node)

        # 定义边
        workflow.set_entry_point("prepare_prompt")
        workflow.add_edge("prepare_prompt", "call_llm")
        workflow.add_edge("call_llm", END)

        return workflow.compile()

    def _prepare_prompt_node(self, state: AgentState) -> AgentState:
        """准备系统提示词节点.

        Args:
            state: 当前状态

        Returns:
            更新后的状态
        """
        try:
            # 构建系统提示词
            system_prompt = self.prompt_builder.build_system_prompt(
                character=state["character"],
                user_portrait=state.get("user_portrait"),
                conversation_context=state.get("conversation_context"),
            )

            # 将系统提示词插入消息列表开头
            messages = state["messages"]
            if not messages or not isinstance(messages[0], SystemMessage):
                messages = [SystemMessage(content=system_prompt)] + messages
            else:
                # 替换已有的系统消息
                messages[0] = SystemMessage(content=system_prompt)

            state["messages"] = messages
            logger.info(f"Prepared system prompt for character: {state['character'].get('name')}")

        except Exception as e:
            logger.error(f"Error preparing prompt: {e}")
            state["error"] = f"Failed to prepare prompt: {str(e)}"

        return state

    def _call_llm_node(self, state: AgentState) -> AgentState:
        """调用 LLM 节点（非流式）.

        Args:
            state: 当前状态

        Returns:
            更新后的状态
        """
        if state.get("error"):
            # 如果已有错误，跳过
            return state

        try:
            # 准备消息（转换为 OpenAI 格式）
            messages = self._messages_to_openai_format(state["messages"])

            # 同步调用 LLM（注意：这里需要在异步上下文中调用）
            # 实际使用时应该通过 async 方法处理
            logger.warning("_call_llm_node is synchronous, use async_chat for production")

            # 这里不执行实际调用，留给 async_chat 方法
            state["error"] = None

        except Exception as e:
            logger.error(f"Error calling LLM: {e}")
            state["error"] = f"LLM call failed: {str(e)}"

        return state

    async def async_chat(
        self,
        user_message: str,
        user_id: str,
        character: Dict[str, Any],
        message_history: Optional[List[Dict[str, str]]] = None,
        user_portrait: Optional[Dict[str, Any]] = None,
        conversation_context: Optional[Dict[str, Any]] = None,
        model: str = "gpt-4o-mini",
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
    ) -> Dict[str, Any]:
        """异步对话（非流式）.

        Args:
            user_message: 用户消息
            user_id: 用户ID
            character: 角色信息
            message_history: 消息历史 [{"role": "user|assistant", "content": "..."}]
            user_portrait: 用户画像
            conversation_context: 对话上下文
            model: 模型标识
            temperature: 温度
            max_tokens: 最大 token 数

        Returns:
            响应字典：
            {
                "response": "AI 回复内容",
                "tokens_used": 100,
                "model": "gpt-4o-mini",
                "error": None
            }
        """
        try:
            # 构建系统提示词
            system_prompt = self.prompt_builder.build_system_prompt(
                character=character,
                user_portrait=user_portrait,
                conversation_context=conversation_context,
            )

            # 准备消息列表
            messages = [{"role": "system", "content": system_prompt}]

            # 添加历史消息
            if message_history:
                messages.extend(message_history)

            # 添加当前用户消息
            messages.append({"role": "user", "content": user_message})

            logger.info(
                f"Chat request: user={user_id}, character={character.get('name')}, "
                f"model={model}, history_length={len(message_history or [])}"
            )

            # 调用 LLM
            response = await self.llm_client.chat_completion(
                messages=messages,
                model=model,
                user_id=user_id,
                temperature=temperature,
                max_tokens=max_tokens,
                stream=False,
            )

            # 提取响应
            assistant_message = response.choices[0].message.content
            tokens_used = response.usage.total_tokens if response.usage else 0

            logger.info(
                f"Chat response: tokens={tokens_used}, "
                f"response_length={len(assistant_message)}"
            )

            return {
                "response": assistant_message,
                "tokens_used": tokens_used,
                "model": response.model,
                "error": None,
            }

        except BudgetRejectionError as e:
            logger.warning(f"Budget rejection for user {user_id}: {e}")
            return {
                "response": None,
                "tokens_used": 0,
                "model": model,
                "error": "INSUFFICIENT_CREDITS",
            }
        except Exception as e:
            logger.error(f"Chat error: {type(e).__name__}: {e}")
            return {
                "response": None,
                "tokens_used": 0,
                "model": model,
                "error": str(e),
            }

    async def async_chat_stream(
        self,
        user_message: str,
        user_id: str,
        character: Dict[str, Any],
        message_history: Optional[List[Dict[str, str]]] = None,
        user_portrait: Optional[Dict[str, Any]] = None,
        conversation_context: Optional[Dict[str, Any]] = None,
        model: str = "gpt-4o-mini",
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
    ) -> AsyncIterator[str]:
        """异步对话（流式）.

        Args:
            user_message: 用户消息
            user_id: 用户ID
            character: 角色信息
            message_history: 消息历史
            user_portrait: 用户画像
            conversation_context: 对话上下文
            model: 模型标识
            temperature: 温度
            max_tokens: 最大 token 数

        Yields:
            文本片段（delta）
        """
        try:
            # 构建系统提示词
            system_prompt = self.prompt_builder.build_system_prompt(
                character=character,
                user_portrait=user_portrait,
                conversation_context=conversation_context,
            )

            # 准备消息列表
            messages = [{"role": "system", "content": system_prompt}]

            if message_history:
                messages.extend(message_history)

            messages.append({"role": "user", "content": user_message})

            logger.info(
                f"Chat stream request: user={user_id}, character={character.get('name')}, "
                f"model={model}"
            )

            # 调用流式 LLM
            stream = await self.llm_client.chat_completion(
                messages=messages,
                model=model,
                user_id=user_id,
                temperature=temperature,
                max_tokens=max_tokens,
                stream=True,
            )

            # 流式返回
            async for chunk in stream:
                if chunk.choices and chunk.choices[0].delta.content:
                    yield chunk.choices[0].delta.content

        except BudgetRejectionError as e:
            logger.warning(f"Budget rejection for user {user_id}: {e}")
            yield f"[ERROR: INSUFFICIENT_CREDITS]"
        except Exception as e:
            logger.error(f"Chat stream error: {type(e).__name__}: {e}")
            yield f"[ERROR: {str(e)}]"

    def _messages_to_openai_format(
        self, messages: List[BaseMessage]
    ) -> List[Dict[str, str]]:
        """将 LangChain 消息转换为 OpenAI 格式.

        Args:
            messages: LangChain 消息列表

        Returns:
            OpenAI 格式消息列表
        """
        openai_messages = []

        for msg in messages:
            if isinstance(msg, SystemMessage):
                role = "system"
            elif isinstance(msg, HumanMessage):
                role = "user"
            elif isinstance(msg, AIMessage):
                role = "assistant"
            else:
                logger.warning(f"Unknown message type: {type(msg)}")
                continue

            openai_messages.append({"role": role, "content": msg.content})

        return openai_messages


# 全局单例实例
_conversation_agent: Optional[ConversationAgent] = None


def get_conversation_agent() -> ConversationAgent:
    """获取对话代理单例.

    Returns:
        ConversationAgent 实例
    """
    global _conversation_agent
    if _conversation_agent is None:
        _conversation_agent = ConversationAgent()
    return _conversation_agent
