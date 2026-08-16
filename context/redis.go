package context

import (
	stlcontext "context"

	"github.com/gin-gonic/gin"
	"github.com/ranxx/go-infra/redis"
)

type rdbKey struct{}

// WithRedis 将 Redis 客户端注入上下文
func WithRedis(ctx Context, rdb redis.RedisClient) Context {
	return stlcontext.WithValue(ctx, rdbKey{}, rdb)
}

// GetRedis 从上下文获取 Redis 客户端；没有则回退到全局单例 redis.Get()
func GetRedis(ctx Context) redis.RedisClient {
	if gc, ok := ctx.(*gin.Context); ok {
		return GinCtxGetRedis(gc)
	}
	v := ctx.Value(rdbKey{})
	if v != nil {
		if rdb, ok := v.(redis.RedisClient); ok {
			return rdb
		}
	}
	return redis.Get()
}

// GinCtxWithRedis 将 Redis 客户端注入 gin.Context
func GinCtxWithRedis(ctx *gin.Context, rdb redis.RedisClient) *gin.Context {
	ctx.Set("-gin-redis-", rdb)
	return ctx
}

// GinCtxGetRedis 从 gin.Context 取 Redis，没有则回退到全局单例
func GinCtxGetRedis(ctx *gin.Context) redis.RedisClient {
	v, ok := ctx.Get("-gin-redis-")
	if ok {
		if rdb, ok := v.(redis.RedisClient); ok {
			return rdb
		}
	}
	return redis.Get()
}