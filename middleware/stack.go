package middleware

import (
	"net/http"

	"github.com/ranxx/go-infra/limiter/rate"
)

// Stack composes multiple HTTP middlewares into a single chain.
// It provides a declarative way to build middleware pipelines.
//
// Example:
//
//	stack := NewStack(
//	    WithRecovery(WithRecoveryHandler(myHandler)),
//	    WithLogger(),
//	    WithCORS(WithAllowedOrigins("*")),
//	    WithRateLimit(limiter, keyFunc),
//	)
//	h := stack.Then(mux)
type Stack struct {
	middlewares []func(http.Handler) http.Handler
}

// StackOption adds a middleware to the stack.
type StackOption func(*Stack)

// NewStack creates a middleware stack from the given options.
// Middlewares are applied in the order they are added (first added = outermost).
func NewStack(opts ...StackOption) *Stack {
	s := &Stack{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Then applies the middleware chain to the given handler and returns the
// wrapped handler. Middlewares are applied in the order they were added
// to the stack: the first added wraps the outermost layer.
func (s *Stack) Then(h http.Handler) http.Handler {
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		h = s.middlewares[i](h)
	}
	return h
}

// ThenFunc is a convenience wrapper around Then for http.HandlerFunc.
func (s *Stack) ThenFunc(h http.HandlerFunc) http.Handler {
	return s.Then(h)
}

// ---- StackOption constructors ----

// WithCORS adds a CORS middleware to the stack.
func WithCORS(opts ...CORSOption) StackOption {
	return func(s *Stack) {
		s.middlewares = append(s.middlewares, CORS(opts...))
	}
}

// WithLogger adds a request logging middleware to the stack.
func WithLogger(opts ...LoggerOption) StackOption {
	return func(s *Stack) {
		s.middlewares = append(s.middlewares, Logger(opts...))
	}
}

// WithRecovery adds a panic recovery middleware to the stack.
func WithRecovery(opts ...RecoveryOption) StackOption {
	return func(s *Stack) {
		s.middlewares = append(s.middlewares, Recovery(opts...))
	}
}

// WithRateLimit adds a rate limiting middleware to the stack.
// limiter is the underlying rate limiter implementation.
// keyFunc extracts a key from the request; empty-key requests are not limited.
func WithRateLimit(limiter rate.RateLimiter, keyFunc func(*http.Request) string, opts ...RateLimitOption) StackOption {
	return func(s *Stack) {
		s.middlewares = append(s.middlewares, RateLimit(limiter, keyFunc, opts...))
	}
}

// WithMiddleware adds an arbitrary middleware to the stack.
// Use this for third-party or project-specific middlewares that need to be
// placed at a specific position in the chain.
//
// Example:
//
//	stack := NewStack(
//	    WithRateLimit(limiter, keyFunc),
//	    WithMiddleware(jwtAuth.Middleware()),
//	    WithLogger(),
//	)
func WithMiddleware(mw func(http.Handler) http.Handler) StackOption {
	return func(s *Stack) {
		s.middlewares = append(s.middlewares, mw)
	}
}
