"""LLM Agent Service - Main Entry Point

This service provides:
- FastAPI endpoints for health checks and debugging
- gRPC service for LLM conversation handling
- Integration with LiteLLM Proxy for multi-model support
"""
import asyncio
import logging
from contextlib import asynccontextmanager
from concurrent import futures

import grpc
from fastapi import FastAPI

from app.config.settings import get_settings
from app.api import health
from app.grpc.server import serve_grpc

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

settings = get_settings()


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application lifespan manager."""
    logger.info("Starting LLM Agent Service...")

    # Start gRPC server in background
    grpc_task = asyncio.create_task(start_grpc_server())

    logger.info(f"FastAPI server ready on port {settings.http_port}")
    logger.info(f"gRPC server ready on port {settings.grpc_port}")

    yield

    # Shutdown
    logger.info("Shutting down LLM Agent Service...")
    grpc_task.cancel()
    try:
        await grpc_task
    except asyncio.CancelledError:
        pass


async def start_grpc_server():
    """Start gRPC server."""
    await serve_grpc(settings.grpc_port)


# Create FastAPI app
app = FastAPI(
    title="LLM Agent Service",
    description="AI conversation agent with LLM integration",
    version="0.1.0",
    lifespan=lifespan,
)

# Include routers
app.include_router(health.router, prefix="/api/v1", tags=["health"])


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "service": "llm-agent-service",
        "version": "0.1.0",
        "status": "running"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=settings.http_port,
        reload=settings.debug,
    )
