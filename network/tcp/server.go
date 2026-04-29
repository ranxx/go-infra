package tcp

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ranxx/go-infra/logger"
	"github.com/ranxx/go-infra/network"
)

// Server TCP 服务器
type Server struct {
	opts              network.Options
	connectionManager network.ConnectionManager
	listener          net.Listener
	running           bool
	runningMux        sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	close             chan struct{}
}

// NewServer 创建新的 TCP 服务器
func NewServer(opts ...network.Option) *Server {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	if options.Manager == nil {
		options.Manager = network.NewConnectionManager(opts...)
	}

	return &Server{
		opts:              options,
		connectionManager: options.Manager,
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

// Start 启动服务器
func (s *Server) Start(ctx context.Context) error {
	s.runningMux.Lock()
	defer s.runningMux.Unlock()

	if s.running {
		return nil
	}

	listener, err := net.Listen("tcp", s.opts.Address)
	if err != nil {
		return fmt.Errorf("TCP 监听失败: %w", err)
	}

	s.listener = listener
	s.running = true
	s.ctx, s.cancel = context.WithCancel(ctx)

	logger.GetFieldLogger().Infof("TCP 服务器启动成功，监听地址: %s", s.opts.Address)

	go s.acceptConnections()
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

	if s.listener != nil {
		s.listener.Close()
	}

	return s.connectionManager.Close()
}

// IsRunning 检查服务器是否运行中
func (s *Server) IsRunning() bool {
	s.runningMux.RLock()
	defer s.runningMux.RUnlock()
	return s.running
}

// acceptConnections 接受新连接
func (s *Server) acceptConnections() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.close:
			return
		default:
			s.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
			conn, err := s.listener.Accept()
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				logger.GetFieldLogger().Errorf("接受连接失败: %v", err)
				continue
			}
			go s.handleNewConnection(conn)
		}
	}
}

// handleNewConnection 处理新连接
func (s *Server) handleNewConnection(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()

	tcpConn := NewConnection(s.ctx, remoteAddr, conn, s.close,
		network.WithCoder(s.opts.Coder),
		network.WithPacker(s.opts.Packer),
		network.WithReadTimeout(s.opts.ReadTimeout),
		network.WithMaxMessageSize(s.opts.MaxMessageSize),
		network.WithMaxErrorCount(s.opts.MaxErrorCount),
	)
	tcpConn.SetConnectionManager(s.connectionManager)

	if err := s.connectionManager.AddConnection(tcpConn); err != nil {
		conn.Close()
		logger.GetFieldLogger().Errorf("添加连接失败: %v", err)
		return
	}

	if s.opts.Handler != nil {
		if err := s.opts.Handler.OnConnect(s.ctx, tcpConn); err != nil {
			s.connectionManager.RemoveConnection(remoteAddr)
			conn.Close()
			logger.GetFieldLogger().Errorf("连接建立回调失败: %v", err)
			return
		}
	}

	go tcpConn.StartHandling(s.ctx)
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
