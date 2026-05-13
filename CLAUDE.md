# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 概述

这是一个基础设施库（`github.com/ranxx/go-infra`），提供 Go 应用的通用组件：Redis、MySQL、MongoDB、Elasticsearch、gRPC、TCP/WebSocket 网络、任务调度、日志、分布式锁和代理支持。

## 构建与测试命令

```bash
# 运行所有测试
go test ./...

# 运行测试（详细输出）
go test -v ./...

# 运行单个包的测试
go test -v ./config/...

# 构建（无 main 包，纯库）
go build ./...
```

## 架构

### 包结构

- **`config/`** — 配置管理，提供 `Provider` 接口和热加载支持（`Reloadable`）。通过 `default` 结构体标签应用默认值。`LoadByKey` 按优先级链式加载配置（回退模式）。

- **`logger/`** — 基于 Logrus 的日志模块。`Init()` 接收函数式选项，可选添加 Elasticsearch Hook。通过 `GetLogger()` / `GetFieldLogger()` 获取全局日志实例。

- **`task/`** — 基于 `robfig/cron` 的任务调度器。使用 `Schedule` 接口支持自定义时间策略，`DurationSchedule` 实现间隔调度。通过 `sync.Once` 保证单例初始化。

- **`network/`** — 抽象网络层，支持 TCP 和 WebSocket。核心接口定义在 `network.go`，默认实现分散在各子包。

- **`grpc/`** — gRPC 服务端封装，带 unary 拦截器实现 trace ID 传播和请求日志。实现 `Register` 接口注册服务，默认注册健康检查和 reflection 服务。

- **`redis/`**、**`mysql/`**、**`mongo/`**、**`elasticsearch/`** — 各存储客户端封装。MySQL 使用 GORM，Redis 支持通过 `proxy.Wrap()` 配置代理拨号。

- **`proxy/`** — SOCKS5/HTTP 代理支持，从 `ALL_PROXY` / `no_proxy` 环境变量读取配置。

- **`dlock/`** — 分布式锁实现。

- **`message/`** — 消息编解码，支持 `JSON` 和 `Protobuf` 两种实现。

- **`eventbus/`** — 类型安全进程内事件总线，基于泛型。`Subscribe[T]` 返回只读 channel，`Publish[T]` 非阻塞发送。

- **`tracer/`** — Trace ID 传播工具。通过 `context.Context` 传递 traceId，默认 gRPC header `x-trace-id`。

- **`interceptor/`** — gRPC 拦截器工厂。`TraceUnaryInterceptor()` 自动注入 traceId 到 gRPC metadata。

- **`key/`** — 带前缀的分隔键生成器（`Keyer`），用于缓存 key、topic 名称等。

- **`topic/`** — 分布式主题封装（基于 NATS），提供 `Topic` + `Register` 接口，自动 JSON 编解码。

- **`nats/`** — NATS 客户端封装（预留空包，具体逻辑在 `topic/` 中）。

- **`utils/`** — 工具函数：hash、math、rand、slices、stack、time、uuid。

### 关键设计模式

- **函数式选项（Functional Options）** — 所有主要组件接受 `Option` 函数类型进行配置（如 `task.Option`、`network.Option`）。
- **单例模式** — `task.Init()` 使用 `sync.Once` 确保只初始化一次。
- **接口契约** — `Connection`、`MessageHandler`、`Server`、`Provider`、`Reloadable` 定义了组件边界，便于替换实现。
- **配置默认值** — `ApplyDefaults()` 读取 `default` 标签填充零值，`LoadByKey()` 链式调用 providers 实现回退加载。

## Network 模块详解

### 目录结构

```
network/
├── network.go       # 核心接口定义
├── connection.go    # BaseConnection 基类
├── manager.go       # DefaultConnectionManager
├── tcp/
│   ├── server.go    # TCP Server
│   └── connection.go # TCP Connection
└── ws/
    ├── server.go    # WebSocket Server
    └── connection.go # WebSocket Connection
```

### 核心接口

```go
// Connection 连接接口
type Connection interface {
    GetID() string
    SetID(id string)
    GetMetadata(key any) (any, bool)
    SetMetadata(key, value any)
    GetMetadatas() map[any]any
    Send(data any) error
    Close() error
    IsClosed() bool
    GetRemoteAddr() net.Addr
    GetConnectedAt() time.Time
}

// MessageHandler 消息处理器
type MessageHandler interface {
    OnConnect(ctx context.Context, conn Connection) error
    OnMessage(ctx context.Context, conn Connection, data any) error
    OnDisconnect(ctx context.Context, conn Connection, err error)
}

// ConnectionManager 连接管理器
type ConnectionManager interface {
    AddConnection(conn Connection) error
    AuthenticateConnection(conn Connection) error
    RemoveConnection(id string) error
    GetConnection(id string) (Connection, bool)
    GetConnections(ids ...string) []Connection
    Broadcast(val any, ids ...string)
    GetConnectionCount() int
    Close() error
}

// Server 服务器接口
type Server interface {
    SetMessageHandler(handler MessageHandler)
    GetConnectionManager() ConnectionManager
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
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
```

### 使用示例

**TCP 服务器：**

```go
tcpServer := tcp.NewServer(
    network.WithAddress(":8080"),
    network.WithMaxConnections(10000),
    network.WithCoder(yourCoder),
    network.WithPacker(yourPacker),
    network.WithMessageHandler(&network.MessageHandlerFunc{
        OnConnectFn: func(ctx context.Context, conn network.Connection) error {
            return nil
        },
        OnMessageFn: func(ctx context.Context, conn network.Connection, data any) error {
            return nil
        },
        OnDisconnectFn: func(ctx context.Context, conn network.Connection, err error) {
        },
    }),
)

tcpServer.Start(context.Background())
defer tcpServer.Stop(context.Background())
```

**WebSocket 服务器：**

```go
wsServer := ws.NewServer(
    network.WithAddress(":8081"),
    network.WithMessageHandler(yourHandler),
)

wsServer.Start(context.Background())
defer wsServer.Stop(context.Background())

// 在 HTTP handler 中处理 WebSocket 升级
http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
    wsServer.HandleWebSocket(w, r, uniqueID, metadata)
})
```
