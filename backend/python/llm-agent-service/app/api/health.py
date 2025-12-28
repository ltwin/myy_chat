"""Health check endpoints."""
from fastapi import APIRouter
from pydantic import BaseModel

router = APIRouter()


class HealthResponse(BaseModel):
    """Health check response."""
    status: str
    service: str
    version: str


class ReadinessResponse(BaseModel):
    """Readiness check response."""
    ready: bool
    checks: dict


@router.get("/health", response_model=HealthResponse)
async def health_check():
    """Basic health check endpoint."""
    return HealthResponse(
        status="healthy",
        service="llm-agent-service",
        version="0.1.0"
    )


@router.get("/ready", response_model=ReadinessResponse)
async def readiness_check():
    """Readiness check with dependency verification."""
    checks = {
        "grpc_server": True,  # TODO: Implement actual check
        "litellm_proxy": True,  # TODO: Implement actual check
    }
    ready = all(checks.values())
    return ReadinessResponse(ready=ready, checks=checks)


@router.get("/live")
async def liveness_check():
    """Kubernetes liveness probe."""
    return {"alive": True}
