package context

import (
	stlcontext "context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Context 是应用级上下文接口，扩展标准库 context.Context。
// *gin.Context 已实现此接口，可直接作为 Context 使用。
type Context interface {
	stlcontext.Context
}

// CancelFunc 取消函数，与 context.CancelFunc 等价。
type CancelFunc = stlcontext.CancelFunc

// ======================== 标准库 context 包装 ========================

// Background 返回一个非 nil 的空 Context，等同于 context.Background()。
func Background() Context {
	return stlcontext.Background()
}

// TODO 返回一个非 nil 的空 Context，用于尚未确定使用哪个上下文的位置。
func TODO() Context {
	return stlcontext.TODO()
}

// WithCancel 返回一个带取消功能的派生 Context，等同于 context.WithCancel。
func WithCancel(parent Context) (Context, CancelFunc) {
	return stlcontext.WithCancel(parent)
}

// WithDeadline 返回一个带截止时间的派生 Context，等同于 context.WithDeadline。
func WithDeadline(parent Context, deadline time.Time) (Context, CancelFunc) {
	return stlcontext.WithDeadline(parent, deadline)
}

// WithTimeout 返回一个带超时的派生 Context，等同于 context.WithTimeout。
func WithTimeout(parent Context, timeout time.Duration) (Context, CancelFunc) {
	return stlcontext.WithTimeout(parent, timeout)
}

// WithoutCancel 返回一个不受父级取消影响的派生 Context，等同于 context.WithoutCancel。
func WithoutCancel(parent Context) Context {
	return stlcontext.WithoutCancel(parent)
}

// AfterFunc 安排在 ctx 完成后（取消或超时）执行 f，等同于 context.AfterFunc。
// 返回的 stop 函数可在 ctx 完成前取消执行。
func AfterFunc(ctx Context, f func()) (stop func() bool) {
	return stlcontext.AfterFunc(ctx, f)
}

// ======================== user ID ========================

type userIDKey struct{}

// WithUserID 将用户 ID 注入上下文。
func WithUserID(ctx Context, uid string) Context {
	return stlcontext.WithValue(ctx, userIDKey{}, uid)
}

// GetUserID 从上下文获取用户 ID。
func GetUserID(ctx Context) (string, error) {
	v := ctx.Value(userIDKey{})
	if v == nil {
		return "", fmt.Errorf("user id not found in context")
	}
	uid, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("user id in context is not a string")
	}
	return uid, nil
}

// ======================== trace ID ========================

type traceIDKey struct{}

// WithTraceID 将链路追踪 ID 注入上下文，用于日志关联和分布式追踪。
func WithTraceID(ctx Context, traceID string) Context {
	return stlcontext.WithValue(ctx, traceIDKey{}, traceID)
}

// GetTraceID 从上下文获取链路追踪 ID，不存在则返回空字符串。
func GetTraceID(ctx Context) string {
	v := ctx.Value(traceIDKey{})
	if v == nil {
		return ""
	}
	traceID, ok := v.(string)
	if !ok {
		return ""
	}
	return traceID
}

// ======================== request ID ========================

type requestIDKey struct{}

// WithRequestID 将请求 ID 注入上下文。通常对应一次 HTTP 请求。
func WithRequestID(ctx Context, requestID string) Context {
	return stlcontext.WithValue(ctx, requestIDKey{}, requestID)
}

// GetRequestID 从上下文获取请求 ID，不存在则返回空字符串。
func GetRequestID(ctx Context) string {
	v := ctx.Value(requestIDKey{})
	if v == nil {
		return ""
	}
	requestID, ok := v.(string)
	if !ok {
		return ""
	}
	return requestID
}

// ======================== gin adapter ========================

// FromGin 将 *gin.Context 显式转换为 Context。
// 由于 *gin.Context 实现了 context.Context，实际可直接传递，
// 此函数仅用于代码可读性。
func FromGin(c *gin.Context) Context {
	return c
}

// IsGinContext 判断当前 Context 是否由 *gin.Context 适配而来。
func IsGinContext(ctx Context) bool {
	_, ok := ctx.(*gin.Context)
	return ok
}

// ======================== mock ========================

type mockKey struct{}

// WithMockValue 将任意值注入上下文，供测试使用。
func WithMockValue(ctx Context, value any) Context {
	return stlcontext.WithValue(ctx, mockKey{}, value)
}

// GetMockValue 从上下文获取测试值。
func GetMockValue(ctx Context) any {
	return ctx.Value(mockKey{})
}

// IsMockContext 判断当前 Context 是否包含测试值。
func IsMockContext(ctx Context) bool {
	return ctx.Value(mockKey{}) != nil
}
