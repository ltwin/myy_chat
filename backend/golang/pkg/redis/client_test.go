package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfig tests configuration defaults
func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, "localhost:6379", cfg.Addr)
	assert.Equal(t, "", cfg.Password)
	assert.Equal(t, 0, cfg.DB)
	assert.Equal(t, 3, cfg.MaxRetries)
	assert.Equal(t, 5, cfg.MinIdleConns)
	assert.Equal(t, 100, cfg.PoolSize)
	assert.Equal(t, 4*time.Second, cfg.PoolTimeout)
	assert.Equal(t, 5*time.Second, cfg.DialTimeout)
	assert.Equal(t, 3*time.Second, cfg.ReadTimeout)
	assert.Equal(t, 3*time.Second, cfg.WriteTimeout)
}

// TestErrors tests error definitions
func TestErrors(t *testing.T) {
	assert.EqualError(t, ErrNil, "redis: nil")
	assert.EqualError(t, ErrNotFound, "redis: key not found")
	assert.EqualError(t, ErrClosed, "redis: client is closed")
	assert.EqualError(t, ErrConnFailed, "redis: connection failed")
}

// Integration tests - require a running Redis instance
// Skip if Redis is not available

func skipIfNoRedis(t *testing.T) *Client {
	cfg := DefaultConfig()
	client, err := NewClient(cfg)
	if err != nil {
		t.Skip("Redis not available, skipping integration test")
	}
	return client
}

func TestNewClient_Success(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	assert.NotNil(t, client)
	assert.NotNil(t, client.Raw())
}

func TestNewClient_InvalidAddr(t *testing.T) {
	cfg := Config{
		Addr:        "invalid:9999",
		DialTimeout: 100 * time.Millisecond,
	}
	_, err := NewClient(cfg)
	assert.ErrorIs(t, err, ErrConnFailed)
}

func TestClient_Ping(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	err := client.Ping(ctx)
	assert.NoError(t, err)
}

func TestClient_Close(t *testing.T) {
	client := skipIfNoRedis(t)

	err := client.Close()
	assert.NoError(t, err)
}

func TestClient_Close_AlreadyClosed(t *testing.T) {
	cfg := DefaultConfig()
	client := &Client{rdb: nil}

	err := client.Close()
	assert.ErrorIs(t, err, ErrClosed)

	// Suppress unused variable warning
	_ = cfg
}

// String operations tests

func TestClient_SetGet(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:setget:" + time.Now().Format("20060102150405")

	// Set
	err := client.Set(ctx, key, "hello", time.Minute)
	require.NoError(t, err)

	// Get
	val, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "hello", val)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

