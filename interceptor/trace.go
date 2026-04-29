package interceptor

import (
	"context"

	"github.com/ranxx/go-infra/tracer"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// TraceUnaryInterceptor gRPC 客户端拦截器，自动将 traceId 注入到 metadata
func TraceUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any,
		cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 从 context 获取 traceId，如果没有则生成
		traceID := tracer.GetOrNewTraceID(ctx)
		// 注入到 gRPC metadata (自动传递到服务端)
		ctx = metadata.AppendToOutgoingContext(ctx, tracer.GetHeaderName(), traceID)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
