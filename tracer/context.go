package tracer

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type traceIDKey struct{}

// HeaderName gRPC metadata 中的 traceId header 名称
// 可通过 SetHeaderName 自定义
var HeaderName = "x-trace-id"

// TraceFieldName ES 文档中 traceId 字段名，默认为 "trace_id"
// 可通过 SetTraceFieldName 自定义
var TraceFieldName = "trace_id"

var headerNameMu sync.RWMutex
var traceFieldMu sync.RWMutex

// SetHeaderName 设置自定义的 header 名称（线程安全）
func SetHeaderName(name string) {
	headerNameMu.Lock()
	defer headerNameMu.Unlock()
	HeaderName = name
}

// GetHeaderName 获取当前的 header 名称（线程安全）
func GetHeaderName() string {
	headerNameMu.RLock()
	defer headerNameMu.RUnlock()
	return HeaderName
}

// SetTraceFieldName 设置 ES 文档中的 traceId 字段名（线程安全）
func SetTraceFieldName(name string) {
	traceFieldMu.Lock()
	defer traceFieldMu.Unlock()
	TraceFieldName = name
}

// GetTraceFieldName 获取当前 ES 文档中的 traceId 字段名
func GetTraceFieldName() string {
	traceFieldMu.RLock()
	defer traceFieldMu.RUnlock()
	return TraceFieldName
}

// WithTraceID 设置 traceId 到 context
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, traceID)
}

// GetTraceID 从 context 获取 traceId
func GetTraceID(ctx context.Context) string {
	if val := ctx.Value(traceIDKey{}); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// GetOrNewTraceID 获取或生成 traceId
func GetOrNewTraceID(ctx context.Context) string {
	if id := GetTraceID(ctx); id != "" {
		return id
	}
	return uuid.New().String()
}

// GenerateTraceID 生成新的 traceId
func GenerateTraceID() string {
	return uuid.New().String()
}