func TestClient_Get_NotFound(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	_, err := client.Get(ctx, "nonexistent:key:12345")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestClient_SetNX(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:setnx:" + time.Now().Format("20060102150405")

	// First SetNX should succeed
	ok, err := client.SetNX(ctx, key, "value1", time.Minute)
	require.NoError(t, err)
	assert.True(t, ok)

	// Second SetNX should fail
	ok, err = client.SetNX(ctx, key, "value2", time.Minute)
	require.NoError(t, err)
	assert.False(t, ok)

	// Value should be the first one
	val, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

func TestClient_Del(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key1 := "test:del:1:" + time.Now().Format("20060102150405")
	key2 := "test:del:2:" + time.Now().Format("20060102150405")

	// Set keys
	_ = client.Set(ctx, key1, "v1", time.Minute)
	_ = client.Set(ctx, key2, "v2", time.Minute)

	// Delete
	count, err := client.Del(ctx, key1, key2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestClient_Exists(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:exists:" + time.Now().Format("20060102150405")

	// Key doesn't exist
	count, err := client.Exists(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Set key
	_ = client.Set(ctx, key, "value", time.Minute)

	// Key exists
	count, err = client.Exists(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

func TestClient_Expire_TTL(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:expire:" + time.Now().Format("20060102150405")

	// Set key without expiration
	_ = client.Set(ctx, key, "value", 0)

	// Set expiration
	ok, err := client.Expire(ctx, key, 10*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)

	// Check TTL
	ttl, err := client.TTL(ctx, key)
	require.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= 10*time.Second)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

func TestClient_IncrDecr(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:incr:" + time.Now().Format("20060102150405")

	// Incr (creates key with value 1)
	val, err := client.Incr(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// IncrBy
	val, err = client.IncrBy(ctx, key, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(6), val)

	// Decr
	val, err = client.Decr(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(5), val)

	// DecrBy
	val, err = client.DecrBy(ctx, key, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(2), val)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

// Hash operations tests

func TestClient_Hash(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:hash:" + time.Now().Format("20060102150405")

	// HSet
	err := client.HSet(ctx, key, "field1", "value1", "field2", "value2")
	require.NoError(t, err)

	// HGet
	val, err := client.HGet(ctx, key, "field1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	// HGet not found
	_, err = client.HGet(ctx, key, "nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)

	// HGetAll
	all, err := client.HGetAll(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"field1": "value1", "field2": "value2"}, all)

	// HExists
	exists, err := client.HExists(ctx, key, "field1")
	require.NoError(t, err)
	assert.True(t, exists)

	// HDel
	count, err := client.HDel(ctx, key, "field1")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

// List operations tests

func TestClient_List(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:list:" + time.Now().Format("20060102150405")

	// LPush
	count, err := client.LPush(ctx, key, "a", "b", "c")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// RPush
	count, err = client.RPush(ctx, key, "d")
	require.NoError(t, err)
	assert.Equal(t, int64(4), count)

	// LLen
	length, err := client.LLen(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(4), length)

	// LRange
	vals, err := client.LRange(ctx, key, 0, -1)
	require.NoError(t, err)
	assert.Equal(t, []string{"c", "b", "a", "d"}, vals)

	// LPop
	val, err := client.LPop(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "c", val)

	// RPop
	val, err = client.RPop(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, "d", val)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

func TestClient_LPop_Empty(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	_, err := client.LPop(ctx, "nonexistent:list")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestClient_RPop_Empty(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	_, err := client.RPop(ctx, "nonexistent:list")
	assert.ErrorIs(t, err, ErrNotFound)
}

// Set operations tests

func TestClient_Set(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:set:" + time.Now().Format("20060102150405")

	// SAdd
	count, err := client.SAdd(ctx, key, "a", "b", "c")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// SCard
	card, err := client.SCard(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(3), card)

	// SIsMember
	isMember, err := client.SIsMember(ctx, key, "a")
	require.NoError(t, err)
	assert.True(t, isMember)

	isMember, err = client.SIsMember(ctx, key, "x")
	require.NoError(t, err)
	assert.False(t, isMember)

	// SMembers
	members, err := client.SMembers(ctx, key)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, members)

	// SRem
	count, err = client.SRem(ctx, key, "a")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

// Sorted Set operations tests

func TestClient_ZSet(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:zset:" + time.Now().Format("20060102150405")

	// ZAdd
	count, err := client.ZAdd(ctx, key,
		redis.Z{Score: 1.0, Member: "a"},
		redis.Z{Score: 2.0, Member: "b"},
		redis.Z{Score: 3.0, Member: "c"},
	)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)

	// ZCard
	card, err := client.ZCard(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(3), card)

	// ZScore
	score, err := client.ZScore(ctx, key, "b")
	require.NoError(t, err)
	assert.Equal(t, 2.0, score)

	// ZRange
	members, err := client.ZRange(ctx, key, 0, -1)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, members)

	// ZRangeByScore
	members, err = client.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: "1",
		Max: "2",
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, members)

	// ZRem
	count, err = client.ZRem(ctx, key, "a")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

// MGet/MSet tests

func TestClient_MGetMSet(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	prefix := "test:mget:" + time.Now().Format("20060102150405")
	key1, key2 := prefix+":1", prefix+":2"

	// MSet
	err := client.MSet(ctx, key1, "v1", key2, "v2")
	require.NoError(t, err)

	// MGet
	vals, err := client.MGet(ctx, key1, key2, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"v1", "v2", nil}, vals)

	// Cleanup
	_, _ = client.Del(ctx, key1, key2)
}

// Pipeline tests

func TestClient_Pipeline(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:pipeline:" + time.Now().Format("20060102150405")

	pipe := client.Pipeline()
	pipe.Set(ctx, key, "value", time.Minute)
	pipe.Get(ctx, key)
	pipe.Del(ctx, key)

	cmds, err := pipe.Exec(ctx)
	require.NoError(t, err)
	assert.Len(t, cmds, 3)
}

func TestClient_TxPipeline(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:txpipeline:" + time.Now().Format("20060102150405")

	pipe := client.TxPipeline()
	pipe.Set(ctx, key, "value", time.Minute)
	pipe.Get(ctx, key)
	pipe.Del(ctx, key)

	cmds, err := pipe.Exec(ctx)
	require.NoError(t, err)
	assert.Len(t, cmds, 3)
}

// Pub/Sub tests

func TestClient_PubSub(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	channel := "test:pubsub:" + time.Now().Format("20060102150405")

	// Subscribe
	pubsub := client.Subscribe(ctx, channel)
	defer pubsub.Close()

	// Wait for subscription confirmation
	_, err := pubsub.Receive(ctx)
	require.NoError(t, err)

	// Publish
	count, err := client.Publish(ctx, channel, "hello")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Receive message
	msg, err := pubsub.ReceiveMessage(ctx)
	require.NoError(t, err)
	assert.Equal(t, "hello", msg.Payload)
}

// Script tests

func TestClient_Eval(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()

	// Simple script
	result := client.Eval(ctx, "return 1+1", nil)
	val, err := result.Int()
	require.NoError(t, err)
	assert.Equal(t, 1+1, val)
}

func TestClient_Eval_WithKeys(t *testing.T) {
	client := skipIfNoRedis(t)
	defer client.Close()

	ctx := context.Background()
	key := "test:eval:" + time.Now().Format("20060102150405")

	// Set a value first
	_ = client.Set(ctx, key, "hello", time.Minute)

	// Script that reads the key
	result := client.Eval(ctx, "return redis.call('GET', KEYS[1])", []string{key})
	val, err := result.Text()
	require.NoError(t, err)
	assert.Equal(t, "hello", val)

	// Cleanup
	_, _ = client.Del(ctx, key)
}

// Benchmark tests

func BenchmarkClient_Set(b *testing.B) {
	cfg := DefaultConfig()
	client, err := NewClient(cfg)
	if err != nil {
		b.Skip("Redis not available")
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Set(ctx, "bench:set", "value", time.Minute)
	}
}

func BenchmarkClient_Get(b *testing.B) {
	cfg := DefaultConfig()
	client, err := NewClient(cfg)
	if err != nil {
		b.Skip("Redis not available")
	}
	defer client.Close()

	ctx := context.Background()
	_ = client.Set(ctx, "bench:get", "value", time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = client.Get(ctx, "bench:get")
	}
}

func BenchmarkClient_Pipeline(b *testing.B) {
	cfg := DefaultConfig()
	client, err := NewClient(cfg)
	if err != nil {
		b.Skip("Redis not available")
	}
	defer client.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipe := client.Pipeline()
		pipe.Set(ctx, "bench:pipe:1", "v1", time.Minute)
		pipe.Set(ctx, "bench:pipe:2", "v2", time.Minute)
		pipe.Set(ctx, "bench:pipe:3", "v3", time.Minute)
		_, _ = pipe.Exec(ctx)
	}
}
