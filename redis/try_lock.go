package redis

import (
	"context"
	"time"
)

// TryLock 尝试获取锁（简化版）
func (c *RedisClientImpl) TryLock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	ok, err := c.SetNX(ctx, key, "1", ttl).Result()
	return ok, err
}

// Unlock 释放锁（简化版）
func (c *RedisClientImpl) Unlock(ctx context.Context, key string) error {
	return c.Del(ctx, key).Err()
}
