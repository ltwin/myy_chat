package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	cleanup := func() {
		client.Close()
		s.Close()
	}

	return client, cleanup
}

func TestLoginLockout_RecordFailedAttempt(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	lockout := NewLoginLockout(LoginLockoutConfig{
		MaxAttempts:  3,
		LockDuration: 5 * time.Minute,
		WindowSize:   5 * time.Minute,
		RedisClient:  client,
		KeyPrefix:    "test:",
	}, log.DefaultLogger)

	ctx := context.Background()
	identifier := "test@example.com"

	// 第一次失败
	attempts, err := lockout.RecordFailedAttempt(ctx, identifier)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}

	// 检查未锁定
	locked, _, err := lockout.IsLocked(ctx, identifier)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if locked {
		t.Error("should not be locked after 1 attempt")
	}

	// 第二次失败
	attempts, _ = lockout.RecordFailedAttempt(ctx, identifier)
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}

	// 第三次失败 - 应该触发锁定
	attempts, _ = lockout.RecordFailedAttempt(ctx, identifier)
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}

	// 检查已锁定
	locked, ttl, err := lockout.IsLocked(ctx, identifier)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !locked {
		t.Error("should be locked after 3 attempts")
	}
	if ttl <= 0 {
		t.Error("TTL should be positive")
	}
}

func TestLoginLockout_RecordSuccessfulLogin(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	lockout := NewLoginLockout(LoginLockoutConfig{
		MaxAttempts:  5,
		LockDuration: 5 * time.Minute,
		WindowSize:   5 * time.Minute,
		RedisClient:  client,
		KeyPrefix:    "test:",
	}, log.DefaultLogger)

	ctx := context.Background()
	identifier := "test@example.com"

	// 记录几次失败
	lockout.RecordFailedAttempt(ctx, identifier)
	lockout.RecordFailedAttempt(ctx, identifier)

	// 检查剩余尝试次数
	remaining, err := lockout.GetRemainingAttempts(ctx, identifier)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 3 {
		t.Errorf("expected 3 remaining attempts, got %d", remaining)
	}

	// 成功登录
	err = lockout.RecordSuccessfulLogin(ctx, identifier)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 检查计数已清除
	remaining, _ = lockout.GetRemainingAttempts(ctx, identifier)
	if remaining != 5 {
		t.Errorf("expected 5 remaining attempts after successful login, got %d", remaining)
	}
}

func TestLoginLockout_Unlock(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	lockout := NewLoginLockout(LoginLockoutConfig{
		MaxAttempts:  2,
		LockDuration: 5 * time.Minute,
		WindowSize:   5 * time.Minute,
		RedisClient:  client,
		KeyPrefix:    "test:",
	}, log.DefaultLogger)

	ctx := context.Background()
	identifier := "test@example.com"

	// 触发锁定
	lockout.RecordFailedAttempt(ctx, identifier)
	lockout.RecordFailedAttempt(ctx, identifier)

	// 确认已锁定
	locked, _, _ := lockout.IsLocked(ctx, identifier)
	if !locked {
		t.Fatal("should be locked")
	}

	// 解锁
	err := lockout.Unlock(ctx, identifier)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 确认已解锁
	locked, _, _ = lockout.IsLocked(ctx, identifier)
	if locked {
		t.Error("should not be locked after unlock")
	}
}

func TestLoginLockout_NoRedis(t *testing.T) {
	// 测试没有 Redis 客户端的情况
	lockout := NewLoginLockout(LoginLockoutConfig{
		MaxAttempts:  5,
		LockDuration: 5 * time.Minute,
		RedisClient:  nil, // 没有 Redis
	}, log.DefaultLogger)

	ctx := context.Background()
	identifier := "test@example.com"

	// 所有操作应该正常返回，不报错
	locked, _, err := lockout.IsLocked(ctx, identifier)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if locked {
		t.Error("should not be locked without Redis")
	}

	attempts, err := lockout.RecordFailedAttempt(ctx, identifier)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if attempts != 0 {
		t.Errorf("expected 0 attempts without Redis, got %d", attempts)
	}

	remaining, err := lockout.GetRemainingAttempts(ctx, identifier)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if remaining != 5 {
		t.Errorf("expected max attempts without Redis, got %d", remaining)
	}
}

func TestLoginLockout_DefaultConfig(t *testing.T) {
	client, cleanup := setupTestRedis(t)
	defer cleanup()

	// 使用空配置，应该使用默认值
	lockout := NewLoginLockout(LoginLockoutConfig{
		RedisClient: client,
	}, log.DefaultLogger)

	if lockout.config.MaxAttempts != 5 {
		t.Errorf("expected default MaxAttempts 5, got %d", lockout.config.MaxAttempts)
	}
	if lockout.config.LockDuration != 15*time.Minute {
		t.Errorf("expected default LockDuration 15m, got %v", lockout.config.LockDuration)
	}
	if lockout.config.KeyPrefix != "login_lockout:" {
		t.Errorf("expected default KeyPrefix, got %s", lockout.config.KeyPrefix)
	}
}
