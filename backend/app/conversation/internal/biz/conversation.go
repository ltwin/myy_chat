// Package biz 会话业务逻辑层
package biz

import (
	"context"
	"errors"
	"time"
)

// 会话相关错误定义
var (
	ErrConversationNotFound      = errors.New("conversation not found")
	ErrConversationAccessDenied  = errors.New("conversation access denied")
	ErrConversationArchived      = errors.New("conversation is archived")
	ErrInvalidConversationTitle  = errors.New("invalid conversation title: must be 1-255 characters")
	ErrInvalidCharacterID        = errors.New("invalid character id")
)

// Conversation 会话实体
type Conversation struct {
	ID            int64     // 雪花ID
	UserID        int64     // 用户ID
	CharacterID   int64     // 角色ID
	Title         string    // 会话标题
	MessageCount  int32     // 消息数量
	TokenCount    int32     // Token消耗总数
	StartedAt     time.Time // 会话开始时间
	LastMessageAt time.Time // 最后消息时间
	IsArchived    bool      // 是否归档
}

// NewConversation 创建新会话
func NewConversation(id, userID, characterID int64, title string) (*Conversation, error) {
	if characterID <= 0 {
		return nil, ErrInvalidCharacterID
	}

	// 如果未提供标题，使用默认值
	if title == "" {
		title = "New Conversation"
	}

	// 验证标题长度
	if err := ValidateConversationTitle(title); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Conversation{
		ID:            id,
		UserID:        userID,
		CharacterID:   characterID,
		Title:         title,
		MessageCount:  0,
		TokenCount:    0,
		StartedAt:     now,
		LastMessageAt: now,
		IsArchived:    false,
	}, nil
}

// ValidateConversationTitle 验证会话标题
func ValidateConversationTitle(title string) error {
	if len(title) < 1 || len(title) > 255 {
		return ErrInvalidConversationTitle
	}
	return nil
}

// UpdateTitle 更新会话标题
func (c *Conversation) UpdateTitle(title string) error {
	if err := ValidateConversationTitle(title); err != nil {
		return err
	}
	c.Title = title
	return nil
}

// IncrementMessageCount 增加消息计数
func (c *Conversation) IncrementMessageCount(tokenCount int32) {
	c.MessageCount++
	c.TokenCount += tokenCount
	c.LastMessageAt = time.Now()
}

// Archive 归档会话
func (c *Conversation) Archive() {
	c.IsArchived = true
}

// Unarchive 取消归档会话
func (c *Conversation) Unarchive() {
	c.IsArchived = false
}

// CanSendMessage 检查是否可以发送消息
func (c *Conversation) CanSendMessage() error {
	if c.IsArchived {
		return ErrConversationArchived
	}
	return nil
}

// BelongsToUser 检查会话是否属于指定用户
func (c *Conversation) BelongsToUser(userID int64) bool {
	return c.UserID == userID
}

// ConversationRepo 会话仓储接口
type ConversationRepo interface {
	// Create 创建会话
	Create(ctx context.Context, conversation *Conversation) error
	// GetByID 根据ID获取会话
	GetByID(ctx context.Context, id int64) (*Conversation, error)
	// GetByIDAndUserID 根据ID和用户ID获取会话（权限检查）
	GetByIDAndUserID(ctx context.Context, id, userID int64) (*Conversation, error)
	// List 列出用户的会话列表（分页）
	List(ctx context.Context, userID int64, characterID int64, archived bool, page, pageSize int32) ([]*Conversation, int32, error)
	// Update 更新会话
	Update(ctx context.Context, conversation *Conversation) error
	// IncrementCounts 原子性增加消息计数和Token计数
	IncrementCounts(ctx context.Context, conversationID int64, tokenCount int32) error
}
