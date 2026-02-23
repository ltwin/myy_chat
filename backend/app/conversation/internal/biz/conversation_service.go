// Package biz 会话业务服务层
package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/myy-chat/backend/pkg/snowflake"
)

// 业务错误
var (
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrLLMServiceFailure   = errors.New("llm service failure")
)

// LLMClient LLM服务客户端接口
type LLMClient interface {
	// Chat 发送聊天请求并获取响应
	Chat(ctx context.Context, messages []ChatMessage, model string) (*ChatResponse, error)
}

// ChatMessage LLM聊天消息
type ChatMessage struct {
	Role    string // user, assistant, system
	Content string
}

// ChatResponse LLM聊天响应
type ChatResponse struct {
	Content    string  // AI响应内容
	TokenCount int32   // 使用的Token数
	Model      string  // 使用的模型
	Cost       float64 // 消耗的积分
	LatencyMs  int32   // 响应延迟（毫秒）
}

// CreditService 积分服务接口
type CreditService interface {
	// DeductCredits 扣除积分
	DeductCredits(ctx context.Context, userID int64, amount float64, reason string) error
	// GetBalance 获取积分余额
	GetBalance(ctx context.Context, userID int64) (float64, error)
}

// ConversationService 会话业务服务
type ConversationService struct {
	conversationRepo ConversationRepo
	messageRepo      MessageRepo
	llmClient        LLMClient
	creditService    CreditService
	idGen            snowflake.Generator
	log              *log.Helper
}

// NewConversationService 创建会话服务
func NewConversationService(
	conversationRepo ConversationRepo,
	messageRepo MessageRepo,
	llmClient LLMClient,
	creditService CreditService,
	idGen snowflake.Generator,
	logger log.Logger,
) *ConversationService {
	return &ConversationService{
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		llmClient:        llmClient,
		creditService:    creditService,
		idGen:            idGen,
		log:              log.NewHelper(logger),
	}
}

// CreateConversationInput 创建会话输入
type CreateConversationInput struct {
	UserID      int64
	CharacterID int64
	Title       string
}

// CreateConversation 创建新会话
func (s *ConversationService) CreateConversation(ctx context.Context, input CreateConversationInput) (*Conversation, error) {
	// 生成雪花ID
	conversationID := s.idGen.Generate()

	// 创建会话实体
	conversation, err := NewConversation(conversationID, input.UserID, input.CharacterID, input.Title)
	if err != nil {
		return nil, err
	}

	// 保存会话
	if err := s.conversationRepo.Create(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}

	s.log.Infof("conversation created: id=%d, user_id=%d, character_id=%d", conversationID, input.UserID, input.CharacterID)
	return conversation, nil
}

// GetConversation 获取会话
func (s *ConversationService) GetConversation(ctx context.Context, conversationID, userID int64) (*Conversation, error) {
	conversation, err := s.conversationRepo.GetByIDAndUserID(ctx, conversationID, userID)
	if err != nil {
		return nil, err
	}
	return conversation, nil
}

// ListConversationsInput 列出会话输入
type ListConversationsInput struct {
	UserID      int64
	CharacterID int64 // 0表示不过滤
	Archived    bool
	Page        int32
	PageSize    int32
}

// ListConversationsOutput 列出会话输出
type ListConversationsOutput struct {
	Conversations []*Conversation
	Total         int32
	Page          int32
	PageSize      int32
}

// ListConversations 列出用户的会话
func (s *ConversationService) ListConversations(ctx context.Context, input ListConversationsInput) (*ListConversationsOutput, error) {
	// 默认分页参数
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	conversations, total, err := s.conversationRepo.List(ctx, input.UserID, input.CharacterID, input.Archived, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}

	return &ListConversationsOutput{
		Conversations: conversations,
		Total:         total,
		Page:          page,
		PageSize:      pageSize,
	}, nil
}

// SendMessageInput 发送消息输入
type SendMessageInput struct {
	UserID         int64
	ConversationID int64
	Content        string
}

// SendMessageOutput 发送消息输出
type SendMessageOutput struct {
	UserMessage      *Message
	AssistantMessage *Message
	TokensUsed       int32
	CreditsCharged   float64
}

