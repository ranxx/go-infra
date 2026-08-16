// Package dedup 提供基于 Redis 的发布去重工具。
// 用于多 Collector 实例并行采集时，确保同一条数据只被发布一次到 NATS。
package dedup

import (
	"context"
	"fmt"
	"time"
)

// Client 去重所需的 Redis 最小接口
type Client interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
}

// Dedup 基于 Redis 的发布去重器
type Dedup struct {
	client Client
	ttl    time.Duration
}

// New 创建去重器
func New(client Client, ttl time.Duration) *Dedup {
	return &Dedup{client: client, ttl: ttl}
}

// TryPublish 尝试原子去重发布。
// 使用 Redis SET NX EX 原子操作:
//   - key 不存在: 创建 key, 执行 publish, 返回 (true, nil)
//   - key 已存在: 不执行 publish, 返回 (false, nil)
//   - Redis 错误: 不执行 publish, 返回 (false, error)
//
// 调用方应在返回 error 时执行降级逻辑 (直接发布, 不做去重)。
func (d *Dedup) TryPublish(
	ctx context.Context,
	dedupKey string,
	publish func() error,
) (bool, error) {
	ok, err := d.client.SetNX(ctx, dedupKey, "1", d.ttl)
	if err != nil {
		return false, fmt.Errorf("dedup setnx: %w", err)
	}
	if !ok {
		return false, nil
	}
	if err := publish(); err != nil {
		return true, fmt.Errorf("dedup publish: %w", err)
	}
	return true, nil
}

// Key 生成去重键，包含 {exchange}:{instType} 隔离。
// 格式: tick:{exchange}:{instType}:{symbol}:{dataType}:{id}:{ts_ns}
func Key(exchange, instType, symbol, dataType string, id int64, ts time.Time) string {
	return fmt.Sprintf("tick:%s:%s:%s:%s:%d:%d",
		exchange, instType, symbol, dataType, id, ts.UnixNano())
}

// KeyString 生成字符串型去重键。
// 用于不需要 int64 id 的场景 (如 depth 的 lastUpdateId 作为字符串)。
func KeyString(exchange, instType, symbol, dataType, uniqueID string) string {
	return fmt.Sprintf("tick:%s:%s:%s:%s:%s",
		exchange, instType, symbol, dataType, uniqueID)
}
