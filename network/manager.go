package network

import (
	"errors"
	"sync"
)

var (
	ErrConnectionNotFound   = errors.New("connection not found")
	ErrMaxConnections      = errors.New("max connections reached")
	ErrInvalidConnection   = errors.New("invalid connection")
)

// DefaultConnectionManager 默认连接管理器实现
type DefaultConnectionManager struct {
	unauthedConnections map[string]Connection // remoteAddr 作为 key
	connections         map[string]Connection
	mutex               sync.RWMutex
	maxConnections      int
}

// NewConnectionManager 创建默认连接管理器
func NewConnectionManager(opts ...Option) *DefaultConnectionManager {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	maxConn := options.MaxConnections
	if maxConn <= 0 {
		maxConn = 10000
	}

	return &DefaultConnectionManager{
		unauthedConnections: make(map[string]Connection),
		connections:         make(map[string]Connection),
		maxConnections:      maxConn,
	}
}

// AddConnection 添加连接
func (m *DefaultConnectionManager) AddConnection(conn Connection) error {
	if conn == nil {
		return ErrInvalidConnection
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.connections) >= m.maxConnections {
		return ErrMaxConnections
	}

	m.unauthedConnections[conn.GetRemoteAddr().String()] = conn
	return nil
}

// AuthenticateConnection 认证连接
func (m *DefaultConnectionManager) AuthenticateConnection(conn Connection) error {
	if conn == nil {
		return ErrInvalidConnection
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	remoteAddr := conn.GetRemoteAddr().String()
	delete(m.unauthedConnections, remoteAddr)

	m.connections[conn.GetID()] = conn
	return nil
}

// RemoveConnection 移除连接
func (m *DefaultConnectionManager) RemoveConnection(id string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	conn, exists := m.connections[id]
	if !exists {
		// 尝试从未认证连接中查找
		for _, c := range m.unauthedConnections {
			if c.GetID() == id {
				delete(m.unauthedConnections, c.GetRemoteAddr().String())
				c.Close()
				return nil
			}
		}
		return ErrConnectionNotFound
	}

	delete(m.connections, id)
	delete(m.unauthedConnections, conn.GetRemoteAddr().String())
	conn.Close()

	return nil
}

// GetConnection 根据ID获取连接
func (m *DefaultConnectionManager) GetConnection(id string) (Connection, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	conn, exists := m.connections[id]
	return conn, exists
}

// GetConnections 获取所有连接
func (m *DefaultConnectionManager) GetConnections(ids ...string) []Connection {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(ids) > 0 {
		connections := make([]Connection, 0, len(ids))
		for _, id := range ids {
			if conn, ok := m.connections[id]; ok && !conn.IsClosed() {
				connections = append(connections, conn)
			}
		}
		return connections
	}

	connections := make([]Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		if !conn.IsClosed() {
			connections = append(connections, conn)
		}
	}

	return connections
}

// Broadcast 广播消息给所有连接
func (m *DefaultConnectionManager) Broadcast(val any, ids ...string) {
	connections := m.GetConnections(ids...)
	for _, conn := range connections {
		conn.Send(val)
	}
}

// GetConnectionCount 获取连接数量
func (m *DefaultConnectionManager) GetConnectionCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	count := 0
	for _, conn := range m.connections {
		if !conn.IsClosed() {
			count++
		}
	}

	return count
}

// Close 关闭所有连接
func (m *DefaultConnectionManager) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var lastErr error
	for id, conn := range m.connections {
		if err := conn.Close(); err != nil {
			lastErr = err
		}
		delete(m.connections, id)
	}

	for addr, conn := range m.unauthedConnections {
		if err := conn.Close(); err != nil {
			lastErr = err
		}
		delete(m.unauthedConnections, addr)
	}

	return lastErr
}
