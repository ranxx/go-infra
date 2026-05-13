package network

import (
	"context"
	"net"
	"time"
)

// Connection 表示一个连接
type Connection interface {
	// GetID 获取连接唯一标识
	GetID() string
	// SetID 设置连接唯一标识
	SetID(id string)
	// GetMetadata 获取连接元数据
	GetMetadata(key any) (any, bool)
	// SetMetadata 设置连接元数据
	SetMetadata(key, value any)
	// GetMetadatas 获取所有连接元数据
	GetMetadatas() map[any]any
	// Send 发送消息
	Send(data any) error
	// Close 关闭连接
	Close() error
	// IsClosed 检查连接是否已关闭
	IsClosed() bool
	// GetRemoteAddr 获取远程地址
	GetRemoteAddr() net.Addr
	// GetConnectedAt 获取连接时间
	GetConnectedAt() time.Time
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	// OnConnect 连接建立时的回调
	OnConnect(ctx context.Context, conn Connection) error
	// OnMessage 收到消息时的回调
	OnMessage(ctx context.Context, conn Connection, data any) error
	// OnDisconnect 连接断开时的回调
	OnDisconnect(ctx context.Context, conn Connection, err error)
}

// MessageHandlerFunc 函数类型实现 MessageHandler
type MessageHandlerFunc struct {
	OnConnectFn     func(ctx context.Context, conn Connection) error
	OnMessageFn     func(ctx context.Context, conn Connection, data any) error
	OnDisconnectFn  func(ctx context.Context, conn Connection, err error)
}

func (f *MessageHandlerFunc) OnConnect(ctx context.Context, conn Connection) error {
	if f.OnConnectFn != nil {
		return f.OnConnectFn(ctx, conn)
	}
	return nil
}

func (f *MessageHandlerFunc) OnMessage(ctx context.Context, conn Connection, data any) error {
	if f.OnMessageFn != nil {
		return f.OnMessageFn(ctx, conn, data)
	}
	return nil
}

func (f *MessageHandlerFunc) OnDisconnect(ctx context.Context, conn Connection, err error) {
	if f.OnDisconnectFn != nil {
		f.OnDisconnectFn(ctx, conn, err)
	}
}

// ConnectionManager 连接管理器接口
type ConnectionManager interface {
	// AddConnection 添加连接
	AddConnection(conn Connection) error
	// AuthenticateConnection 认证连接
	AuthenticateConnection(conn Connection) error
	// RemoveConnection 移除连接
	RemoveConnection(id string) error
	// GetConnection 根据ID获取连接
	GetConnection(id string) (Connection, bool)
	// GetConnections 获取所有连接
	GetConnections(ids ...string) []Connection
	// Broadcast 广播消息给所有连接
	Broadcast(val any, ids ...string)
	// GetConnectionCount 获取连接数量
	GetConnectionCount() int
	// Close 关闭所有连接
	Close() error
}

// Server 服务器接口
type Server interface {
	// SetMessageHandler 设置消息处理器
	SetMessageHandler(handler MessageHandler)
	// GetConnectionManager 获取连接管理器
	GetConnectionManager() ConnectionManager
	// Start 启动服务器
	Start(ctx context.Context) error
	// Stop 停止服务器
	Stop(ctx context.Context) error
	// IsRunning 检查服务器是否运行中
	IsRunning() bool
}

// Coder 编码器接口
type Coder interface {
	Unmarshal(data []byte) (interface{}, error)
	Marshal(v interface{}) ([]byte, error)
}

// Packer 打包器接口
type Packer interface {
	GetHeadLength() uint32
	UnpackHead([]byte) (uint32, error)
	Pack([]byte) ([]byte, error)
}

// Options 选项
type Options struct {
	Address        string
	Coder          Coder
	Packer         Packer
	MaxConnections int
	Handler        MessageHandler
	Manager        ConnectionManager

	MaxMessageSize     int64         // 最大消息字节数
	MinHeartbeatPeriod time.Duration // 最小允许的心跳周期
	MaxErrorCount      int           // 最大累积错误次数
	ReadTimeout        time.Duration // 读取超时时间
	Metadata           map[any]any   // 连接元数据
}

// Option 选项函数
type Option func(*Options)

// WithMetadata 设置连接元数据
func WithMetadata(metadata map[any]any) Option {
	return func(o *Options) {
		o.Metadata = metadata
	}
}

// WithReadTimeout 设置读取超时时间
func WithReadTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = timeout
	}
}

// WithMaxMessageSize 设置最大消息字节数
func WithMaxMessageSize(size int64) Option {
	return func(o *Options) {
		o.MaxMessageSize = size
	}
}

// WithMinHeartbeatPeriod 设置最小允许的心跳周期
func WithMinHeartbeatPeriod(period time.Duration) Option {
	return func(o *Options) {
		o.MinHeartbeatPeriod = period
	}
}

// WithMaxErrorCount 设置最大累积错误次数
func WithMaxErrorCount(count int) Option {
	return func(o *Options) {
		o.MaxErrorCount = count
	}
}

// WithAddress 设置地址
func WithAddress(address string) Option {
	return func(o *Options) {
		o.Address = address
	}
}

// WithCoder 设置编码器
func WithCoder(coder Coder) Option {
	return func(o *Options) {
		o.Coder = coder
	}
}

// WithPacker 设置打包器
func WithPacker(packer Packer) Option {
	return func(o *Options) {
		o.Packer = packer
	}
}

// WithMaxConnections 设置最大连接数
func WithMaxConnections(maxConn int) Option {
	return func(o *Options) {
		o.MaxConnections = maxConn
	}
}

// WithMessageHandler 设置消息处理器
func WithMessageHandler(handler MessageHandler) Option {
	return func(o *Options) {
		o.Handler = handler
	}
}

// WithConnectionManager 设置连接管理器
func WithConnectionManager(manager ConnectionManager) Option {
	return func(o *Options) {
		o.Manager = manager
	}
}
