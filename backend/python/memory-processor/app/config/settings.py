"""Application settings configuration."""
from functools import lru_cache
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """Application settings."""

    # Service configuration
    service_name: str = "memory-processor"
    debug: bool = False

    # Server ports
    http_port: int = 8081
    grpc_port: int = 50052

    # Database configuration
    database_url: str = "postgresql://localhost:5432/myychat"

    # Redis configuration
    redis_url: str = "redis://localhost:6379"

    # Embedding model configuration
    embedding_model: str = "text-embedding-3-small"
    embedding_dim: int = 1536

    # LiteLLM Proxy for embedding
    litellm_proxy_url: str = "http://localhost:4000"
    litellm_api_key: str = ""

    # Neo4j configuration (for knowledge graph)
    neo4j_uri: str = "bolt://localhost:7687"
    neo4j_user: str = "neo4j"
    neo4j_password: str = ""

    # Processing configuration
    batch_size: int = 32
    extraction_timeout: int = 30

    # Tracing configuration
    otel_enabled: bool = False
    otel_endpoint: str = "http://localhost:4317"

    class Config:
        env_file = ".env"
        env_prefix = "MEMORY_PROC_"


@lru_cache()
def get_settings() -> Settings:
    """Get cached settings instance."""
    return Settings()
