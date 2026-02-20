// Package biz 角色业务逻辑层
package biz

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"unicode/utf8"
)

// 角色相关错误定义
var (
	ErrCharacterNotFound      = errors.New("character not found")
	ErrCharacterAlreadyExists = errors.New("character already exists")
	ErrInvalidName            = errors.New("invalid character name: must be 2-100 characters")
	ErrInvalidPersonality     = errors.New("invalid personality JSON format")
	ErrInvalidWorldView       = errors.New("invalid world view JSON format")
	ErrEmptySystemPrompt      = errors.New("system prompt cannot be empty")
	ErrPermissionDenied       = errors.New("permission denied: not the character owner")
	ErrCannotDeletePreset     = errors.New("cannot delete preset character")
	ErrCannotUpdatePreset     = errors.New("cannot update preset character")
)

// Character 角色实体
type Character struct {
	ID              int64           // 雪花ID
	UserID          int64           // 创建者ID (0表示预设角色)
	Name            string          // 角色名称 (2-100字符)
	AvatarURL       string          // 头像URL
	Description     string          // 角色描述
	Personality     json.RawMessage // 性格设定 JSONB
	BackgroundStory string          // 背景故事
	SpeakingStyle   string          // 说话风格
	SystemPrompt    string          // 系统提示词
	WorldView       json.RawMessage // 世界观设定 JSONB
	IsPublic        bool            // 是否公开
	IsPreset        bool            // 是否预设角色
	Version         int             // 版本号 (乐观锁)
	CreatedAt       time.Time       // 创建时间
	UpdatedAt       time.Time       // 更新时间
	IsDeleted       bool            // 软删除标记
}

// NewCharacter 创建新角色
func NewCharacter(id int64, userID int64, name, description, personalityJSON, systemPrompt string) (*Character, error) {
	// 验证名称
	if err := ValidateName(name); err != nil {
		return nil, err
	}

	// 验证系统提示词
	if systemPrompt == "" {
		return nil, ErrEmptySystemPrompt
	}

	// 验证并解析 personality JSON
	var personality json.RawMessage
	if personalityJSON != "" {
		if !json.Valid([]byte(personalityJSON)) {
			return nil, ErrInvalidPersonality
		}
		personality = json.RawMessage(personalityJSON)
	} else {
		personality = json.RawMessage("{}")
	}

	now := time.Now()
	return &Character{
		ID:           id,
		UserID:       userID,
		Name:         name,
		Description:  description,
		Personality:  personality,
		SystemPrompt: systemPrompt,
		WorldView:    json.RawMessage("{}"),
		IsPublic:     false,
		IsPreset:     false,
		Version:      1,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsDeleted:    false,
	}, nil
}

// NewPresetCharacter 创建预设角色
func NewPresetCharacter(id int64, name, avatarURL, description, personalityJSON, backgroundStory, speakingStyle, systemPrompt, worldViewJSON string) (*Character, error) {
	char, err := NewCharacter(id, 0, name, description, personalityJSON, systemPrompt)
	if err != nil {
		return nil, err
	}

	char.AvatarURL = avatarURL
	char.BackgroundStory = backgroundStory
	char.SpeakingStyle = speakingStyle
	char.IsPreset = true
	char.IsPublic = true

	// 设置世界观
	if worldViewJSON != "" {
		if !json.Valid([]byte(worldViewJSON)) {
			return nil, ErrInvalidWorldView
		}
		char.WorldView = json.RawMessage(worldViewJSON)
	}

	return char, nil
}

// ValidateName 验证角色名称
func ValidateName(name string) error {
	length := utf8.RuneCountInString(name)
	if length < 2 || length > 100 {
		return ErrInvalidName
	}
	return nil
}

// Update 更新角色信息
func (c *Character) Update(name, avatarURL, description, personalityJSON, backgroundStory, speakingStyle, systemPrompt, worldViewJSON string, isPublic bool) error {
	// 预设角色不允许修改
	if c.IsPreset {
		return ErrCannotUpdatePreset
	}

	// 更新名称
	if name != "" && name != c.Name {
		if err := ValidateName(name); err != nil {
			return err
		}
		c.Name = name
	}

	// 更新头像
	if avatarURL != c.AvatarURL {
		c.AvatarURL = avatarURL
	}

	// 更新描述
	if description != c.Description {
		c.Description = description
	}

	// 更新性格设定
	if personalityJSON != "" {
		if !json.Valid([]byte(personalityJSON)) {
			return ErrInvalidPersonality
		}
		c.Personality = json.RawMessage(personalityJSON)
	}

	// 更新背景故事
	if backgroundStory != c.BackgroundStory {
		c.BackgroundStory = backgroundStory
	}

	// 更新说话风格
	if speakingStyle != c.SpeakingStyle {
		c.SpeakingStyle = speakingStyle
	}

	// 更新系统提示词
	if systemPrompt != "" && systemPrompt != c.SystemPrompt {
		c.SystemPrompt = systemPrompt
	}

	// 更新世界观
	if worldViewJSON != "" {
		if !json.Valid([]byte(worldViewJSON)) {
			return ErrInvalidWorldView
		}
		c.WorldView = json.RawMessage(worldViewJSON)
	}

	// 更新公开状态
	c.IsPublic = isPublic

	c.UpdatedAt = time.Now()
	c.Version++
	return nil
}

// MarkDeleted 标记角色已删除
func (c *Character) MarkDeleted() error {
	// 预设角色不允许删除
	if c.IsPreset {
		return ErrCannotDeletePreset
	}

	c.IsDeleted = true
	c.UpdatedAt = time.Now()
	return nil
}

// CanBeAccessedBy 检查角色是否可以被指定用户访问
// 规则: 预设角色、公开角色、用户自己的角色可访问
func (c *Character) CanBeAccessedBy(userID int64) bool {
	if c.IsDeleted {
		return false
	}
	// 预设角色或公开角色
	if c.IsPreset || c.IsPublic {
		return true
	}
	// 用户自己的角色
	return c.UserID == userID
}

// CanBeModifiedBy 检查角色是否可以被指定用户修改
// 规则: 仅创建者可以修改，预设角色不可修改
func (c *Character) CanBeModifiedBy(userID int64) bool {
	if c.IsDeleted || c.IsPreset {
		return false
	}
	return c.UserID == userID
}

// PersonalityString 返回 Personality 的字符串形式
func (c *Character) PersonalityString() string {
	if len(c.Personality) == 0 {
		return "{}"
	}
	return string(c.Personality)
}

// WorldViewString 返回 WorldView 的字符串形式
func (c *Character) WorldViewString() string {
	if len(c.WorldView) == 0 {
		return "{}"
	}
	return string(c.WorldView)
}

// CharacterRepo 角色仓储接口
type CharacterRepo interface {
	// Create 创建角色
	Create(ctx context.Context, character *Character) error
	// GetByID 根据ID获取角色
	GetByID(ctx context.Context, id int64) (*Character, error)
	// Update 更新角色
	Update(ctx context.Context, character *Character) error
	// Delete 删除角色 (软删除)
	Delete(ctx context.Context, id int64) error
	// List 列出角色 (支持分页和过滤)
	List(ctx context.Context, userID int64, filter string, page, pageSize int) ([]*Character, int, error)
	// ExistsByName 检查角色名称是否存在 (针对特定用户)
	ExistsByName(ctx context.Context, userID int64, name string) (bool, error)
}
