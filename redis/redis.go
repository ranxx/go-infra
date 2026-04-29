package redis

import (
	"context"
	"net"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ranxx/go-infra/proxy"
)

type RedisClient interface {
	GetClient() *redis.Client
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Exists(ctx context.Context, keys ...string) *redis.IntCmd
	Incr(ctx context.Context, key string) *redis.IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd
}

type RedisClientImpl struct {
	*redis.Client
}

func (c *RedisClientImpl) GetClient() *redis.Client {
	return c.Client
}

func NewClient(cfg *Config) RedisClient {
	opts := &redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	if cfg.Proxy {
		proxy.Wrap(func(dr proxy.Dialer) {
			opts.Dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dr.Dial(network, addr)
			}
		})
	}

	rdb := redis.NewClient(opts)
	return &RedisClientImpl{rdb}
}
