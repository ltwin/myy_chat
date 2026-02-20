// Package middleware 提供 HTTP/gRPC 中间件
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/redis/go-redis/v9"
)

// LoginLockoutConfig 登录锁定配置
type LoginLockoutConfig struct {
	MaxAttempts   int           // 最大尝试次数，默认5次
	LockDuration  time.Duration // 锁定时长，默认15分钟
	WindowSize    time.Duration // 计数窗口，默认15分钟
	RedisClient   *redis.Client // Redis 客户端
	KeyPrefix     string        // Redis key 前缀
}

// DefaultLoginLockoutConfig 默认配置
var DefaultLoginLockoutConfig = LoginLockoutConfig{
	MaxAttempts:  5,
	LockDuration: 15 * time.Minute,
	WindowSize:   15 * time.Minute,
	KeyPrefix:    "login_lockout:",
}

// LoginLockout 登录锁定中间件
// 在5次登录失败后锁定账户15分钟 (T234)
type LoginLockout struct {
	config LoginLockoutConfig
	log    *log.Helper
}

// NewLoginLockout 创建登录锁定中间件
func NewLoginLockout(config LoginLockoutConfig, logger log.Logger) *LoginLockout {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = DefaultLoginLockoutConfig.MaxAttempts
	}
	if config.LockDuration <= 0 {
		config.LockDuration = DefaultLoginLockoutConfig.LockDuration
	}
	if config.WindowSize <= 0 {
		config.WindowSize = DefaultLoginLockoutConfig.WindowSize
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = DefaultLoginLockoutConfig.KeyPrefix
	}

	return &LoginLockout{
		config: config,
		log:    log.NewHelper(logger),
	}
}

// attemptKey 生成尝试次数的 Redis key
func (l *LoginLockout) attemptKey(identifier string) string {
	return fmt.Sprintf("%sattempts:%s", l.config.KeyPrefix, identifier)
}

// lockKey 生成锁定状态的 Redis key
func (l *LoginLockout) lockKey(identifier string) string {
	return fmt.Sprintf("%slocked:%s", l.config.KeyPrefix, identifier)
}

// IsLocked 检查是否被锁定
func (l *LoginLockout) IsLocked(ctx context.Context, identifier string) (bool, time.Duration, error) {
	if l.config.RedisClient == nil {
		return false, 0, nil
	}

	ttl, err := l.config.RedisClient.TTL(ctx, l.lockKey(identifier)).Result()
	if err != nil {
		if err == redis.Nil {
			return false, 0, nil
		}
		return false, 0, err
	}

	if ttl > 0 {
		return true, ttl, nil
	}

	return false, 0, nil
}

// RecordFailedAttempt 记录登录失败
func (l *LoginLockout) RecordFailedAttempt(ctx context.Context, identifier string) (int, error) {
	if l.config.RedisClient == nil {
		return 0, nil
	}

	key := l.attemptKey(identifier)

	// 使用 Redis 事务增加计数
	pipe := l.config.RedisClient.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, l.config.WindowSize)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	attempts := int(incr.Val())

	// 如果超过最大尝试次数，设置锁定
	if attempts >= l.config.MaxAttempts {
		err = l.config.RedisClient.Set(ctx, l.lockKey(identifier), "1", l.config.LockDuration).Err()
		if err != nil {
			l.log.Warnf("failed to set lock: %v", err)
		}
		// 清除尝试计数
		l.config.RedisClient.Del(ctx, key)
		l.log.Infof("account locked due to too many failed attempts: %s", identifier)
	}

	return attempts, nil
}

// RecordSuccessfulLogin 记录登录成功，清除失败计数
func (l *LoginLockout) RecordSuccessfulLogin(ctx context.Context, identifier string) error {
	if l.config.RedisClient == nil {
		return nil
	}

	pipe := l.config.RedisClient.Pipeline()
	pipe.Del(ctx, l.attemptKey(identifier))
	pipe.Del(ctx, l.lockKey(identifier))
	_, err := pipe.Exec(ctx)
	return err
}

// GetRemainingAttempts 获取剩余尝试次数
func (l *LoginLockout) GetRemainingAttempts(ctx context.Context, identifier string) (int, error) {
	if l.config.RedisClient == nil {
		return l.config.MaxAttempts, nil
	}

	attempts, err := l.config.RedisClient.Get(ctx, l.attemptKey(identifier)).Int()
	if err != nil {
		if err == redis.Nil {
			return l.config.MaxAttempts, nil
		}
		return 0, err
	}

	remaining := l.config.MaxAttempts - attempts
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// Unlock 手动解锁 (管理员操作)
func (l *LoginLockout) Unlock(ctx context.Context, identifier string) error {
	if l.config.RedisClient == nil {
		return nil
	}

	pipe := l.config.RedisClient.Pipeline()
	pipe.Del(ctx, l.attemptKey(identifier))
	pipe.Del(ctx, l.lockKey(identifier))
	_, err := pipe.Exec(ctx)
	return err
}

// Middleware 返回 Kratos 中间件
// 用于在登录请求前检查锁定状态
func (l *LoginLockout) Middleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 获取客户端标识符 (可以是 IP 或邮箱)
			identifier := l.getIdentifier(ctx, req)
			if identifier == "" {
				return handler(ctx, req)
			}

			// 检查是否被锁定
			locked, ttl, err := l.IsLocked(ctx, identifier)
			if err != nil {
				l.log.Warnf("failed to check lock status: %v", err)
				// 出错时不阻止登录
				return handler(ctx, req)
			}

			if locked {
				return nil, errors.New(429, "ACCOUNT_LOCKED",
					fmt.Sprintf("Account is locked. Please try again in %d minutes.", int(ttl.Minutes())+1))
			}

			return handler(ctx, req)
		}
	}
}

// getIdentifier 从请求中获取标识符
func (l *LoginLockout) getIdentifier(ctx context.Context, req interface{}) string {
	// 尝试从请求中获取邮箱
	if loginReq, ok := req.(interface{ GetEmail() string }); ok {
		return loginReq.GetEmail()
	}

	// 回退到 IP 地址
	if tr, ok := transport.FromServerContext(ctx); ok {
		if httpTr, ok := tr.(*http.Transport); ok {
			return httpTr.Request().RemoteAddr
		}
	}

	return ""
}

// ErrAccountLocked 账户锁定错误
var ErrAccountLocked = errors.New(429, "ACCOUNT_LOCKED", "Account is locked due to too many failed login attempts")
