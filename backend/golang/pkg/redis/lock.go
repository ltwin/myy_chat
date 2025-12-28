// Package redis provides Redis client and utilities
package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

// Lock errors
var (
	ErrLockNotAcquired = errors.New("redis: lock not acquired")
	ErrLockNotHeld     = errors.New("redis: lock not held")
)

// Lock represents a distributed lock
type Lock struct {
	client *Client
	key    string
	value  string
	ttl    time.Duration
}

// LockOptions configures lock behavior
type LockOptions struct {
	// TTL is the lock expiration time
	TTL time.Duration
	// RetryCount is the number of retry attempts
	RetryCount int
	// RetryDelay is the delay between retries
	RetryDelay time.Duration
}

// DefaultLockOptions returns default lock options
func DefaultLockOptions() LockOptions {
	return LockOptions{
		TTL:        30 * time.Second,
		RetryCount: 0,
		RetryDelay: 100 * time.Millisecond,
	}
}

// AcquireLock attempts to acquire a distributed lock
func (c *Client) AcquireLock(ctx context.Context, key string, opts LockOptions) (*Lock, error) {
	// Generate unique lock value
	value, err := generateLockValue()
	if err != nil {
		return nil, err
	}

	lock := &Lock{
		client: c,
		key:    lockKey(key),
		value:  value,
		ttl:    opts.TTL,
	}

	// Try to acquire lock
	for attempt := 0; attempt <= opts.RetryCount; attempt++ {
		ok, err := c.SetNX(ctx, lock.key, lock.value, lock.ttl)
		if err != nil {
			return nil, err
		}
		if ok {
			return lock, nil
		}

		// If no more retries, fail
		if attempt == opts.RetryCount {
			break
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(opts.RetryDelay):
		}
	}

	return nil, ErrLockNotAcquired
}

// TryLock attempts to acquire a lock without retrying
func (c *Client) TryLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	return c.AcquireLock(ctx, key, LockOptions{
		TTL:        ttl,
		RetryCount: 0,
	})
}

// Release releases the lock
func (l *Lock) Release(ctx context.Context) error {
	// Lua script to release lock only if we hold it
	// This ensures atomic check-and-delete
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result := l.client.Eval(ctx, script, []string{l.key}, l.value)
	deleted, err := result.Int64()
	if err != nil {
		return err
	}

	if deleted == 0 {
		return ErrLockNotHeld
	}

	return nil
}

// Extend extends the lock TTL
func (l *Lock) Extend(ctx context.Context, ttl time.Duration) error {
	// Lua script to extend lock only if we hold it
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result := l.client.Eval(ctx, script, []string{l.key}, l.value, ttl.Milliseconds())
	extended, err := result.Int64()
	if err != nil {
		return err
	}

	if extended == 0 {
		return ErrLockNotHeld
	}

	l.ttl = ttl
	return nil
}

// TTL returns the remaining TTL of the lock
func (l *Lock) TTL(ctx context.Context) (time.Duration, error) {
	return l.client.TTL(ctx, l.key)
}

// Key returns the lock key
func (l *Lock) Key() string {
	return l.key
}

// WithLock executes a function while holding a lock
func (c *Client) WithLock(ctx context.Context, key string, opts LockOptions, fn func(ctx context.Context) error) error {
	lock, err := c.AcquireLock(ctx, key, opts)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	return fn(ctx)
}

// Mutex provides a higher-level mutex interface
type Mutex struct {
	client *Client
	key    string
	opts   LockOptions
	lock   *Lock
}

// NewMutex creates a new mutex
func (c *Client) NewMutex(key string, opts LockOptions) *Mutex {
	return &Mutex{
		client: c,
		key:    key,
		opts:   opts,
	}
}

// Lock acquires the mutex
func (m *Mutex) Lock(ctx context.Context) error {
	lock, err := m.client.AcquireLock(ctx, m.key, m.opts)
	if err != nil {
		return err
	}
	m.lock = lock
	return nil
}

// TryLock attempts to acquire the mutex without blocking
func (m *Mutex) TryLock(ctx context.Context) bool {
	lock, err := m.client.TryLock(ctx, m.key, m.opts.TTL)
	if err != nil {
		return false
	}
	m.lock = lock
	return true
}

// Unlock releases the mutex
func (m *Mutex) Unlock(ctx context.Context) error {
	if m.lock == nil {
		return ErrLockNotHeld
	}
	err := m.lock.Release(ctx)
	m.lock = nil
	return err
}

