// Package biz 消息业务逻辑层
package biz

import (
	"context"
	"errors"
	"time"
)

// 消息相关错误定义
var (
	ErrMessageNotFound       = errors.New("message not found")
	ErrInvalidMessageRole    = errors.New("invalid message role: must be user, assistant, or system")
	ErrInvalidMessageContent = errors.New("invalid message content: cannot be empty")
	ErrInvalidTokenCount     = errors.New("invalid token count: must be >= 0")
)

// MessageRole 消息角色类型
type MessageRole string

const (
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleSystem    MessageRole = "system"
)

// Message 消息实体
type Message struct {
	ID             int64           // 雪花ID
	ConversationID int64           // 会话ID
	Role           MessageRole     // 消息角色
	Content        string          // 消息内容
	TokenCount     int32           // Token数量
	Metadata       *MessageMetadata // 元数据
	CreatedAt      time.Time       // 创建时间
}

// MessageMetadata 消息元数据
type MessageMetadata struct {
	LatencyMs int32   // 响应延迟（毫秒）
	Model     string  // 使用的模型
	Cost      float64 // 成本（积分）
}

// NewMessage 创建新消息
func NewMessage(id, conversationID int64, role MessageRole, content string, tokenCount int32) (*Message, error) {
	// 验证角色
	if err := ValidateMessageRole(role); err != nil {
		return nil, err
	}

	// 验证内容
	if err := ValidateMessageContent(content); err != nil {
		return nil, err
	}

	// 验证Token数量
	if tokenCount < 0 {
		return nil, ErrInvalidTokenCount
	}

	return &Message{
		ID:             id,
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		TokenCount:     tokenCount,
		CreatedAt:      time.Now(),
	}, nil
}

// ValidateMessageRole 验证消息角色
func ValidateMessageRole(role MessageRole) error {
	switch role {
	case MessageRoleUser, MessageRoleAssistant, MessageRoleSystem:
		return nil
	default:
		return ErrInvalidMessageRole
	}
}

// ValidateMessageContent 验证消息内容
func ValidateMessageContent(content string) error {
	if content == "" {
		return ErrInvalidMessageContent
	}
	return nil
}

// SetMetadata 设置元数据
func (m *Message) SetMetadata(latencyMs int32, model string, cost float64) {
	m.Metadata = &MessageMetadata{
		LatencyMs: latencyMs,
		Model:     model,
		Cost:      cost,
	}
}

// IsUserMessage 检查是否是用户消息
func (m *Message) IsUserMessage() bool {
	return m.Role == MessageRoleUser
}

// IsAssistantMessage 检查是否是助手消息
func (m *Message) IsAssistantMessage() bool {
	return m.Role == MessageRoleAssistant
}

// MessageRepo 消息仓储接口
type MessageRepo interface {
	// Create 创建消息
	Create(ctx context.Context, message *Message) error
	// CreateBatch 批量创建消息（用于一次性保存用户和助手消息）
	CreateBatch(ctx context.Context, messages []*Message) error
	// GetByID 根据ID获取消息
	GetByID(ctx context.Context, id int64) (*Message, error)
	// ListByConversationID 列出会话的消息列表（分页，按时间倒序）
	ListByConversationID(ctx context.Context, conversationID, beforeMessageID int64, limit int32) ([]*Message, error)
	// CountByConversationID 统计会话的消息数量
	CountByConversationID(ctx context.Context, conversationID int64) (int32, error)
}
