package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testSigningKey = []byte("test-secret-key-for-unit-tests")

func TestNewJWTGenerator(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)
	assert.NotNil(t, generator)
}

func TestGenerateToken(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	tests := []struct {
		name     string
		userID   int64
		username string
	}{
		{"normal user", 12345, "testuser"},
		{"user with special chars", 99999, "user@example.com"},
		{"user with chinese", 1, "用户名"},
		{"snowflake id", 1735344000000001, "snowflake_user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := generator.GenerateToken(tt.userID, tt.username)
			require.NoError(t, err)
			assert.NotEmpty(t, token)

			// 验证 token 可以被解析
			claims, err := generator.ParseToken(token)
			require.NoError(t, err)
			assert.Equal(t, tt.userID, claims.UserID)
			assert.Equal(t, tt.username, claims.Username)
			assert.Equal(t, "access_token", claims.Subject)
			assert.Equal(t, "myy-chat", claims.Issuer)
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	userID := int64(12345)
	username := "testuser"

	token, err := generator.GenerateRefreshToken(userID, username)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// 验证 token 可以被解析
	claims, err := generator.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "refresh_token", claims.Subject)
}

func TestGenerateTokenPair(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	userID := int64(12345)
	username := "testuser"

	accessToken, refreshToken, err := generator.GenerateTokenPair(userID, username)
	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)

	// 验证 access token
	accessClaims, err := generator.ParseToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, "access_token", accessClaims.Subject)

	// 验证 refresh token
	refreshClaims, err := generator.ParseToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, "refresh_token", refreshClaims.Subject)
}

func TestParseToken_Valid(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	token, err := generator.GenerateToken(12345, "testuser")
	require.NoError(t, err)

	claims, err := generator.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(12345), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestParseToken_Invalid(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "not.a.valid.jwt"},
		{"random string", "abcdefghijklmnop"},
		{"malformed jwt", "eyJhbGciOiJIUzI1NiJ9.invalid.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := generator.ParseToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestParseToken_WrongSigningKey(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	// 使用不同的密钥生成 token
	wrongConfig := DefaultJWTConfig([]byte("different-secret-key"))
	wrongGenerator := NewJWTGenerator(wrongConfig)

	token, err := wrongGenerator.GenerateToken(12345, "testuser")
	require.NoError(t, err)

	// 使用原始密钥解析应该失败
	_, err = generator.ParseToken(token)
	assert.Error(t, err)
}

func TestParseToken_Expired(t *testing.T) {
	// 创建一个短过期时间的配置
	config := DefaultJWTConfig(testSigningKey)
	config.TokenExpiration = -1 * time.Second // 已过期

	generator := NewJWTGenerator(config)

	token, err := generator.GenerateToken(12345, "testuser")
	require.NoError(t, err)

	// 解析过期 token 应该返回错误
	_, err = generator.ParseToken(token)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func TestRefreshToken_Valid(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	// 生成刷新令牌
	_, refreshToken, err := generator.GenerateTokenPair(12345, "testuser")
	require.NoError(t, err)

	// 刷新令牌
	newAccessToken, newRefreshToken, err := generator.RefreshToken(refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	// 新令牌应该不同于空字符串，但可能与原令牌相同（取决于时间精度）
	// 改为验证令牌可以正确解析
	claims, err := generator.ParseToken(newAccessToken)
	require.NoError(t, err)
	assert.Equal(t, int64(12345), claims.UserID)
}

func TestRefreshToken_WithAccessToken(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	// 使用访问令牌尝试刷新应该失败
	accessToken, _, err := generator.GenerateTokenPair(12345, "testuser")
	require.NoError(t, err)

	_, _, err = generator.RefreshToken(accessToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}

func TestContextFunctions(t *testing.T) {
	ctx := context.Background()

	// 测试 UserID
	ctx = WithUserID(ctx, 12345)
	userID, ok := UserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, int64(12345), userID)

	// 测试 Username
	ctx = WithUsername(ctx, "testuser")
	username, ok := UsernameFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "testuser", username)

	// 测试 Claims
	claims := &UserClaims{UserID: 12345, Username: "testuser"}
	ctx = WithClaims(ctx, claims)
	retrievedClaims, ok := ClaimsFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, claims, retrievedClaims)
}

func TestContextFunctions_NotSet(t *testing.T) {
	ctx := context.Background()

	// 未设置时应返回 false
	_, ok := UserIDFromContext(ctx)
	assert.False(t, ok)

	_, ok = UsernameFromContext(ctx)
	assert.False(t, ok)

	_, ok = ClaimsFromContext(ctx)
	assert.False(t, ok)
}

func TestHTTPStatusFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"missing token", ErrMissingToken, 401},
		{"invalid token", ErrInvalidToken, 401},
		{"expired token", ErrExpiredToken, 401},
		{"invalid signature", ErrInvalidSignature, 401},
		{"unknown error", assert.AnError, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := HTTPStatusFromError(tt.err)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestDefaultJWTConfig(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)

	assert.Equal(t, testSigningKey, config.SigningKey)
	assert.Equal(t, jwt.SigningMethodHS256, config.SigningMethod)
	assert.Equal(t, "header:Authorization", config.TokenLookup)
	assert.Equal(t, "Bearer", config.AuthScheme)
	assert.Equal(t, 24*time.Hour, config.TokenExpiration)
	assert.Equal(t, 7*24*time.Hour, config.RefreshExpiration)
	assert.NotNil(t, config.Claims)
}

func TestUserClaims(t *testing.T) {
	claims := &UserClaims{
		UserID:   12345,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "test",
		},
	}

	// 验证 claims 实现了 jwt.Claims 接口
	var _ jwt.Claims = claims
}

func TestSkipPaths(t *testing.T) {
	// 由于 SkipPaths 依赖 transport 上下文，这里仅测试基本功能
	skipFn := SkipPaths("/api/v1/users/login", "/api/v1/users/register")
	assert.NotNil(t, skipFn)

	// 没有 transport 上下文时应返回 false
	ctx := context.Background()
	assert.False(t, skipFn(ctx))
}

func TestTokenExpiration(t *testing.T) {
	config := DefaultJWTConfig(testSigningKey)
	config.TokenExpiration = 1 * time.Second // 使用1秒，更可靠
	generator := NewJWTGenerator(config)

	token, err := generator.GenerateToken(12345, "testuser")
	require.NoError(t, err)

	// 立即解析应该成功
	claims, err := generator.ParseToken(token)
	require.NoError(t, err)
	assert.Equal(t, int64(12345), claims.UserID)

	// 等待过期
	time.Sleep(1100 * time.Millisecond)

	// 过期后解析应该失败
	_, err = generator.ParseToken(token)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

func BenchmarkGenerateToken(b *testing.B) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.GenerateToken(12345, "testuser")
	}
}

func BenchmarkParseToken(b *testing.B) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)
	token, _ := generator.GenerateToken(12345, "testuser")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = generator.ParseToken(token)
	}
}

func BenchmarkGenerateTokenPair(b *testing.B) {
	config := DefaultJWTConfig(testSigningKey)
	generator := NewJWTGenerator(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = generator.GenerateTokenPair(12345, "testuser")
	}
}
