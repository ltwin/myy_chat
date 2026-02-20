// Package middleware 提供 HTTP/gRPC 中间件实现
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt/v5"
)

// 错误定义
var (
	ErrMissingToken     = errors.New("missing authorization token")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidClaims    = errors.New("invalid token claims")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
	ErrTokenRevoked     = errors.New("token has been revoked")
	ErrBlacklistCheck   = errors.New("failed to validate token revocation status")
)

// UserClaims JWT 用户声明
type UserClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	// SessionID 对应 sessions.id，作为会话绑定锚点
	SessionID int64 `json:"sid"`
	// TokenType 取值 access/refresh
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// TokenBlacklistChecker 用于检查 token/session 是否已撤销
type TokenBlacklistChecker interface {
	IsRevoked(ctx context.Context, sid int64, jti string) (bool, error)
}

// contextKey 上下文键类型
type contextKey string

const (
	// 上下文键常量
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
	claimsKey   contextKey = "claims"
)

// JWTConfig JWT 中间件配置
type JWTConfig struct {
	// SigningKey 签名密钥
	SigningKey []byte
	// SigningMethod 签名算法，默认 HS256
	SigningMethod jwt.SigningMethod
	// TokenLookup token 获取方式，格式: "header:Authorization", "query:token", "cookie:jwt"
	TokenLookup string
	// AuthScheme 认证方案，默认 "Bearer"
	AuthScheme string
	// Claims 自定义 Claims 工厂函数
	Claims func() jwt.Claims
	// Skipper 跳过中间件的函数
	Skipper func(ctx context.Context) bool
	// ErrorHandler 错误处理函数
	ErrorHandler func(ctx context.Context, err error) error
	// SuccessHandler 成功处理函数
	SuccessHandler func(ctx context.Context, claims jwt.Claims) context.Context
	// TokenExpiration Token 过期时间
	TokenExpiration time.Duration
	// RefreshExpiration 刷新令牌过期时间
	RefreshExpiration time.Duration
	// Issuer 签发方
	Issuer string
	// Audience 受众
	Audience string
	// Leeway 时钟偏移容忍
	Leeway time.Duration
	// RequireSIDAccess 校验 access token 时是否强制要求 sid>0
	RequireSIDAccess bool
	// BlacklistChecker token 黑名单检查器（可选）
	BlacklistChecker TokenBlacklistChecker
	// FailOpenOnBlacklistError 黑名单检查异常时是否放行
	FailOpenOnBlacklistError bool
}

// DefaultJWTConfig 默认 JWT 配置
func DefaultJWTConfig(signingKey []byte) JWTConfig {
	return JWTConfig{
		SigningKey:               signingKey,
		SigningMethod:            jwt.SigningMethodHS256,
		TokenLookup:              "header:Authorization",
		AuthScheme:               "Bearer",
		Claims:                   func() jwt.Claims { return &UserClaims{} },
		TokenExpiration:          15 * time.Minute,   // 15分钟
		RefreshExpiration:        7 * 24 * time.Hour, // 7天
		Issuer:                   "myy-chat",
		Audience:                 "myy-chat-api",
		Leeway:                   30 * time.Second,
		RequireSIDAccess:         true,
		FailOpenOnBlacklistError: false,
	}
}

