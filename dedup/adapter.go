package dedup

import (
	"context"
	"time"

	infraredis "github.com/ranxx/go-infra/redis"
)

// NewFromGoInfraRedis 从 go-infra/redis.RedisClient 创建去重器
func NewFromGoInfraRedis(client infraredis.RedisClient, ttl time.Duration) *Dedup {
	return New(&redisAdapter{client: client}, ttl)
}

// redisAdapter 将 go-infra/redis.RedisClient 适配为 dedup.Client
type redisAdapter struct {
	client infraredis.RedisClient
}

func (a *redisAdapter) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return a.client.SetNX(ctx, key, value, expiration).Result()
}
