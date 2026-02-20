// Package biz 角色业务逻辑层
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/myy-chat/backend/golang/pkg/snowflake"
)

// CharacterService 角色业务服务
type CharacterService struct {
	characterRepo CharacterRepo
	idGen         snowflake.Generator
	log           *log.Helper
}

// NewCharacterService 创建角色服务
func NewCharacterService(
	characterRepo CharacterRepo,
	idGen snowflake.Generator,
	logger log.Logger,
) *CharacterService {
	return &CharacterService{
		characterRepo: characterRepo,
		idGen:         idGen,
		log:           log.NewHelper(logger),
	}
}

// ListCharactersInput 列出角色输入
type ListCharactersInput struct {
	UserID   int64  // 当前用户ID (0表示未登录)
	Filter   string // 过滤条件: "all", "public", "preset", "mine"
	Page     int    // 页码
	PageSize int    // 每页数量
}

// ListCharactersOutput 列出角色输出
type ListCharactersOutput struct {
	Characters []*Character
	Total      int
	Page       int
	PageSize   int
}

// ListCharacters 列出角色
// 规则:
// - "all": 预设角色 + 公开角色 + 用户自己的角色
// - "public": 公开角色 + 预设角色
// - "preset": 仅预设角色
// - "mine": 仅用户自己的角色 (需要登录)
func (s *CharacterService) ListCharacters(ctx context.Context, input ListCharactersInput) (*ListCharactersOutput, error) {
	// 验证分页参数
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 10
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	// 默认过滤条件
	if input.Filter == "" {
		input.Filter = "all"
	}

	characters, total, err := s.characterRepo.List(ctx, input.UserID, input.Filter, input.Page, input.PageSize)
	if err != nil {
		return nil, err
	}

	return &ListCharactersOutput{
		Characters: characters,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
	}, nil
}

// GetCharacter 获取角色详情
func (s *CharacterService) GetCharacter(ctx context.Context, characterID int64, userID int64) (*Character, error) {
	character, err := s.characterRepo.GetByID(ctx, characterID)
	if err != nil {
		return nil, err
	}

	// 检查访问权限
	if !character.CanBeAccessedBy(userID) {
		return nil, ErrCharacterNotFound // 不暴露权限错误，统一返回 not found
	}

	return character, nil
}

// CreateCharacterInput 创建角色输入
type CreateCharacterInput struct {
	UserID          int64  // 创建者ID
	Name            string // 角色名称
	AvatarURL       string // 头像URL
	Description     string // 角色描述
	PersonalityJSON string // 性格设定 JSON
	BackgroundStory string // 背景故事
	SpeakingStyle   string // 说话风格
	SystemPrompt    string // 系统提示词
	WorldViewJSON   string // 世界观设定 JSON
	IsPublic        bool   // 是否公开
}

// CreateCharacter 创建角色
func (s *CharacterService) CreateCharacter(ctx context.Context, input CreateCharacterInput) (*Character, error) {
	// 检查角色名称是否已存在 (同一用户不能创建同名角色)
	exists, err := s.characterRepo.ExistsByName(ctx, input.UserID, input.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCharacterAlreadyExists
	}

	// 生成雪花ID
	characterID := s.idGen.Generate()

	// 创建角色实体
	character, err := NewCharacter(characterID, input.UserID, input.Name, input.Description, input.PersonalityJSON, input.SystemPrompt)
	if err != nil {
		return nil, err
	}

	// 设置可选字段
	character.AvatarURL = input.AvatarURL
	character.BackgroundStory = input.BackgroundStory
	character.SpeakingStyle = input.SpeakingStyle
	character.IsPublic = input.IsPublic

	// 设置世界观
	if input.WorldViewJSON != "" {
		if err := character.Update("", "", "", "", "", "", "", input.WorldViewJSON, input.IsPublic); err != nil {
			return nil, err
		}
		// 恢复 Version 为 1 (Update 会增加版本号)
		character.Version = 1
	}

	// 保存角色
	if err := s.characterRepo.Create(ctx, character); err != nil {
		return nil, err
	}

	s.log.Infof("character created: id=%d, name=%s, user_id=%d", characterID, input.Name, input.UserID)
	return character, nil
}

// UpdateCharacterInput 更新角色输入
type UpdateCharacterInput struct {
	CharacterID     int64  // 角色ID
	UserID          int64  // 当前用户ID
	Name            string // 新名称
	AvatarURL       string // 新头像URL
	Description     string // 新描述
	PersonalityJSON string // 新性格设定
	BackgroundStory string // 新背景故事
	SpeakingStyle   string // 新说话风格
	SystemPrompt    string // 新系统提示词
	WorldViewJSON   string // 新世界观设定
	IsPublic        bool   // 是否公开
}

// UpdateCharacter 更新角色
func (s *CharacterService) UpdateCharacter(ctx context.Context, input UpdateCharacterInput) (*Character, error) {
	// 获取角色
	character, err := s.characterRepo.GetByID(ctx, input.CharacterID)
	if err != nil {
		return nil, err
	}

	// 检查修改权限
	if !character.CanBeModifiedBy(input.UserID) {
		if character.IsPreset {
			return nil, ErrCannotUpdatePreset
		}
		return nil, ErrPermissionDenied
	}

	// 如果要修改名称，检查新名称是否已存在
	if input.Name != "" && input.Name != character.Name {
		exists, err := s.characterRepo.ExistsByName(ctx, input.UserID, input.Name)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCharacterAlreadyExists
		}
	}

	// 更新角色
	if err := character.Update(
		input.Name,
		input.AvatarURL,
		input.Description,
		input.PersonalityJSON,
		input.BackgroundStory,
		input.SpeakingStyle,
		input.SystemPrompt,
		input.WorldViewJSON,
		input.IsPublic,
	); err != nil {
		return nil, err
	}

	// 保存更新
	if err := s.characterRepo.Update(ctx, character); err != nil {
		return nil, err
	}

	s.log.Infof("character updated: id=%d, name=%s", character.ID, character.Name)
	return character, nil
}

// DeleteCharacter 删除角色 (软删除)
func (s *CharacterService) DeleteCharacter(ctx context.Context, characterID int64, userID int64) error {
	// 获取角色
	character, err := s.characterRepo.GetByID(ctx, characterID)
	if err != nil {
		return err
	}

	// 检查修改权限
	if !character.CanBeModifiedBy(userID) {
		if character.IsPreset {
			return ErrCannotDeletePreset
		}
		return ErrPermissionDenied
	}

	// 标记删除
	if err := character.MarkDeleted(); err != nil {
		return err
	}

	// 保存更新
	if err := s.characterRepo.Update(ctx, character); err != nil {
		return err
	}

	s.log.Infof("character deleted: id=%d, name=%s", characterID, character.Name)
	return nil
}
