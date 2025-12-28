"""gRPC Server implementation."""
import asyncio
import logging
from concurrent import futures

import grpc

# TODO: Import generated proto files after proto generation
# from proto.conversation.v1 import conversation_pb2
# from proto.conversation.v1 import conversation_pb2_grpc

logger = logging.getLogger(__name__)


class ConversationServicer:
    """gRPC Conversation service implementation.

    This service handles:
    - Chat message processing
    - Streaming responses
    - Context management
    """

    async def SendMessage(self, request, context):
        """Process a chat message and return response.

        Args:
            request: ChatRequest with user message
            context: gRPC context

        Returns:
            ChatResponse with AI response
        """
        # TODO: Implement after proto generation
        # response = conversation_pb2.ChatResponse()
        # response.message = "Hello from LLM Agent Service"
        # return response
        raise NotImplementedError("SendMessage not yet implemented")

    async def StreamMessage(self, request, context):
        """Process a chat message and stream response.

        Args:
            request: ChatRequest with user message
            context: gRPC context

        Yields:
            MessageChunk with partial response
        """
        # TODO: Implement streaming response
        raise NotImplementedError("StreamMessage not yet implemented")


async def serve_grpc(port: int):
    """Start the gRPC server.

    Args:
        port: Port number to listen on
    """
    server = grpc.aio.server(
        futures.ThreadPoolExecutor(max_workers=10),
        options=[
            ("grpc.max_send_message_length", 50 * 1024 * 1024),  # 50MB
            ("grpc.max_receive_message_length", 50 * 1024 * 1024),  # 50MB
        ],
    )

    # TODO: Register service after proto generation
    # conversation_pb2_grpc.add_ConversationServiceServicer_to_server(
    #     ConversationServicer(), server
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
