// Package middleware 提供 HTTP/gRPC 中间件实现
package middleware

import (
	"context"
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
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidSignature = errors.New("invalid token signature")
)

// UserClaims JWT 用户声明
type UserClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
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
}

// DefaultJWTConfig 默认 JWT 配置
func DefaultJWTConfig(signingKey []byte) JWTConfig {
	return JWTConfig{
		SigningKey:        signingKey,
		SigningMethod:     jwt.SigningMethodHS256,
		TokenLookup:       "header:Authorization",
		AuthScheme:        "Bearer",
		Claims:            func() jwt.Claims { return &UserClaims{} },
		TokenExpiration:   time.Hour * 24,       // 24小时
		RefreshExpiration: time.Hour * 24 * 7,   // 7天
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
			parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
				// 验证签名算法
				if t.Method.Alg() != config.SigningMethod.Alg() {
					return nil, ErrInvalidSignature
				}
				return config.SigningKey, nil
			})

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

			// 将 claims 存入上下文
			if config.SuccessHandler != nil {
				ctx = config.SuccessHandler(ctx, claims)
			} else {
				ctx = WithClaims(ctx, claims)
				if userClaims, ok := claims.(*UserClaims); ok {
					ctx = WithUserID(ctx, userClaims.UserID)
					ctx = WithUsername(ctx, userClaims.Username)
				}
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

// JWTGenerator JWT 生成器
type JWTGenerator struct {
	config JWTConfig
}

// NewJWTGenerator 创建 JWT 生成器
func NewJWTGenerator(config JWTConfig) *JWTGenerator {
	return &JWTGenerator{config: config}
}

// GenerateToken 生成访问令牌
func (g *JWTGenerator) GenerateToken(userID int64, username string) (string, error) {
	now := time.Now()
	claims := &UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(g.config.TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "myy-chat",
			Subject:   "access_token",
		},
	}

	token := jwt.NewWithClaims(g.config.SigningMethod, claims)
	return token.SignedString(g.config.SigningKey)
}

// GenerateRefreshToken 生成刷新令牌
func (g *JWTGenerator) GenerateRefreshToken(userID int64, username string) (string, error) {
	now := time.Now()
	claims := &UserClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(g.config.RefreshExpiration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "myy-chat",
			Subject:   "refresh_token",
		},
	}

	token := jwt.NewWithClaims(g.config.SigningMethod, claims)
	return token.SignedString(g.config.SigningKey)
}

// GenerateTokenPair 生成访问令牌和刷新令牌对
func (g *JWTGenerator) GenerateTokenPair(userID int64, username string) (accessToken, refreshToken string, err error) {
	accessToken, err = g.GenerateToken(userID, username)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = g.GenerateRefreshToken(userID, username)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ParseToken 解析令牌
func (g *JWTGenerator) ParseToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != g.config.SigningMethod.Alg() {
			return nil, ErrInvalidSignature
		}
		return g.config.SigningKey, nil
	})

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
		return nil, ErrInvalidToken
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
	if claims.Subject != "refresh_token" {
		return "", "", ErrInvalidToken
	}

	return g.GenerateTokenPair(claims.UserID, claims.Username)
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
	case errors.Is(err, ErrExpiredToken):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidSignature):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
