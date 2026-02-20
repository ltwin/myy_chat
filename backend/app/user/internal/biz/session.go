// Package biz 用户业务逻辑层
package biz

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Session 用户会话实体
type Session struct {
	ID               int64      // 会话ID (雪花ID)
	UserID           int64      // 用户ID
	RefreshTokenHash string     // 刷新令牌哈希
	UserAgent        string     // 用户代理
	IPAddress        string     // IP地址
	ExpiresAt        time.Time  // 过期时间
	CreatedAt        time.Time  // 创建时间
	RevokedAt        *time.Time // 撤销时间
}

// SessionConfig 会话配置
type SessionConfig struct {
	RefreshTokenTTL time.Duration // 刷新令牌有效期
}

// DefaultSessionConfig 默认会话配置
var DefaultSessionConfig = SessionConfig{
	RefreshTokenTTL: 7 * 24 * time.Hour, // 7天
}

// NewSession 创建新会话
func NewSession(id, userID int64, userAgent, ipAddress string, config SessionConfig) (*Session, string, error) {
	// 生成刷新令牌
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, "", err
	}

	// 哈希刷新令牌用于存储
	tokenHash := hashToken(refreshToken)

	now := time.Now()
	session := &Session{
		ID:               id,
		UserID:           userID,
		RefreshTokenHash: tokenHash,
		UserAgent:        userAgent,
		IPAddress:        ipAddress,
		ExpiresAt:        now.Add(config.RefreshTokenTTL),
		CreatedAt:        now,
	}

	return session, refreshToken, nil
}

// generateRefreshToken 生成随机刷新令牌
func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken 哈希令牌
func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// HashRefreshToken 公开的令牌哈希函数
func HashRefreshToken(token string) string {
	return hashToken(token)
}

// IsExpired 检查会话是否过期
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsRevoked 检查会话是否已撤销
func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

// Revoke 撤销会话
func (s *Session) Revoke() {
	now := time.Now()
	s.RevokedAt = &now
}

// IsValid 检查会话是否有效
func (s *Session) IsValid() bool {
	return !s.IsExpired() && !s.IsRevoked()
}

// VerifyRefreshToken 验证刷新令牌
func (s *Session) VerifyRefreshToken(token string) bool {
	return s.RefreshTokenHash == hashToken(token)
}

// RotateRefreshToken 轮换刷新令牌并重置过期时间
func (s *Session) RotateRefreshToken(ttl time.Duration) (string, error) {
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return "", err
	}
	s.RefreshTokenHash = hashToken(refreshToken)
	s.ExpiresAt = time.Now().Add(ttl)
	return refreshToken, nil
}

// Extend 延长会话有效期
func (s *Session) Extend(duration time.Duration) {
	s.ExpiresAt = time.Now().Add(duration)
}

// SessionRepo 会话仓储接口
type SessionRepo interface {
	// Create 创建会话
	Create(ctx context.Context, session *Session) error
	// GetByID 根据ID获取会话
	GetByID(ctx context.Context, id int64) (*Session, error)
	// GetByRefreshTokenHash 根据刷新令牌哈希获取会话
	GetByRefreshTokenHash(ctx context.Context, hash string) (*Session, error)
	// Update 更新会话
	Update(ctx context.Context, session *Session) error
	// Revoke 撤销会话
	Revoke(ctx context.Context, id int64) error
	// RevokeAllByUserID 撤销用户所有会话
	RevokeAllByUserID(ctx context.Context, userID int64) error
	// DeleteExpired 删除过期会话
	DeleteExpired(ctx context.Context) (int64, error)
	// ListByUserID 列出用户所有活跃会话
	ListByUserID(ctx context.Context, userID int64) ([]*Session, error)
}
