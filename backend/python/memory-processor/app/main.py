"""Memory Processor Service - Main Entry Point

This service provides:
- FastAPI endpoints for health checks and debugging
- gRPC service for memory extraction and retrieval
- Vector embedding generation and storage
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
    logger.info("Starting Memory Processor Service...")

    # Start gRPC server in background
    grpc_task = asyncio.create_task(start_grpc_server())

    logger.info(f"FastAPI server ready on port {settings.http_port}")
    logger.info(f"gRPC server ready on port {settings.grpc_port}")

    yield

    # Shutdown
    logger.info("Shutting down Memory Processor Service...")
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
    title="Memory Processor Service",
    description="Memory extraction, embedding, and retrieval service",
    version="0.1.0",
    lifespan=lifespan,
)

# Include routers
app.include_router(health.router, prefix="/api/v1", tags=["health"])


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "service": "memory-processor",
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