// SendMessage 发送消息并获取AI响应
func (s *ConversationService) SendMessage(ctx context.Context, input SendMessageInput) (*SendMessageOutput, error) {
	// 获取会话并验证权限
	conversation, err := s.conversationRepo.GetByIDAndUserID(ctx, input.ConversationID, input.UserID)
	if err != nil {
		return nil, err
	}

	// 检查是否可以发送消息
	if err := conversation.CanSendMessage(); err != nil {
		return nil, err
	}

	// 检查积分余额（预估）
	// TODO: 实现更精确的积分预估
	balance, err := s.creditService.GetBalance(ctx, input.UserID)
	if err != nil {
		s.log.Warnf("failed to get credit balance: %v", err)
	} else if balance < 0.01 { // 至少需要0.01积分
		return nil, ErrInsufficientCredits
	}

	// 获取历史消息（最近10条，用于上下文）
	historyMessages, err := s.messageRepo.ListByConversationID(ctx, input.ConversationID, 0, 10)
	if err != nil {
		s.log.Warnf("failed to get history messages: %v", err)
		historyMessages = []*Message{}
	}

	// 构建LLM请求
	chatMessages := s.buildChatMessages(historyMessages, input.Content)

	// 调用LLM服务
	startTime := time.Now()
	llmResp, err := s.llmClient.Chat(ctx, chatMessages, "gpt-4o") // TODO: 从角色配置获取模型
	if err != nil {
		s.log.Errorf("llm service error: %v", err)
		return nil, ErrLLMServiceFailure
	}
	latencyMs := int32(time.Since(startTime).Milliseconds())

	// 生成消息ID
	userMessageID := s.idGen.Generate()
	assistantMessageID := s.idGen.Generate()

	// 创建用户消息
	userMessage, err := NewMessage(userMessageID, input.ConversationID, MessageRoleUser, input.Content, 0)
	if err != nil {
		return nil, err
	}

	// 创建助手消息
	assistantMessage, err := NewMessage(assistantMessageID, input.ConversationID, MessageRoleAssistant, llmResp.Content, llmResp.TokenCount)
	if err != nil {
		return nil, err
	}
	assistantMessage.SetMetadata(latencyMs, llmResp.Model, llmResp.Cost)

	// 批量保存消息
	if err := s.messageRepo.CreateBatch(ctx, []*Message{userMessage, assistantMessage}); err != nil {
		return nil, fmt.Errorf("failed to save messages: %w", err)
	}

	// 更新会话统计（增加消息计数和Token计数）
	if err := s.conversationRepo.IncrementCounts(ctx, input.ConversationID, llmResp.TokenCount); err != nil {
		s.log.Warnf("failed to update conversation counts: %v", err)
	}

	// 扣除积分
	if llmResp.Cost > 0 {
		creditReason := fmt.Sprintf("conversation:%d:assistant_message:%d", input.ConversationID, assistantMessageID)
		if err := s.creditService.DeductCredits(ctx, input.UserID, llmResp.Cost, creditReason); err != nil {
			s.log.Errorf("failed to deduct credits: %v", err)
			// 非致命错误，继续
		}
	}

	s.log.Infof("message sent: conversation_id=%d, tokens=%d, cost=%.4f, latency=%dms",
		input.ConversationID, llmResp.TokenCount, llmResp.Cost, latencyMs)

	return &SendMessageOutput{
		UserMessage:      userMessage,
		AssistantMessage: assistantMessage,
		TokensUsed:       llmResp.TokenCount,
		CreditsCharged:   llmResp.Cost,
	}, nil
}

// GetMessagesInput 获取消息输入
type GetMessagesInput struct {
	UserID          int64
	ConversationID  int64
	BeforeMessageID int64
	Limit           int32
}

// GetMessagesOutput 获取消息输出
type GetMessagesOutput struct {
	Messages []*Message
	HasMore  bool
}

// GetMessages 获取会话消息列表
func (s *ConversationService) GetMessages(ctx context.Context, input GetMessagesInput) (*GetMessagesOutput, error) {
	// 验证会话权限
	_, err := s.conversationRepo.GetByIDAndUserID(ctx, input.ConversationID, input.UserID)
	if err != nil {
		return nil, err
	}

	// 默认限制
	limit := input.Limit
	if limit < 1 || limit > 100 {
		limit = 50
	}

	// 获取消息（多获取1条用于判断是否有更多）
	messages, err := s.messageRepo.ListByConversationID(ctx, input.ConversationID, input.BeforeMessageID, limit+1)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// 判断是否有更多消息
	hasMore := len(messages) > int(limit)
	if hasMore {
		messages = messages[:limit]
	}

	return &GetMessagesOutput{
		Messages: messages,
		HasMore:  hasMore,
	}, nil
}

// ArchiveConversation 归档/取消归档会话
func (s *ConversationService) ArchiveConversation(ctx context.Context, conversationID, userID int64, archived bool) error {
	// 获取会话并验证权限
	conversation, err := s.conversationRepo.GetByIDAndUserID(ctx, conversationID, userID)
	if err != nil {
		return err
	}

	// 更新归档状态
	if archived {
		conversation.Archive()
	} else {
		conversation.Unarchive()
	}

	// 保存
	if err := s.conversationRepo.Update(ctx, conversation); err != nil {
		return fmt.Errorf("failed to update conversation: %w", err)
	}

	s.log.Infof("conversation archived: id=%d, archived=%v", conversationID, archived)
	return nil
}

// buildChatMessages 构建LLM聊天消息列表
func (s *ConversationService) buildChatMessages(historyMessages []*Message, newContent string) []ChatMessage {
	// 反转历史消息（从旧到新）
	var chatMessages []ChatMessage
	for i := len(historyMessages) - 1; i >= 0; i-- {
		msg := historyMessages[i]
		chatMessages = append(chatMessages, ChatMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	// 添加新用户消息
	chatMessages = append(chatMessages, ChatMessage{
		Role:    string(MessageRoleUser),
		Content: newContent,
	})

	return chatMessages
}
