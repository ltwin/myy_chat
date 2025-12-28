// Package redis provides Redis client and utilities
package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Common errors
var (
	ErrNil       = errors.New("redis: nil")
	ErrNotFound  = errors.New("redis: key not found")
	ErrClosed    = errors.New("redis: client is closed")
	ErrConnFailed = errors.New("redis: connection failed")
)

// Config holds Redis connection configuration
type Config struct {
	// Addr is the Redis server address (host:port)
	Addr string
	// Password is the Redis password
	Password string
	// DB is the Redis database number
	DB int
	// MaxRetries is the maximum number of retries
	MaxRetries int
	// MinIdleConns is the minimum number of idle connections
	MinIdleConns int
	// PoolSize is the maximum number of socket connections
	PoolSize int
	// PoolTimeout is the timeout for waiting for a connection from the pool
	PoolTimeout time.Duration
	// DialTimeout is the timeout for establishing new connections
	DialTimeout time.Duration
	// ReadTimeout is the timeout for socket reads
	ReadTimeout time.Duration
	// WriteTimeout is the timeout for socket writes
	WriteTimeout time.Duration
}

// DefaultConfig returns a default Redis configuration
func DefaultConfig() Config {
	return Config{
		Addr:         "localhost:6379",
		Password:     "",
		DB:           0,
		MaxRetries:   3,
		MinIdleConns: 5,
		PoolSize:     100,
		PoolTimeout:  4 * time.Second,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

// Client wraps the Redis client with additional functionality
type Client struct {
	rdb *redis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		MinIdleConns: cfg.MinIdleConns,
		PoolSize:     cfg.PoolSize,
		PoolTimeout:  cfg.PoolTimeout,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, ErrConnFailed
	}

	return &Client{rdb: rdb}, nil
}

// Close closes the Redis client
func (c *Client) Close() error {
	if c.rdb == nil {
		return ErrClosed
	}
	return c.rdb.Close()
}

// Ping checks if the connection is alive
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// Raw returns the underlying Redis client for advanced operations
func (c *Client) Raw() *redis.Client {
	return c.rdb
}

// String operations

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return val, err
}

// Set sets a key-value pair
func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

// SetNX sets a key-value pair only if the key does not exist
func (c *Client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, value, expiration).Result()
}

// Del deletes one or more keys
func (c *Client) Del(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Del(ctx, keys...).Result()
}

// Exists checks if keys exist
func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	return c.rdb.Exists(ctx, keys...).Result()
}

// Expire sets a timeout on key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.rdb.Expire(ctx, key, expiration).Result()
}

// TTL returns the remaining time to live of a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

// Incr increments the number stored at key by one
func (c *Client) Incr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

// IncrBy increments the number stored at key by increment
func (c *Client) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.IncrBy(ctx, key, value).Result()
}

// Decr decrements the number stored at key by one
func (c *Client) Decr(ctx context.Context, key string) (int64, error) {
	return c.rdb.Decr(ctx, key).Result()
}

// DecrBy decrements the number stored at key by decrement
func (c *Client) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.rdb.DecrBy(ctx, key, value).Result()
}

// Hash operations

// HGet gets the value of a hash field
func (c *Client) HGet(ctx context.Context, key, field string) (string, error) {
	val, err := c.rdb.HGet(ctx, key, field).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return val, err
}

// HSet sets the value of a hash field
func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.rdb.HSet(ctx, key, values...).Err()
}

// HGetAll gets all the fields and values in a hash
func (c *Client) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.rdb.HGetAll(ctx, key).Result()
}

// HDel deletes one or more hash fields
func (c *Client) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return c.rdb.HDel(ctx, key, fields...).Result()
}

// HExists determines if a hash field exists
func (c *Client) HExists(ctx context.Context, key, field string) (bool, error) {
	return c.rdb.HExists(ctx, key, field).Result()
}

// List operations

// LPush inserts values at the head of the list
func (c *Client) LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.rdb.LPush(ctx, key, values...).Result()
}

// RPush inserts values at the tail of the list
func (c *Client) RPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return c.rdb.RPush(ctx, key, values...).Result()
}

// LPop removes and returns the first element of the list
func (c *Client) LPop(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.LPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return val, err
}

// RPop removes and returns the last element of the list
func (c *Client) RPop(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.RPop(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return val, err
}

// LRange returns a range of elements from the list
func (c *Client) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.LRange(ctx, key, start, stop).Result()
}

// LLen returns the length of the list
func (c *Client) LLen(ctx context.Context, key string) (int64, error) {
	return c.rdb.LLen(ctx, key).Result()
}

// Set operations

// SAdd adds members to a set
func (c *Client) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.rdb.SAdd(ctx, key, members...).Result()
}

// SRem removes members from a set
func (c *Client) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.rdb.SRem(ctx, key, members...).Result()
}

// SMembers returns all members of a set
func (c *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.rdb.SMembers(ctx, key).Result()
}

// SIsMember checks if a member is in a set
func (c *Client) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.rdb.SIsMember(ctx, key, member).Result()
}

// SCard returns the number of members in a set
func (c *Client) SCard(ctx context.Context, key string) (int64, error) {
	return c.rdb.SCard(ctx, key).Result()
}

// Sorted Set operations

// ZAdd adds members to a sorted set
func (c *Client) ZAdd(ctx context.Context, key string, members ...redis.Z) (int64, error) {
	return c.rdb.ZAdd(ctx, key, members...).Result()
}

// ZRem removes members from a sorted set
func (c *Client) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return c.rdb.ZRem(ctx, key, members...).Result()
}

// ZRange returns a range of members in a sorted set by index
func (c *Client) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.rdb.ZRange(ctx, key, start, stop).Result()
}

// ZRangeByScore returns a range of members in a sorted set by score
func (c *Client) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) ([]string, error) {
	return c.rdb.ZRangeByScore(ctx, key, opt).Result()
}

// ZScore returns the score of a member in a sorted set
func (c *Client) ZScore(ctx context.Context, key, member string) (float64, error) {
	return c.rdb.ZScore(ctx, key, member).Result()
}

// ZCard returns the number of members in a sorted set
func (c *Client) ZCard(ctx context.Context, key string) (int64, error) {
	return c.rdb.ZCard(ctx, key).Result()
}

// Pub/Sub operations

// Publish publishes a message to a channel
func (c *Client) Publish(ctx context.Context, channel string, message interface{}) (int64, error) {
	return c.rdb.Publish(ctx, channel, message).Result()
}

// Subscribe subscribes to channels
func (c *Client) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return c.rdb.Subscribe(ctx, channels...)
}

// Script operations

// Eval evaluates a Lua script
func (c *Client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd {
	return c.rdb.Eval(ctx, script, keys, args...)
}

// EvalSha evaluates a Lua script by its SHA1 digest
func (c *Client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return c.rdb.EvalSha(ctx, sha1, keys, args...)
}

// Pipeline operations

// Pipeline returns a pipelined client
func (c *Client) Pipeline() redis.Pipeliner {
	return c.rdb.Pipeline()
}

// TxPipeline returns a transaction pipelined client
func (c *Client) TxPipeline() redis.Pipeliner {
	return c.rdb.TxPipeline()
}

// Helper functions

// MGet retrieves multiple values by keys
func (c *Client) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	return c.rdb.MGet(ctx, keys...).Result()
}

// MSet sets multiple key-value pairs
func (c *Client) MSet(ctx context.Context, values ...interface{}) error {
	return c.rdb.MSet(ctx, values...).Err()
}
