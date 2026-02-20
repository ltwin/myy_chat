package redis

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultLockOptions(t *testing.T) {
	opts := DefaultLockOptions()

	assert.Equal(t, 30*time.Second, opts.TTL)
	assert.Equal(t, 0, opts.RetryCount)
	assert.Equal(t, 100*time.Millisecond, opts.RetryDelay)
}

func TestLockErrors(t *testing.T) {
	assert.EqualError(t, ErrLockNotAcquired, "redis: lock not acquired")
	assert.EqualError(t, ErrLockNotHeld, "redis: lock not held")
}

func TestGenerateLockValue(t *testing.T) {
	v1, err := generateLockValue()
	require.NoError(t, err)
	assert.Len(t, v1, 32) // 16 bytes = 32 hex chars

	v2, err := generateLockValue()
	require.NoError(t, err)
	assert.NotEqual(t, v1, v2)
}

func TestLockKey(t *testing.T) {
	assert.Equal(t, "lock:mykey", lockKey("mykey"))
	assert.Equal(t, "lock:user:123", lockKey("user:123"))
}

// Integration tests

func TestClient_TryLock(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:trylock:" + time.Now().Format("20060102150405")

	// Acquire lock
	lock, err := client.TryLock(ctx, key, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, lock)

	// Try to acquire again - should fail
	_, err = client.TryLock(ctx, key, time.Minute)
	assert.ErrorIs(t, err, ErrLockNotAcquired)

	// Release lock
	err = lock.Release(ctx)
	require.NoError(t, err)

	// Now should be able to acquire
	lock2, err := client.TryLock(ctx, key, time.Minute)
	require.NoError(t, err)
	require.NotNil(t, lock2)

	// Cleanup
	_ = lock2.Release(ctx)
}

func TestClient_AcquireLock_WithRetry(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:acquireretry:" + time.Now().Format("20060102150405")

	// Acquire lock with short TTL
	lock1, err := client.TryLock(ctx, key, 200*time.Millisecond)
	require.NoError(t, err)

	// Try to acquire with retries - should succeed after lock expires
	opts := LockOptions{
		TTL:        time.Minute,
		RetryCount: 5,
		RetryDelay: 100 * time.Millisecond,
	}
	lock2, err := client.AcquireLock(ctx, key, opts)
	require.NoError(t, err)
	require.NotNil(t, lock2)

	// Cleanup
	_ = lock1.Release(ctx) // This may fail since lock expired
	_ = lock2.Release(ctx)
}

func TestClient_AcquireLock_ContextCancelled(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	key := "test:lockcancel:" + time.Now().Format("20060102150405")

	// Acquire lock first
	ctx := context.Background()
	lock1, err := client.TryLock(ctx, key, time.Minute)
	require.NoError(t, err)
	defer lock1.Release(ctx)

	// Create context that cancels quickly
	ctx2, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	opts := LockOptions{
		TTL:        time.Minute,
		RetryCount: 10,
		RetryDelay: 100 * time.Millisecond,
	}

	_, err = client.AcquireLock(ctx2, key, opts)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestLock_Release(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:release:" + time.Now().Format("20060102150405")

	lock, err := client.TryLock(ctx, key, time.Minute)
	require.NoError(t, err)

	// Release
	err = lock.Release(ctx)
	require.NoError(t, err)

	// Release again should fail
	err = lock.Release(ctx)
	assert.ErrorIs(t, err, ErrLockNotHeld)
}

func TestLock_Extend(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:extend:" + time.Now().Format("20060102150405")

	lock, err := client.TryLock(ctx, key, 1*time.Second)
	require.NoError(t, err)
	defer lock.Release(ctx)

	// Extend TTL
	err = lock.Extend(ctx, 10*time.Second)
	require.NoError(t, err)

	// Check TTL is extended
	ttl, err := lock.TTL(ctx)
	require.NoError(t, err)
	assert.True(t, ttl > 5*time.Second)
}

func TestLock_Extend_NotHeld(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:extend:notheld:" + time.Now().Format("20060102150405")

	lock, err := client.TryLock(ctx, key, time.Minute)
	require.NoError(t, err)

	// Release lock
	err = lock.Release(ctx)
	require.NoError(t, err)

	// Try to extend
	err = lock.Extend(ctx, time.Minute)
	assert.ErrorIs(t, err, ErrLockNotHeld)
}

func TestLock_Key(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:lockkey:" + time.Now().Format("20060102150405")

	lock, err := client.TryLock(ctx, key, time.Minute)
	require.NoError(t, err)
	defer lock.Release(ctx)

	assert.Equal(t, "lock:"+key, lock.Key())
}

func TestClient_WithLock(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:withlock:" + time.Now().Format("20060102150405")

	executed := false
	err := client.WithLock(ctx, key, DefaultLockOptions(), func(ctx context.Context) error {
		executed = true
		return nil
	})

	require.NoError(t, err)
	assert.True(t, executed)

	// Lock should be released
	lock, err := client.TryLock(ctx, key, time.Minute)
	require.NoError(t, err)
	_ = lock.Release(ctx)
}

// Mutex tests

func TestMutex_LockUnlock(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:mutex:" + time.Now().Format("20060102150405")

	mutex := client.NewMutex(key, DefaultLockOptions())

	// Lock
	err := mutex.Lock(ctx)
	require.NoError(t, err)

	// Unlock
	err = mutex.Unlock(ctx)
	require.NoError(t, err)

	// Unlock again should fail
	err = mutex.Unlock(ctx)
	assert.ErrorIs(t, err, ErrLockNotHeld)
}

func TestMutex_TryLock(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:mutex:trylock:" + time.Now().Format("20060102150405")

	mutex := client.NewMutex(key, DefaultLockOptions())

	// TryLock should succeed
	ok := mutex.TryLock(ctx)
	assert.True(t, ok)

	// Another mutex trying to lock same key should fail
	mutex2 := client.NewMutex(key, DefaultLockOptions())
	ok = mutex2.TryLock(ctx)
	assert.False(t, ok)

	// Cleanup
	_ = mutex.Unlock(ctx)
}

func TestMutex_Extend(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:mutex:extend:" + time.Now().Format("20060102150405")

	mutex := client.NewMutex(key, LockOptions{TTL: time.Second})

	err := mutex.Lock(ctx)
	require.NoError(t, err)
	defer mutex.Unlock(ctx)

	err = mutex.Extend(ctx, 10*time.Second)
	require.NoError(t, err)
}

func TestMutex_Extend_NotHeld(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:mutex:extend:notheld:" + time.Now().Format("20060102150405")

	mutex := client.NewMutex(key, DefaultLockOptions())

	// Extend without locking should fail
	err := mutex.Extend(ctx, time.Second)
	assert.ErrorIs(t, err, ErrLockNotHeld)
}

// Concurrency test

func TestLock_Concurrent(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:concurrent:" + time.Now().Format("20060102150405")

	var counter int64
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			opts := LockOptions{
				TTL:        time.Second,
				RetryCount: 100,
				RetryDelay: 10 * time.Millisecond,
			}

			lock, err := client.AcquireLock(ctx, key, opts)
			if err != nil {
				return
			}
			defer lock.Release(ctx)

			// Critical section
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(10), counter)
}

