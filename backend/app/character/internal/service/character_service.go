// Package service gRPC 服务层
package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/myy-chat/backend/golang/api/character/v1"
	"github.com/myy-chat/backend/golang/app/character/internal/biz"
	"github.com/myy-chat/backend/golang/pkg/middleware"
)

// CharacterService gRPC 角色服务实现
type CharacterService struct {
	pb.UnimplementedCharacterServiceServer

	characterService *biz.CharacterService
}

// NewCharacterService 创建 gRPC 角色服务
func NewCharacterService(cs *biz.CharacterService) *CharacterService {
	return &CharacterService{
		characterService: cs,
	}
}

// ListCharacters 列出角色 (公开角色 + 用户自己的角色)
func (s *CharacterService) ListCharacters(ctx context.Context, req *pb.ListCharactersRequest) (*pb.ListCharactersResponse, error) {
	// 从 context 获取用户 ID (未登录为0，可以看公开角色)
	userID := getOptionalUserIDFromContext(ctx)

	// 设置默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	filter := req.Filter
	if filter == "" {
		filter = "all"
	}

	output, err := s.characterService.ListCharacters(ctx, biz.ListCharactersInput{
		UserID:   userID,
		Filter:   filter,
		Page:     int(page),
		PageSize: int(pageSize),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	characters := make([]*pb.Character, len(output.Characters))
	for i, char := range output.Characters {
		characters[i] = characterToProto(char)
	}

	return &pb.ListCharactersResponse{
		Characters: characters,
		Total:      int32(output.Total),
		Page:       int32(output.Page),
		PageSize:   int32(output.PageSize),
	}, nil
}

// GetCharacter 获取角色详情
func (s *CharacterService) GetCharacter(ctx context.Context, req *pb.GetCharacterRequest) (*pb.GetCharacterResponse, error) {
	// 未登录用户可以查看公开角色
	userID := getOptionalUserIDFromContext(ctx)

	character, err := s.characterService.GetCharacter(ctx, req.Id, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.GetCharacterResponse{
		Character: characterToProto(character),
	}, nil
}

// CreateCharacter 创建角色 (需要登录)
func (s *CharacterService) CreateCharacter(ctx context.Context, req *pb.CreateCharacterRequest) (*pb.CreateCharacterResponse, error) {
	userID, err := getRequiredUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	character, err := s.characterService.CreateCharacter(ctx, biz.CreateCharacterInput{
		UserID:          userID,
		Name:            req.Name,
		AvatarURL:       req.AvatarUrl,
		Description:     req.Description,
		PersonalityJSON: req.PersonalityJson,
		BackgroundStory: req.BackgroundStory,
		SpeakingStyle:   req.SpeakingStyle,
		SystemPrompt:    req.SystemPrompt,
		WorldViewJSON:   req.WorldViewJson,
		IsPublic:        req.IsPublic,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.CreateCharacterResponse{
		Character: characterToProto(character),
	}, nil
}

// UpdateCharacter 更新角色 (仅限创建者)
func (s *CharacterService) UpdateCharacter(ctx context.Context, req *pb.UpdateCharacterRequest) (*pb.UpdateCharacterResponse, error) {
	userID, err := getRequiredUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	character, err := s.characterService.UpdateCharacter(ctx, biz.UpdateCharacterInput{
		CharacterID:     req.Id,
		UserID:          userID,
		Name:            req.Name,
		AvatarURL:       req.AvatarUrl,
		Description:     req.Description,
		PersonalityJSON: req.PersonalityJson,
		BackgroundStory: req.BackgroundStory,
		SpeakingStyle:   req.SpeakingStyle,
		SystemPrompt:    req.SystemPrompt,
		WorldViewJSON:   req.WorldViewJson,
		IsPublic:        req.IsPublic,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.UpdateCharacterResponse{
		Character: characterToProto(character),
	}, nil
}

// DeleteCharacter 删除角色 (仅限创建者, 软删除)
func (s *CharacterService) DeleteCharacter(ctx context.Context, req *pb.DeleteCharacterRequest) (*pb.DeleteCharacterResponse, error) {
	userID, err := getRequiredUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.characterService.DeleteCharacter(ctx, req.Id, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.DeleteCharacterResponse{
		Success: true,
	}, nil
}

// 辅助函数

// getOptionalUserIDFromContext 从上下文获取用户ID (可选，未登录返回0)
func getOptionalUserIDFromContext(ctx context.Context) int64 {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return 0
	}
	return userID
}

// getRequiredUserIDFromContext 从上下文获取用户ID (必需，未登录返回错误)
func getRequiredUserIDFromContext(ctx context.Context) (int64, error) {
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
	case errors.Is(err, biz.ErrCharacterNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrCharacterAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, biz.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, biz.ErrCannotDeletePreset),
		errors.Is(err, biz.ErrCannotUpdatePreset):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, biz.ErrInvalidName),
		errors.Is(err, biz.ErrInvalidPersonality),
		errors.Is(err, biz.ErrInvalidWorldView),
		errors.Is(err, biz.ErrEmptySystemPrompt):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}

// characterToProto 将角色实体转换为 protobuf 消息
func characterToProto(char *biz.Character) *pb.Character {
	if char == nil {
		return nil
	}

	return &pb.Character{
		Id:              char.ID,
		UserId:          char.UserID,
		Name:            char.Name,
		AvatarUrl:       char.AvatarURL,
		Description:     char.Description,
		PersonalityJson: char.PersonalityString(),
		BackgroundStory: char.BackgroundStory,
		SpeakingStyle:   char.SpeakingStyle,
		SystemPrompt:    char.SystemPrompt,
		WorldViewJson:   char.WorldViewString(),
		IsPublic:        char.IsPublic,
		IsPreset:        char.IsPreset,
		CreatedAt:       timestamppb.New(char.CreatedAt),
		UpdatedAt:       timestamppb.New(char.UpdatedAt),
	}
}
