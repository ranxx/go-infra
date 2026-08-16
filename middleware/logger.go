package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ranxx/go-infra/logger"
)

// Printfer is the minimal logging interface used by middlewares.
// It captures only the Printf method actually needed.
// The standard library's *log.Logger satisfies this interface.
type Printfer interface {
	Printf(format string, v ...any)
}

// LoggerOption configures the Logger middleware.
type LoggerOption func(*loggerConfig)

type loggerConfig struct {
	logger        Printfer
	traceIDHeader string
}

// WithLoggerOutput sets the logger for the Logger middleware.
// Defaults to the standard library's default logger.
func WithLoggerOutput(l Printfer) LoggerOption {
	return func(c *loggerConfig) {
		c.logger = l
	}
}

// WithTraceIDHeader sets the header name used for trace ID propagation.
// Defaults to "X-Trace-ID".
func WithTraceIDHeader(header string) LoggerOption {
	return func(c *loggerConfig) {
		c.traceIDHeader = header
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logger returns an HTTP middleware that logs each request with method, path,
// status code, latency, and a trace ID.
func Logger(opts ...LoggerOption) func(http.Handler) http.Handler {
	cfg := &loggerConfig{
		logger:        logger.GetLogger(),
		traceIDHeader: "X-Trace-ID",
	}
	for _, o := range opts {
		o(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			traceID := r.Header.Get(cfg.traceIDHeader)
			if traceID == "" {
				traceID = uuid.New().String()
			}
			r.Header.Set(cfg.traceIDHeader, traceID)

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			cfg.logger.Printf("[%s] %s %s %d %v",
				traceID[:8], r.Method, r.URL.Path, rw.statusCode, time.Since(start))
		})
	}
}

// RecoveryOption configures the Recovery middleware.
type RecoveryOption func(*recoveryConfig)

type recoveryConfig struct {
	logger  Printfer
	handler func(w http.ResponseWriter, r *http.Request, err any)
}

// WithRecoveryLogger sets the logger for the Recovery middleware.
// Defaults to the standard library's default logger.
func WithRecoveryLogger(l Printfer) RecoveryOption {
	return func(c *recoveryConfig) {
		c.logger = l
	}
}

// WithRecoveryHandler sets a custom panic handler for the Recovery middleware.
// By default, http.Error with status 500 is used.
func WithRecoveryHandler(fn func(w http.ResponseWriter, r *http.Request, err any)) RecoveryOption {
	return func(c *recoveryConfig) {
		c.handler = fn
	}
}

// Recovery returns an HTTP middleware that recovers from panics, logs them,
// and returns a 500 error to the client.
func Recovery(opts ...RecoveryOption) func(http.Handler) http.Handler {
	cfg := &recoveryConfig{
		logger: log.Default(),
	}
	for _, o := range opts {
		o(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					cfg.logger.Printf("PANIC recovered: %v", err)
					if cfg.handler != nil {
						cfg.handler(w, r, err)
					} else {
						http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
