package network

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// BaseConnection 基础连接实现，提供公共功能
type BaseConnection struct {
	id          string
	closed      bool
	closedMux   sync.RWMutex
	connectedAt time.Time
	metadata    map[any]any
	metadataMux sync.RWMutex

	// 限制相关
	errorCount int
	limitMux   sync.Mutex
	maxErrorCount int
}

// NewBaseConnection 创建基础连接
func NewBaseConnection(id string, maxErrorCount int) *BaseConnection {
	return &BaseConnection{
		id:            id,
		connectedAt:   time.Now(),
		metadata:      make(map[any]any),
		maxErrorCount: maxErrorCount,
	}
}

// GetID 获取连接唯一标识
func (c *BaseConnection) GetID() string {
	return c.id
}

// SetID 设置连接唯一标识
func (c *BaseConnection) SetID(id string) {
	c.id = id
}

// AddErrorCount 增加错误次数并检查是否超过限制
func (c *BaseConnection) AddErrorCount() bool {
	if c.maxErrorCount <= 0 {
		return false
	}
	c.limitMux.Lock()
	defer c.limitMux.Unlock()
	c.errorCount++
	return c.errorCount >= c.maxErrorCount
}

// GetMetadata 获取连接元数据
func (c *BaseConnection) GetMetadata(key any) (any, bool) {
	c.metadataMux.RLock()
	defer c.metadataMux.RUnlock()
	val, ok := c.metadata[key]
	return val, ok
}

// SetMetadata 设置连接元数据
func (c *BaseConnection) SetMetadata(key, value any) {
	c.metadataMux.Lock()
	defer c.metadataMux.Unlock()
	c.metadata[key] = value
}

// GetMetadatas 获取所有连接元数据
func (c *BaseConnection) GetMetadatas() map[any]any {
	c.metadataMux.RLock()
	defer c.metadataMux.RUnlock()
	result := make(map[any]any, len(c.metadata))
	for k, v := range c.metadata {
		result[k] = v
	}
	return result
}

// IsClosed 检查连接是否已关闭
func (c *BaseConnection) IsClosed() bool {
	c.closedMux.RLock()
	defer c.closedMux.RUnlock()
	return c.closed
}

// MarkClosed 标记连接已关闭
func (c *BaseConnection) MarkClosed() {
	c.closedMux.Lock()
	defer c.closedMux.Unlock()
	c.closed = true
}

// GetConnectedAt 获取连接时间
func (c *BaseConnection) GetConnectedAt() time.Time {
	return c.connectedAt
}

// GetRemoteAddr 获取远程地址 - 子类需要实现
func (c *BaseConnection) GetRemoteAddr() net.Addr {
	return nil
}

// Send 发送消息 - 子类需要实现
func (c *BaseConnection) Send(_data any) error {
	return fmt.Errorf("not implemented")
}

// Close 关闭连接 - 子类需要实现
func (c *BaseConnection) Close() error {
	return fmt.Errorf("not implemented")
}
