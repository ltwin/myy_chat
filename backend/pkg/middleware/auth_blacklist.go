package middleware

import (
	"context"
	"strconv"
	"time"

	redisclient "github.com/myy-chat/backend/pkg/redis"
)

const (
	authBlacklistSIDPrefix = "auth:blacklist:sid:"
	authBlacklistJTIPrefix = "auth:blacklist:jti:"
	authBlacklistMarker    = "1"
)

// RedisTokenBlacklist 基于 Redis 的 token 黑名单实现。
type RedisTokenBlacklist struct {
	client *redisclient.Client
}

func NewRedisTokenBlacklist(client *redisclient.Client) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{client: client}
}

func (b *RedisTokenBlacklist) IsRevoked(ctx context.Context, sid int64, jti string) (bool, error) {
	keys := make([]string, 0, 2)
	if sid > 0 {
		keys = append(keys, sidBlacklistKey(sid))
	}
	if jti != "" {
		keys = append(keys, jtiBlacklistKey(jti))
	}
	if len(keys) == 0 {
		return false, nil
	}

	n, err := b.client.Exists(ctx, keys...)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (b *RedisTokenBlacklist) RevokeSID(ctx context.Context, sid int64, ttl time.Duration) error {
	if sid <= 0 || ttl <= 0 {
		return nil
	}
	return b.client.Set(ctx, sidBlacklistKey(sid), authBlacklistMarker, ttl)
}

func (b *RedisTokenBlacklist) RevokeJTI(ctx context.Context, jti string, ttl time.Duration) error {
	if jti == "" || ttl <= 0 {
		return nil
	}
	return b.client.Set(ctx, jtiBlacklistKey(jti), authBlacklistMarker, ttl)
}

func sidBlacklistKey(sid int64) string {
	return authBlacklistSIDPrefix + strconv.FormatInt(sid, 10)
}

func jtiBlacklistKey(jti string) string {
	return authBlacklistJTIPrefix + jti
}
