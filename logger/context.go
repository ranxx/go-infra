package logger

import (
	"context"

	"github.com/ranxx/go-infra/tracer"

	"github.com/sirupsen/logrus"
)

// WithContext 创建带 traceID 的 logger
// 使用方式: log := logger.WithContext(ctx, s.logger)
// 之后直接使用 log.Info() 会自动带上 traceID
func WithContext(ctx context.Context, logger logrus.FieldLogger) logrus.FieldLogger {
	traceID := tracer.GetTraceID(ctx)
	if traceID != "" {
		return logger.WithField(tracer.GetTraceFieldName(), traceID)
	}
	return logger
}