// Extend extends the mutex TTL
func (m *Mutex) Extend(ctx context.Context, ttl time.Duration) error {
	if m.lock == nil {
		return ErrLockNotHeld
	}
	return m.lock.Extend(ctx, ttl)
}

// Helper functions

func lockKey(key string) string {
	return "lock:" + key
}

func generateLockValue() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Semaphore provides a distributed counting semaphore
type Semaphore struct {
	client   *Client
	key      string
	maxCount int64
	ttl      time.Duration
}

// NewSemaphore creates a new semaphore
func (c *Client) NewSemaphore(key string, maxCount int64, ttl time.Duration) *Semaphore {
	return &Semaphore{
		client:   c,
		key:      "semaphore:" + key,
		maxCount: maxCount,
		ttl:      ttl,
	}
}

// Acquire attempts to acquire a semaphore slot
func (s *Semaphore) Acquire(ctx context.Context) (string, error) {
	token, err := generateLockValue()
	if err != nil {
		return "", err
	}

	now := time.Now().UnixNano()
	deadline := now + s.ttl.Nanoseconds()

	// Lua script for atomic semaphore acquisition
	script := `
		-- Remove expired entries
		redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", ARGV[1])

		-- Check current count
		local count = redis.call("ZCARD", KEYS[1])
		if count < tonumber(ARGV[2]) then
			-- Add new entry
			redis.call("ZADD", KEYS[1], ARGV[3], ARGV[4])
			return 1
		else
			return 0
		end
	`

	result := s.client.Eval(ctx, script, []string{s.key}, now, s.maxCount, deadline, token)
	acquired, err := result.Int64()
	if err != nil {
		return "", err
	}

	if acquired == 0 {
		return "", ErrLockNotAcquired
	}

	return token, nil
}

// Release releases a semaphore slot
func (s *Semaphore) Release(ctx context.Context, token string) error {
	_, err := s.client.ZRem(ctx, s.key, token)
	return err
}

// Count returns the current number of acquired slots
func (s *Semaphore) Count(ctx context.Context) (int64, error) {
	// Remove expired entries first
	now := time.Now().UnixNano()
	_, _ = s.client.Raw().ZRemRangeByScore(ctx, s.key, "-inf", string(rune(now))).Result()

	return s.client.ZCard(ctx, s.key)
}

// RateLimiter provides a sliding window rate limiter
type RateLimiter struct {
	client    *Client
	keyPrefix string
	limit     int64
	window    time.Duration
}

// NewRateLimiter creates a new rate limiter
func (c *Client) NewRateLimiter(keyPrefix string, limit int64, window time.Duration) *RateLimiter {
	return &RateLimiter{
		client:    c,
		keyPrefix: "ratelimit:" + keyPrefix,
		limit:     limit,
		window:    window,
	}
}

// Allow checks if a request is allowed and records it
func (r *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	fullKey := r.keyPrefix + ":" + key
	now := time.Now().UnixNano()
	windowStart := now - r.window.Nanoseconds()

	// Lua script for atomic rate limiting
	script := `
		-- Remove entries outside the window
		redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", ARGV[1])

		-- Count entries in the window
		local count = redis.call("ZCARD", KEYS[1])

		if count < tonumber(ARGV[2]) then
			-- Add new entry
			redis.call("ZADD", KEYS[1], ARGV[3], ARGV[3])
			-- Set expiry on the key
			redis.call("PEXPIRE", KEYS[1], ARGV[4])
			return 1
		else
			return 0
		end
	`

	result := r.client.Eval(ctx, script, []string{fullKey}, windowStart, r.limit, now, r.window.Milliseconds())
	allowed, err := result.Int64()
	if err != nil {
		return false, err
	}

	return allowed == 1, nil
}

// Remaining returns the number of remaining requests allowed
func (r *RateLimiter) Remaining(ctx context.Context, key string) (int64, error) {
	fullKey := r.keyPrefix + ":" + key
	now := time.Now().UnixNano()
	windowStart := now - r.window.Nanoseconds()

	// Remove expired entries
	_, _ = r.client.Raw().ZRemRangeByScore(ctx, fullKey, "-inf", string(rune(windowStart))).Result()

	count, err := r.client.ZCard(ctx, fullKey)
	if err != nil {
		return 0, err
	}

	remaining := r.limit - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// Reset resets the rate limiter for a key
func (r *RateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := r.keyPrefix + ":" + key
	_, err := r.client.Del(ctx, fullKey)
	return err
}
