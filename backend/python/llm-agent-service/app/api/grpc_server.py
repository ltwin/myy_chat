"""gRPC Server implementation for LLM Agent Service.

此模块实现 gRPC 服务接口，提供：
- ChatCompletion RPC（非流式对话）
- StreamChatCompletion RPC（流式对话）
- HealthCheck RPC

注意：需要先运行 generate_proto.sh 生成 proto 代码。
"""
import asyncio
import logging
from typing import AsyncIterator, Dict, Any, List, Optional

import grpc

from app.services.agent import get_conversation_agent
from app.core.llm_proxy_client import get_llm_proxy_client

logger = logging.getLogger(__name__)

# TODO: 在运行 generate_proto.sh 后取消注释以下导入
# from app.api.generated import llm_agent_pb2, llm_agent_pb2_grpc


class LLMAgentServicer:
    """LLM Agent Service gRPC 实现.

    提供 AI 对话能力的 gRPC 接口。
    """

    def __init__(self):
        """初始化 servicer."""
        self.agent = get_conversation_agent()
        self.llm_client = get_llm_proxy_client()

    async def ChatCompletion(self, request, context):
        """非流式对话 RPC.

        Args:
            request: ChatCompletionRequest
            context: gRPC context

        Returns:
            ChatCompletionResponse
        """
        try:
            # 提取请求参数
            user_id = request.user_id
            user_message = request.user_message

            # 转换角色信息
            character = self._proto_character_to_dict(request.character)

            # 转换消息历史
            message_history = self._proto_messages_to_list(request.message_history)

            # 转换用户画像
            user_portrait = (
                self._proto_user_portrait_to_dict(request.user_portrait)
                if request.HasField("user_portrait")
                else None
            )

            # 转换对话上下文
            conversation_context = (
                self._proto_conversation_context_to_dict(request.conversation_context)
                if request.HasField("conversation_context")
                else None
            )

            model = request.model or "gpt-4o-mini"
            temperature = request.temperature or 0.7
            max_tokens = request.max_tokens if request.max_tokens > 0 else None

            logger.info(
                f"ChatCompletion request: user={user_id}, "
                f"character={character.get('name')}, model={model}"
            )

            # 调用 agent
            result = await self.agent.async_chat(
                user_message=user_message,
                user_id=user_id,
                character=character,
                message_history=message_history,
                user_portrait=user_portrait,
                conversation_context=conversation_context,
                model=model,
                temperature=temperature,
                max_tokens=max_tokens,
            )

            # 构建响应（TODO: 使用生成的 proto 类）
            # response = llm_agent_pb2.ChatCompletionResponse(
            #     response=result.get("response", ""),
            #     tokens_used=result.get("tokens_used", 0),
            #     model=result.get("model", model),
            #     error=result.get("error", ""),
            # )

            logger.info(
                f"ChatCompletion response: tokens={result.get('tokens_used')}, "
                f"error={result.get('error')}"
            )

            # return response
            raise NotImplementedError(
                "ChatCompletion requires proto generation. Run generate_proto.sh first."
            )

        except Exception as e:
            logger.error(f"ChatCompletion error: {type(e).__name__}: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            raise

    async def StreamChatCompletion(self, request, context):
        """流式对话 RPC.

        Args:
            request: ChatCompletionRequest
            context: gRPC context

        Yields:
            ChatCompletionChunk
        """
        try:
            # 提取请求参数（与 ChatCompletion 相同）
            user_id = request.user_id
            user_message = request.user_message
            character = self._proto_character_to_dict(request.character)
            message_history = self._proto_messages_to_list(request.message_history)
            user_portrait = (
                self._proto_user_portrait_to_dict(request.user_portrait)
                if request.HasField("user_portrait")
                else None
            )
            conversation_context = (
                self._proto_conversation_context_to_dict(request.conversation_context)
                if request.HasField("conversation_context")
                else None
            )
            model = request.model or "gpt-4o-mini"
            temperature = request.temperature or 0.7
            max_tokens = request.max_tokens if request.max_tokens > 0 else None

            logger.info(
                f"StreamChatCompletion request: user={user_id}, "
                f"character={character.get('name')}, model={model}"
            )

            # 调用流式 agent
            stream = self.agent.async_chat_stream(
                user_message=user_message,
                user_id=user_id,
                character=character,
                message_history=message_history,
                user_portrait=user_portrait,
                conversation_context=conversation_context,
                model=model,
                temperature=temperature,
                max_tokens=max_tokens,
            )

            # 流式返回
            async for delta in stream:
                # chunk = llm_agent_pb2.ChatCompletionChunk(
                #     delta=delta,
                #     is_final=False,
                #     error="",
                # )
                # yield chunk
                pass

            # 发送最后一个空片段标记结束
            # final_chunk = llm_agent_pb2.ChatCompletionChunk(
            #     delta="",
            #     is_final=True,
            #     error="",
            # )
            # yield final_chunk

            logger.info(f"StreamChatCompletion completed for user={user_id}")

            raise NotImplementedError(
                "StreamChatCompletion requires proto generation. Run generate_proto.sh first."
            )

        except Exception as e:
            logger.error(f"StreamChatCompletion error: {type(e).__name__}: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            raise

    async def HealthCheck(self, request, context):
        """健康检查 RPC.

        Args:
            request: HealthCheckRequest
            context: gRPC context

        Returns:
            HealthCheckResponse
        """
        try:
            # 检查 LiteLLM Proxy 连接
            llm_healthy = await self.llm_client.health_check()

            # response = llm_agent_pb2.HealthCheckResponse(
            #     healthy=llm_healthy,
            #     message="OK" if llm_healthy else "LiteLLM Proxy unreachable",
            # )
            # return response

            raise NotImplementedError(
                "HealthCheck requires proto generation. Run generate_proto.sh first."
            )

        except Exception as e:
            logger.error(f"HealthCheck error: {type(e).__name__}: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            raise

    def _proto_character_to_dict(self, character) -> Dict[str, Any]:
        """转换 proto Character 为字典.

        Args:
            character: proto Character 对象

        Returns:
            角色信息字典
        """
        return {
            "name": character.name,
            "personality": character.personality,
            "background": character.background,
            "speaking_style": character.speaking_style,
            "world_view": character.world_view,
            "moderation_style": character.moderation_style or "default",
        }

    def _proto_messages_to_list(self, messages) -> List[Dict[str, str]]:
        """转换 proto Message 列表为字典列表.

        Args:
            messages: proto Message 列表

        Returns:
            消息字典列表
        """
        return [{"role": msg.role, "content": msg.content} for msg in messages]

    def _proto_user_portrait_to_dict(self, user_portrait) -> Optional[Dict[str, Any]]:
        """转换 proto UserPortrait 为字典.

        Args:
            user_portrait: proto UserPortrait 对象

        Returns:
            用户画像字典
        """
        if not user_portrait:
            return None

        return {
            "name": user_portrait.name,
            "age": user_portrait.age if user_portrait.age > 0 else None,
            "interests": list(user_portrait.interests),
            "preferences": user_portrait.preferences,
        }

    def _proto_conversation_context_to_dict(
        self, context
    ) -> Optional[Dict[str, Any]]:
        """转换 proto ConversationContext 为字典.

        Args:
            context: proto ConversationContext 对象

        Returns:
            对话上下文字典
        """
        if not context:
            return None

        return {
            "relation_type": context.relation_type,
            "interaction_mode": context.interaction_mode,
        }


async def serve_grpc(port: int):
    """启动 gRPC 服务器.

    Args:
        port: 监听端口
    """
    server = grpc.aio.server()

    # TODO: 在运行 generate_proto.sh 后取消注释以下行
    # llm_agent_pb2_grpc.add_LLMAgentServiceServicer_to_server(
    #     LLMAgentServicer(), server
    # )

    listen_addr = f"[::]:{port}"
    server.add_insecure_port(listen_addr)

    logger.info(f"Starting gRPC server on {listen_addr}")
    await server.start()

    try:
        await server.wait_for_termination()
    except asyncio.CancelledError:
        logger.info("Shutting down gRPC server...")
        await server.stop(grace=5)
