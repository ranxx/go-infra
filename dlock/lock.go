package dlock

import (
	"context"
	"fmt"
	"sync"
	"time"

	redisV8 "github.com/go-redis/redis/v8"
	"github.com/ranxx/go-infra/key"
	"github.com/ranxx/go-infra/redis"
)

// 错误定义
var (
	ErrLockNotAcquired = fmt.Errorf("lock not acquired")
	ErrRedisNotReady   = fmt.Errorf("redis not ready")
)

// Locker 分布式锁接口
type Locker interface {
	// Acquire 尝试获取锁，返回锁对象（需要 defer 释放）
	Acquire(ctx context.Context, key string, ttl time.Duration) (Lock, error)
	// TryAcquireWithRetry 尝试获取锁，支持重试
	TryAcquireWithRetry(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) (Lock, error)
	// TryAcquirePeriodLock 尝试获取周期锁（整个周期只执行一次）
	TryAcquirePeriodLock(ctx context.Context, key string, periodID string, ttl time.Duration) (Lock, error)
}

// Lock 锁接口
type Lock interface {
	// Unlock 释放锁
	Unlock(ctx context.Context) error
	// Extend 续期锁
	Extend(ctx context.Context, ttl time.Duration) error
}

// EmptyLock 空锁实现（当锁不可用时返回，调用方无需处理 nil）
type emptyLock struct{}

// EmptyLock 获取一个空锁实例
func EmptyLock() Lock {
	return &emptyLock{}
}
func (l *emptyLock) Unlock(ctx context.Context) error                    { return nil }
func (l *emptyLock) Extend(ctx context.Context, ttl time.Duration) error { return nil }

var (
	_client *RedisLocker
	once    sync.Once
)

// RedisLocker 基于 Redis 的分布式锁实现
type RedisLocker struct {
	client redis.RedisClient
	k      key.Keyer
}

// Init 初始化全局 RedisLocker 实例（单例）
func Init(client redis.RedisClient, k key.Keyer) *RedisLocker {
	once.Do(func() {
		_client = NewRedisLocker(client, k)
	})
	return _client
}

// GetLocker 获取全局 RedisLocker 实例
func GetLocker() *RedisLocker {
	return _client
}

// NewRedisLocker 创建 Redis 锁
func NewRedisLocker(client redis.RedisClient, k key.Keyer) *RedisLocker {
	return &RedisLocker{client: client, k: k}
}

// Acquire 尝试获取锁
// 使用 Lua 脚本保证原子性：只有锁不存在或值匹配时才设置
func (l *RedisLocker) Acquire(ctx context.Context, key string, ttl time.Duration) (Lock, error) {
	// 检查 Redis 连接
	if l.client == nil {
		return nil, ErrRedisNotReady
	}

	if l.k != nil {
		key = l.k.Key(key)
	}

	// 使用时间戳+随机数作为锁的值，防止误释放
	lockValue := fmt.Sprintf("lock:%s:%d", key, time.Now().UnixNano())

	ok, err := l.client.SetNX(ctx, key, lockValue, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}
	if !ok {
		return nil, ErrLockNotAcquired
	}

	return &redisLock{
		client: l.client,
		key:    key,
		value:  lockValue,
	}, nil
}

// TryAcquireWithRetry 尝试获取锁，支持重试
func (l *RedisLocker) TryAcquireWithRetry(ctx context.Context, key string, ttl time.Duration, maxRetries int, retryInterval time.Duration) (Lock, error) {
	for i := 0; i < maxRetries; i++ {
		lock, err := l.Acquire(ctx, key, ttl)
		if err == nil {
			return lock, nil
		}
		if err == ErrLockNotAcquired {
			time.Sleep(retryInterval)
			continue
		}
		return nil, err
	}
	return nil, ErrLockNotAcquired
}

// TryAcquirePeriodLock 尝试获取周期锁
// 使用场景：定时任务，每个周期只想让一个 Pod 执行一次
// 原理：使用 periodID 作为锁值，整个周期内有效，周期结束后自动过期
func (l *RedisLocker) TryAcquirePeriodLock(ctx context.Context, key string, periodID string, ttl time.Duration) (Lock, error) {
	if l.client == nil {
		return nil, ErrRedisNotReady
	}

	// SET key periodID NX PX ttl
	// 只有 key 不存在时才设置成功
	ok, err := l.client.SetNX(ctx, key, periodID, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}
	if !ok {
		return nil, ErrLockNotAcquired
	}

	return &redisLock{
		client: l.client,
		key:    key,
		value:  periodID,
	}, nil
}

// redisLock Redis 锁实现
type redisLock struct {
	client redis.RedisClient
	key    string
	value  string
}

// Extend 续期锁
func (l *redisLock) Extend(ctx context.Context, ttl time.Duration) error {
	script := redisV8.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("pexpire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`)
	return script.Run(ctx, l.client.GetClient(), []string{l.key}, l.value, ttl.Milliseconds()).Err()
}

// Unlock 释放锁
// 使用 Lua 脚本：只有值匹配时才删除，防止误删
func (l *redisLock) Unlock(ctx context.Context) error {
	script := redisV8.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`)
	return script.Run(ctx, l.client.GetClient(), []string{l.key}, l.value).Err()
}
