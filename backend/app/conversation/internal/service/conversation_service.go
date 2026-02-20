// Package service gRPC 服务层
package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/myy-chat/backend/api/conversation/v1"
	"github.com/myy-chat/backend/app/conversation/internal/biz"
	"github.com/myy-chat/backend/pkg/middleware"
)

// ConversationService gRPC 会话服务实现
type ConversationService struct {
	pb.UnimplementedConversationServiceServer

	conversationService *biz.ConversationService
}

// NewConversationService 创建 gRPC 会话服务
func NewConversationService(cs *biz.ConversationService) *ConversationService {
	return &ConversationService{
		conversationService: cs,
	}
}

// CreateConversation 创建新会话
func (s *ConversationService) CreateConversation(ctx context.Context, req *pb.CreateConversationRequest) (*pb.CreateConversationResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	conversation, err := s.conversationService.CreateConversation(ctx, biz.CreateConversationInput{
		UserID:      userID,
		CharacterID: req.CharacterId,
		Title:       req.Title,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateConversationResponse{
		Conversation: conversationToProto(conversation),
	}, nil
}

// GetConversation 获取会话详情
func (s *ConversationService) GetConversation(ctx context.Context, req *pb.GetConversationRequest) (*pb.GetConversationResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	conversation, err := s.conversationService.GetConversation(ctx, req.ConversationId, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetConversationResponse{
		Conversation: conversationToProto(conversation),
	}, nil
}

// ListConversations 列出用户的会话列表
func (s *ConversationService) ListConversations(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	output, err := s.conversationService.ListConversations(ctx, biz.ListConversationsInput{
		UserID:      userID,
		CharacterID: req.CharacterId,
		Archived:    req.Archived,
		Page:        req.Page,
		PageSize:    req.PageSize,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	conversations := make([]*pb.Conversation, len(output.Conversations))
	for i, conv := range output.Conversations {
		conversations[i] = conversationToProto(conv)
	}

	return &pb.ListConversationsResponse{
		Conversations: conversations,
		Total:         output.Total,
		Page:          output.Page,
		PageSize:      output.PageSize,
	}, nil
}

// SendMessage 发送消息并获取AI响应
func (s *ConversationService) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	output, err := s.conversationService.SendMessage(ctx, biz.SendMessageInput{
		UserID:         userID,
		ConversationID: req.ConversationId,
		Content:        req.Content,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.SendMessageResponse{
		UserMessage:      messageToProto(output.UserMessage),
		AssistantMessage: messageToProto(output.AssistantMessage),
		TokensUsed:       output.TokensUsed,
		CreditsCharged:   output.CreditsCharged,
	}, nil
}

// GetMessages 获取会话消息列表（分页）
func (s *ConversationService) GetMessages(ctx context.Context, req *pb.GetMessagesRequest) (*pb.GetMessagesResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	output, err := s.conversationService.GetMessages(ctx, biz.GetMessagesInput{
		UserID:          userID,
		ConversationID:  req.ConversationId,
		BeforeMessageID: req.BeforeMessageId,
		Limit:           req.Limit,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	messages := make([]*pb.Message, len(output.Messages))
	for i, msg := range output.Messages {
		messages[i] = messageToProto(msg)
	}

	return &pb.GetMessagesResponse{
		Messages: messages,
		HasMore:  output.HasMore,
	}, nil
}

// ArchiveConversation 归档会话
func (s *ConversationService) ArchiveConversation(ctx context.Context, req *pb.ArchiveConversationRequest) (*pb.ArchiveConversationResponse, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.conversationService.ArchiveConversation(ctx, req.ConversationId, userID, req.Archived)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ArchiveConversationResponse{
		Success: true,
	}, nil
}

// 辅助函数

// getUserIDFromContext 从上下文获取用户ID (通过 JWT 中间件注入)
func getUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "authentication required")
	}
	if userID <= 0 {
		return 0, status.Error(codes.InvalidArgument, "invalid user id")
	}
	return userID, nil
}

// toGRPCError 将业务错误转换为 gRPC 状态错误
func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// 已经是 gRPC status error
	if _, ok := status.FromError(err); ok {
		return err
	}

	// 根据错误类型转换
	switch {
	case errors.Is(err, biz.ErrConversationNotFound),
		errors.Is(err, biz.ErrMessageNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrConversationAccessDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, biz.ErrConversationArchived):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, biz.ErrInsufficientCredits):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, biz.ErrInvalidConversationTitle),
		errors.Is(err, biz.ErrInvalidCharacterID),
		errors.Is(err, biz.ErrInvalidMessageRole),
		errors.Is(err, biz.ErrInvalidMessageContent):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, biz.ErrLLMServiceFailure):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// conversationToProto 将会话实体转换为 protobuf 消息
func conversationToProto(conv *biz.Conversation) *pb.Conversation {
	if conv == nil {
		return nil
	}

	return &pb.Conversation{
		Id:            conv.ID,
		UserId:        conv.UserID,
		CharacterId:   conv.CharacterID,
		Title:         conv.Title,
		MessageCount:  conv.MessageCount,
		TokenCount:    conv.TokenCount,
		StartedAt:     timestamppb.New(conv.StartedAt),
		LastMessageAt: timestamppb.New(conv.LastMessageAt),
		IsArchived:    conv.IsArchived,
	}
}

// messageToProto 将消息实体转换为 protobuf 消息
func messageToProto(msg *biz.Message) *pb.Message {
	if msg == nil {
		return nil
	}

	protoMsg := &pb.Message{
		Id:             msg.ID,
		ConversationId: msg.ConversationID,
		Role:           string(msg.Role),
		Content:        msg.Content,
		TokenCount:     msg.TokenCount,
		CreatedAt:      timestamppb.New(msg.CreatedAt),
	}

	if msg.Metadata != nil {
		protoMsg.Metadata = &pb.MessageMetadata{
			LatencyMs: msg.Metadata.LatencyMs,
			Model:     msg.Metadata.Model,
			Cost:      msg.Metadata.Cost,
		}
	}

	return protoMsg
}
