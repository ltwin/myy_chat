"""Application settings configuration."""
from functools import lru_cache
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings."""

    # Service configuration
    service_name: str = "llm-agent-service"
    debug: bool = False

    # Server ports
    http_port: int = 8080
    grpc_port: int = 50051

    # LiteLLM Proxy configuration
    litellm_proxy_url: str = "http://localhost:4000"
    litellm_api_key: str = ""

    # Redis configuration
    redis_url: str = "redis://localhost:6379"

    # Database configuration (read-only for context)
    database_url: str = "postgresql://localhost:5432/myychat"

    # Tracing configuration
    otel_enabled: bool = False
    otel_endpoint: str = "http://localhost:4317"

    # Rate limiting
    rate_limit_requests: int = 100
    rate_limit_window: int = 60

    class Config:
        env_file = ".env"
        env_prefix = "LLM_AGENT_"


@lru_cache()
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()