// Semaphore tests

func TestSemaphore_AcquireRelease(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:semaphore:" + time.Now().Format("20060102150405")

	sem := client.NewSemaphore(key, 2, time.Minute)

	// Acquire first slot
	token1, err := sem.Acquire(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, token1)

	// Acquire second slot
	token2, err := sem.Acquire(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, token2)

	// Third acquire should fail
	_, err = sem.Acquire(ctx)
	assert.ErrorIs(t, err, ErrLockNotAcquired)

	// Release one slot
	err = sem.Release(ctx, token1)
	require.NoError(t, err)

	// Now acquire should succeed
	token3, err := sem.Acquire(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, token3)

	// Cleanup
	_ = sem.Release(ctx, token2)
	_ = sem.Release(ctx, token3)
}

// Rate limiter tests

func TestRateLimiter_Allow(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	prefix := "test:ratelimit:" + time.Now().Format("20060102150405")

	limiter := client.NewRateLimiter(prefix, 3, time.Second)
	key := "user123"

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, key)
		require.NoError(t, err)
		assert.True(t, allowed, "request %d should be allowed", i+1)
	}

	// 4th request should be denied
	allowed, err := limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Wait for window to pass
	time.Sleep(1100 * time.Millisecond)

	// Now should be allowed again
	allowed, err = limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.True(t, allowed)

	// Cleanup
	_ = limiter.Reset(ctx, key)
}

func TestRateLimiter_Remaining(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	prefix := "test:ratelimit:remaining:" + time.Now().Format("20060102150405")

	limiter := client.NewRateLimiter(prefix, 5, time.Minute)
	key := "user456"

	// Initially 5 remaining
	remaining, err := limiter.Remaining(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(5), remaining)

	// Use 2
	_, _ = limiter.Allow(ctx, key)
	_, _ = limiter.Allow(ctx, key)

	// Should have 3 remaining
	remaining, err = limiter.Remaining(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(3), remaining)

	// Cleanup
	_ = limiter.Reset(ctx, key)
}

func TestRateLimiter_Reset(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	prefix := "test:ratelimit:reset:" + time.Now().Format("20060102150405")

	limiter := client.NewRateLimiter(prefix, 2, time.Minute)
	key := "user789"

	// Use up the limit
	_, _ = limiter.Allow(ctx, key)
	_, _ = limiter.Allow(ctx, key)

	// Should be denied
	allowed, err := limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.False(t, allowed)

	// Reset
	err = limiter.Reset(ctx, key)
	require.NoError(t, err)

	// Should be allowed again
	allowed, err = limiter.Allow(ctx, key)
	require.NoError(t, err)
	assert.True(t, allowed)
}

// Benchmark tests

func BenchmarkClient_TryLock(b *testing.B) {
	cfg := DefaultConfig()
	client, err := NewClient(cfg)
	if err != nil {
		b.Skip("Redis not available")
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lock, _ := client.TryLock(ctx, "bench:lock", time.Second)
		if lock != nil {
			_ = lock.Release(ctx)
		}
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	cfg := DefaultConfig()
	client, err := NewClient(cfg)
	if err != nil {
		b.Skip("Redis not available")
	}
	defer client.Close()

	ctx := context.Background()
	limiter := client.NewRateLimiter("bench:ratelimit", 10000, time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = limiter.Allow(ctx, "user")
	}
}
