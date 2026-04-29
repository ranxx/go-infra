package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ranxx/go-infra/tracer"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

// Register 服务注册接口
type Register interface {
	Register(s *grpc.Server)
}

// ServiceRegistrar 服务注册接口
type ServiceRegistrar func(s *grpc.Server)

// Server gRPC 服务器
type Server struct {
	config     Config
	grpcServer *grpc.Server
	logger     logrus.FieldLogger
}

// NewServer 创建 gRPC 服务器实例
// config 必须包含 Port 字段（使用匿名接口）
// services 用于注册具体业务服务
func NewServer(config Config, logger logrus.FieldLogger, services ...Register) *Server {
	s := &Server{
		config: config,
		logger: logger,
	}
	s.grpcServer = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
	)
	// 注册业务服务
	for _, service := range services {
		service.Register(s.grpcServer)
	}
	// 注册标准 Health Check 服务 (K8s gRPC 探针)
	grpc_health_v1.RegisterHealthServer(s.grpcServer, &HealthServer{})
	reflection.Register(s.grpcServer)
	return s
}

func (s *Server) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()

	// 1. 从 metadata 获取或生成 traceID
	md, ok := metadata.FromIncomingContext(ctx)
	headerName := tracer.GetHeaderName()
	var traceID string
	if ok {
		vals := md.Get(headerName)
		if len(vals) > 0 {
			traceID = vals[0]
		}
	}
	if traceID == "" {
		traceID = uuid.New().String()
	}
	ctx = metadata.AppendToOutgoingContext(ctx, headerName, traceID)

	// 2. 创建带 traceID 的日志
	log := s.logger.WithFields(logrus.Fields{
		"trace_id": traceID,
	})

	// 3. 执行 handler（不打印请求开始日志）
	resp, err := handler(ctx, req)

	// 4. 记录请求完成（带耗时）
	latency := time.Since(startTime)
	fields := logrus.Fields{
		"method":     shortMethod(info.FullMethod),
		"latency_ms": latency.Milliseconds(),
	}
	if err != nil {
		fields["error"] = err.Error()
		log.WithFields(fields).Error("request failed")
	} else {
		log.WithFields(fields).Info("request completed")
	}

	return resp, err
}

// shortMethod 简化方法名
// /package.v1.ServiceName/MethodName -> ServiceName.MethodName
func shortMethod(m string) string {
	if len(m) == 0 {
		return m
	}
	// 去掉开头的 /
	m = strings.TrimPrefix(m, "/")

	// 找到最后一个 /，取后半部分 (MethodName)
	if idx := strings.LastIndex(m, "/"); idx >= 0 {
		m = m[idx+1:]
	}

	// 去掉包名前缀 (package.v1.)，保留 ServiceName.MethodName
	// 例如: member.v1.MemberService.Register -> MemberService.Register
	parts := strings.Split(m, ".")
	if len(parts) >= 3 {
		// 去掉前两个部分 (package.v1)
		m = strings.Join(parts[2:], ".")
	}

	return m
}

// Start 启动 gRPC 服务器，监听指定端口
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.config.Port))
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(lis)
}

// Stop 优雅停止服务器，等待正在处理的请求完成，或在超时后强制停止
func (s *Server) Stop(ctx context.Context) error {
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()
	select {
	case <-ctx.Done():
		s.grpcServer.Stop()
		return ctx.Err()
	case <-stopped:
		return nil
	}
}
