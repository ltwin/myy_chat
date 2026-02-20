// Package data LLM 客户端实现
package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/myy-chat/backend/app/conversation/internal/biz"
)

// llmClient LLM 服务客户端实现
// 通过 gRPC 调用 Python LLM Agent Service
type llmClient struct {
	log *log.Helper
	// TODO: 添加 gRPC 客户端连接
	// client llm_agent_v1.LLMAgentServiceClient
}

// NewLLMClient 创建 LLM 客户端
func NewLLMClient(logger log.Logger) biz.LLMClient {
	return &llmClient{
		log: log.NewHelper(logger),
	}
}

// Chat 发送聊天请求并获取响应
func (c *llmClient) Chat(ctx context.Context, messages []biz.ChatMessage, model string) (*biz.ChatResponse, error) {
	c.log.Infof("LLM Chat request: model=%s, messages=%d", model, len(messages))

	// TODO: 实现实际的 gRPC 调用
	// 临时返回占位响应
	return &biz.ChatResponse{
		Content:    "Hello! I'm a placeholder response. The actual LLM integration is pending.",
		TokenCount: 50,
		Model:      model,
		Cost:       0.001,
		LatencyMs:  100,
	}, nil
}