// JWTAuth 创建 JWT 认证中间件
func JWTAuth(config JWTConfig) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 检查是否跳过
			if config.Skipper != nil && config.Skipper(ctx) {
				return handler(ctx, req)
			}

			// 从上下文获取 transport
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				if config.ErrorHandler != nil {
					return nil, config.ErrorHandler(ctx, ErrMissingToken)
				}
				return nil, ErrMissingToken
			}

			// 获取 token
			token := extractToken(tr.RequestHeader(), config)
			if token == "" {
				if config.ErrorHandler != nil {
					return nil, config.ErrorHandler(ctx, ErrMissingToken)
				}
				return nil, ErrMissingToken
			}

			// 解析和验证 token
			claims := config.Claims()
			parsedToken, err := jwt.ParseWithClaims(
				token,
				claims,
				func(t *jwt.Token) (interface{}, error) {
					// 验证签名算法
					if t.Method.Alg() != config.SigningMethod.Alg() {
						return nil, ErrInvalidSignature
					}
					return config.SigningKey, nil
				},
				parserOptions(config)...,
			)

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					if config.ErrorHandler != nil {
						return nil, config.ErrorHandler(ctx, ErrExpiredToken)
					}
					return nil, ErrExpiredToken
				}
				if config.ErrorHandler != nil {
					return nil, config.ErrorHandler(ctx, ErrInvalidToken)
				}
				return nil, ErrInvalidToken
			}

			if !parsedToken.Valid {
				if config.ErrorHandler != nil {
					return nil, config.ErrorHandler(ctx, ErrInvalidToken)
				}
				return nil, ErrInvalidToken
			}

			userClaims, ok := claims.(*UserClaims)
			if !ok {
				if config.ErrorHandler != nil {
					return nil, config.ErrorHandler(ctx, ErrInvalidClaims)
				}
				return nil, ErrInvalidClaims
			}

			if err := validateUserClaims(userClaims, config); err != nil {
				if config.ErrorHandler != nil {
					return nil, config.ErrorHandler(ctx, err)
				}
				return nil, err
			}

			if config.BlacklistChecker != nil {
				revoked, err := config.BlacklistChecker.IsRevoked(ctx, userClaims.SessionID, userClaims.ID)
				if err != nil {
					if !config.FailOpenOnBlacklistError {
						if config.ErrorHandler != nil {
							return nil, config.ErrorHandler(ctx, ErrBlacklistCheck)
						}
						return nil, ErrBlacklistCheck
					}
				}
				if revoked {
					if config.ErrorHandler != nil {
						return nil, config.ErrorHandler(ctx, ErrTokenRevoked)
					}
					return nil, ErrTokenRevoked
				}
			}

			// 将 claims 存入上下文
			if config.SuccessHandler != nil {
				ctx = config.SuccessHandler(ctx, claims)
			} else {
				ctx = WithClaims(ctx, claims)
				ctx = WithUserID(ctx, userClaims.UserID)
				ctx = WithUsername(ctx, userClaims.Username)
			}

			return handler(ctx, req)
		}
	}
}

// extractToken 从请求头中提取 token
func extractToken(header transport.Header, config JWTConfig) string {
	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		return ""
	}

	switch parts[0] {
	case "header":
		auth := header.Get(parts[1])
		if auth == "" {
			return ""
		}
		// 移除 Bearer 前缀
		if config.AuthScheme != "" {
			prefix := config.AuthScheme + " "
			if strings.HasPrefix(auth, prefix) {
				return strings.TrimPrefix(auth, prefix)
			}
		}
		return auth
	case "query":
		return header.Get(parts[1])
	case "cookie":
		// Cookie 需要在 HTTP 上下文中处理
		return header.Get("Cookie")
	}
	return ""
}

func parserOptions(config JWTConfig) []jwt.ParserOption {
	opts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{config.SigningMethod.Alg()}),
	}
	if config.Issuer != "" {
		opts = append(opts, jwt.WithIssuer(config.Issuer))
	}
	if config.Audience != "" {
		opts = append(opts, jwt.WithAudience(config.Audience))
	}
	if config.Leeway > 0 {
		opts = append(opts, jwt.WithLeeway(config.Leeway))
	}
	return opts
}

func validateUserClaims(claims *UserClaims, config JWTConfig) error {
	if claims == nil || claims.UserID <= 0 {
		return ErrInvalidClaims
	}

	// 兼容历史 token：TokenType 缺失时回退到 Subject 推断。
	if claims.TokenType == "" {
		switch claims.Subject {
		case "access_token":
			claims.TokenType = "access"
		case "refresh_token":
			claims.TokenType = "refresh"
		default:
			return ErrInvalidClaims
		}
	}

	switch claims.TokenType {
	case "access":
		if claims.Subject != "access_token" {
			return ErrInvalidClaims
		}
		if config.RequireSIDAccess && claims.SessionID <= 0 {
			return ErrInvalidClaims
		}
	case "refresh":
		if claims.Subject != "refresh_token" {
			return ErrInvalidClaims
		}
	default:
		return ErrInvalidClaims
	}

	if claims.ID == "" {
		return ErrInvalidClaims
	}
	return nil
}

// JWTGenerator JWT 生成器
type JWTGenerator struct {
	config JWTConfig
}

func generateTokenID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// NewJWTGenerator 创建 JWT 生成器
func NewJWTGenerator(config JWTConfig) *JWTGenerator {
	return &JWTGenerator{config: config}
}

// GenerateToken 生成访问令牌
func (g *JWTGenerator) GenerateToken(userID int64, username string) (string, error) {
	// 兼容旧调用方：未显式传 sid 时回退使用 userID。
	return g.GenerateTokenWithSession(userID, username, userID)
}

