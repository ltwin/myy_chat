"""LLM Client for LiteLLM Proxy integration."""
import logging
from typing import AsyncIterator, Optional

import httpx

from app.config.settings import get_settings

logger = logging.getLogger(__name__)
settings = get_settings()


class LLMClient:
    """Client for LiteLLM Proxy API.

    This client handles:
    - Chat completions (streaming and non-streaming)
    - Model selection
    - Rate limiting
    - Error handling
    """

    def __init__(self):
        """Initialize the LLM client."""
        self.base_url = settings.litellm_proxy_url
        self.api_key = settings.litellm_api_key
        self.timeout = httpx.Timeout(60.0, connect=10.0)

    async def chat_completion(
        self,
        messages: list[dict],
        model: str = "gpt-4o-mini",
        temperature: float = 0.7,
        max_tokens: Optional[int] = None,
        stream: bool = False,
    ) -> dict | AsyncIterator[str]:
        """Send a chat completion request.

        Args:
            messages: List of message dicts with role and content
            model: Model identifier (routed through LiteLLM)
            temperature: Sampling temperature (0.0 to 2.0)
            max_tokens: Maximum tokens in response
            stream: Whether to stream the response

        Returns:
            Response dict or async iterator of chunks
        """
        headers = {
            "Content-Type": "application/json",
        }
        if self.api_key:
            headers["Authorization"] = f"Bearer {self.api_key}"

        payload = {
            "model": model,
            "messages": messages,
            "temperature": temperature,
            "stream": stream,
        }
        if max_tokens:
            payload["max_tokens"] = max_tokens

        if stream:
            return self._stream_completion(headers, payload)
        else:
            return await self._sync_completion(headers, payload)

    async def _sync_completion(
        self, headers: dict, payload: dict
    ) -> dict:
        """Non-streaming chat completion."""
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            response = await client.post(
                f"{self.base_url}/v1/chat/completions",
                headers=headers,
                json=payload,
            )
            response.raise_for_status()
            return response.json()

    async def _stream_completion(
        self, headers: dict, payload: dict
    ) -> AsyncIterator[str]:
        """Streaming chat completion."""
        async with httpx.AsyncClient(timeout=self.timeout) as client:
            async with client.stream(
                "POST",
                f"{self.base_url}/v1/chat/completions",
                headers=headers,
                json=payload,
            ) as response:
                response.raise_for_status()
                async for line in response.aiter_lines():
                    if line.startswith("data: "):
                        data = line[6:]
                        if data == "[DONE]":
                            break
                        yield data

    async def health_check(self) -> bool:
        """Check if LiteLLM Proxy is healthy.

        Returns:
            True if healthy, False otherwise
        """
        try:
            async with httpx.AsyncClient(timeout=httpx.Timeout(5.0)) as client:
                response = await client.get(f"{self.base_url}/health")
                return response.status_code == 200
        except Exception as e:
            logger.warning(f"LiteLLM health check failed: {e}")
            return False


# Singleton instance
_llm_client: Optional[LLMClient] = None


def get_llm_client() -> LLMClient:
    """Get the LLM client singleton."""
    global _llm_client
    if _llm_client is None:
        _llm_client = LLMClient()
    return _llm_client
