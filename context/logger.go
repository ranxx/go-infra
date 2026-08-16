package context

import (
	stlcontext "context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type loggerKey struct{}

// WithLogger 将 logrus.Logger 注入上下文
func WithLogger(ctx Context, log *logrus.Logger) Context {
	return stlcontext.WithValue(ctx, loggerKey{}, log)
}

// GetLogger 从上下文获取 logrus.Logger；没有则返回 nil。
// 业务代码应该优先用 go-infra/logger.GetLogger() 全局单例，
// 这里仅用于"按请求维度切换 logger"等少数场景。
func GetLogger(ctx Context) *logrus.Logger {
	if gc, ok := ctx.(*gin.Context); ok {
		v, ok := gc.Get("-gin-logger-")
		if ok {
			if log, ok := v.(*logrus.Logger); ok {
				return log
			}
		}
	}
	v := ctx.Value(loggerKey{})
	if v != nil {
		if log, ok := v.(*logrus.Logger); ok {
			return log
		}
	}
	return nil
}

// GinCtxWithLogger 将 logger 注入 gin.Context
func GinCtxWithLogger(ctx *gin.Context, log *logrus.Logger) *gin.Context {
	ctx.Set("-gin-logger-", log)
	return ctx
}