// GenerateTokenWithSession 生成携带 sid/jti 的访问令牌
func (g *JWTGenerator) GenerateTokenWithSession(userID int64, username string, sid int64) (string, error) {
	if sid <= 0 {
		return "", ErrInvalidClaims
	}
	jti, err := generateTokenID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	registered := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(g.config.TokenExpiration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    g.config.Issuer,
		Subject:   "access_token",
		ID:        jti,
	}
	if g.config.Audience != "" {
		registered.Audience = jwt.ClaimStrings{g.config.Audience}
	}
	claims := &UserClaims{
		UserID:           userID,
		Username:         username,
		SessionID:        sid,
		TokenType:        "access",
		RegisteredClaims: registered,
	}

	token := jwt.NewWithClaims(g.config.SigningMethod, claims)
	return token.SignedString(g.config.SigningKey)
}

// GenerateRefreshToken 生成刷新令牌
func (g *JWTGenerator) GenerateRefreshToken(userID int64, username string) (string, error) {
	// 兼容旧调用方：未显式传 sid 时回退使用 userID。
	return g.GenerateRefreshTokenWithSession(userID, username, userID)
}

// GenerateRefreshTokenWithSession 生成携带 sid/jti 的刷新令牌
func (g *JWTGenerator) GenerateRefreshTokenWithSession(userID int64, username string, sid int64) (string, error) {
	if sid <= 0 {
		return "", ErrInvalidClaims
	}
	jti, err := generateTokenID()
	if err != nil {
		return "", err
	}

	now := time.Now()
	registered := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(g.config.RefreshExpiration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    g.config.Issuer,
		Subject:   "refresh_token",
		ID:        jti,
	}
	if g.config.Audience != "" {
		registered.Audience = jwt.ClaimStrings{g.config.Audience}
	}
	claims := &UserClaims{
		UserID:           userID,
		Username:         username,
		SessionID:        sid,
		TokenType:        "refresh",
		RegisteredClaims: registered,
	}

	token := jwt.NewWithClaims(g.config.SigningMethod, claims)
	return token.SignedString(g.config.SigningKey)
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (g *JWTGenerator) GenerateTokenPair(userID int64, username string) (accessToken, refreshToken string, err error) {
	return g.GenerateTokenPairWithSession(userID, username, userID)
}

// GenerateTokenPairWithSession 生成同会话绑定的令牌对
func (g *JWTGenerator) GenerateTokenPairWithSession(userID int64, username string, sid int64) (accessToken, refreshToken string, err error) {
	accessToken, err = g.GenerateTokenWithSession(userID, username, sid)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = g.GenerateRefreshTokenWithSession(userID, username, sid)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken 解析令牌
func (g *JWTGenerator) ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&UserClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != g.config.SigningMethod.Alg() {
				return nil, ErrInvalidSignature
			}
			return g.config.SigningKey, nil
		},
		parserOptions(g.config)...,
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	if err := validateUserClaims(claims, g.config); err != nil {
		return nil, err
	}

	return claims, nil
}

// RefreshToken 刷新令牌
func (g *JWTGenerator) RefreshToken(refreshTokenString string) (newAccessToken, newRefreshToken string, err error) {
	claims, err := g.ParseToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	// 验证是刷新令牌
	if claims.Subject != "refresh_token" || claims.TokenType != "refresh" {
		return "", "", ErrInvalidToken
	}

	return g.GenerateTokenPairWithSession(claims.UserID, claims.Username, claims.SessionID)
}

// Context 辅助函数

// WithUserID 将用户 ID 存入上下文
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext 从上下文获取用户 ID
func UserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(userIDKey).(int64)
	return userID, ok
}

// WithUsername 将用户名存入上下文
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameKey, username)
}

// UsernameFromContext 从上下文获取用户名
func UsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(usernameKey).(string)
	return username, ok
}

// WithClaims 将 Claims 存入上下文
func WithClaims(ctx context.Context, claims jwt.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, claims)
}

// ClaimsFromContext 从上下文获取 Claims
func ClaimsFromContext(ctx context.Context) (jwt.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(jwt.Claims)
	return claims, ok
}

// SkipPaths 创建路径跳过函数
func SkipPaths(paths ...string) func(ctx context.Context) bool {
	pathMap := make(map[string]bool, len(paths))
	for _, p := range paths {
		pathMap[p] = true
	}

	return func(ctx context.Context) bool {
		tr, ok := transport.FromServerContext(ctx)
		if !ok {
			return false
		}

		// 从 HTTP header 获取路径
		path := tr.Operation()
		return pathMap[path]
	}
}

// HTTPStatusFromError 将 JWT 错误转换为 HTTP 状态码
func HTTPStatusFromError(err error) int {
	switch {
	case errors.Is(err, ErrMissingToken):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidToken):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidClaims):
		return http.StatusUnauthorized
	case errors.Is(err, ErrExpiredToken):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidSignature):
		return http.StatusUnauthorized
	case errors.Is(err, ErrTokenRevoked):
		return http.StatusUnauthorized
	case errors.Is(err, ErrBlacklistCheck):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
