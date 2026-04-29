package ws

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ranxx/go-infra/network"
)

// Server WebSocket 服务器
type Server struct {
	opts              network.Options
	connectionManager network.ConnectionManager
	upgrader          *websocket.Upgrader
	running           bool
	runningMux        sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	close             chan struct{}
}

// NewServer 创建新的 WebSocket 服务器
func NewServer(opts ...network.Option) *Server {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if options.Manager == nil {
		options.Manager = network.NewConnectionManager(opts...)
	}

	upgrader := &websocket.Upgrader{
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return &Server{
		opts:              options,
		connectionManager: options.Manager,
		upgrader:          upgrader,
		close:             make(chan struct{}),
	}
}

// SetMessageHandler 设置消息处理器
func (s *Server) SetMessageHandler(handler network.MessageHandler) {
	s.opts.Handler = handler
}

// GetConnectionManager 获取连接管理器
func (s *Server) GetConnectionManager() network.ConnectionManager {
	return s.connectionManager
}

// HandleWebSocket 处理 WebSocket 升级和连接
func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request, id string, metadata map[any]any) error {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade failed: %w", err)
	}

	if s.opts.MaxMessageSize > 0 {
		conn.SetReadLimit(s.opts.MaxMessageSize)
	}

	if s.ctx == nil {
		s.ctx = context.Background()
	}

	wsConn := NewConnection(s.ctx, id, conn, s.close, metadata,
		network.WithCoder(s.opts.Coder),
		network.WithReadTimeout(s.opts.ReadTimeout),
		network.WithMaxMessageSize(s.opts.MaxMessageSize),
		network.WithMaxErrorCount(s.opts.MaxErrorCount),
	)
	wsConn.SetConnectionManager(s.connectionManager)

	if err := s.connectionManager.AddConnection(wsConn); err != nil {
		conn.Close()
		return err
	}

	if s.opts.Handler != nil {
		if err := s.opts.Handler.OnConnect(s.ctx, wsConn); err != nil {
			s.connectionManager.RemoveConnection(id)
			conn.Close()
			return err
		}
	}

	go wsConn.StartHandling(s.ctx)

	return nil
}

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	s.runningMux.Lock()
	defer s.runningMux.Unlock()

	if s.running {
		return nil
	}

	s.running = true
	s.ctx, s.cancel = context.WithCancel(ctx)

	go s.cleanupRoutine()

	return nil
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	s.runningMux.Lock()
	defer s.runningMux.Unlock()

	if !s.running {
		return nil
	}

	s.running = false
	close(s.close)
	if s.cancel != nil {
		s.cancel()
	}

	return s.connectionManager.Close()
}

// IsRunning 检查服务器是否运行中
func (s *Server) IsRunning() bool {
	s.runningMux.RLock()
	defer s.runningMux.RUnlock()
	return s.running
}

// cleanupRoutine 定期清理已关闭的连接
func (s *Server) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.close:
			return
		case <-ticker.C:
			connections := s.connectionManager.GetConnections()
			for _, conn := range connections {
				if conn.IsClosed() {
					s.connectionManager.RemoveConnection(conn.GetID())
				}
			}
		}
	}
}

func defaultOptions() network.Options {
	return network.Options{
		Address:        ":8080",
		ReadTimeout:    60 * time.Second,
		MaxMessageSize: 1024 * 1024,
	}
}
