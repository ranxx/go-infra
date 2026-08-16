package middleware

import (
	"net/http"

	"github.com/ranxx/go-infra/limiter/rate"
)

// RateLimitOption configures the rate limit middleware.
type RateLimitOption func(*rateLimitConfig)

type rateLimitConfig struct {
	onLimit func(w http.ResponseWriter, r *http.Request)
}

// WithOnLimit sets a custom handler for rate-limited requests.
// By default, http.Error with status 429 is returned.
func WithOnLimit(fn func(w http.ResponseWriter, r *http.Request)) RateLimitOption {
	return func(c *rateLimitConfig) {
		c.onLimit = fn
	}
}

// RateLimit returns an HTTP middleware that rate-limits requests using the
// provided rate.RateLimiter. The keyFunc extracts a key from the request
// (e.g. IP address or user ID). Requests with an empty key are not limited.
//
// Example:
//
//	lim := rate.NewTokenBucket(100, time.Minute)
//	h = RateLimit(lim, func(r *http.Request) string { return r.RemoteAddr })(h)
//
//	// With custom on-limit handler:
//	h = RateLimit(lim, keyFunc, WithOnLimit(func(w, r) {
//	    http.Error(w, `{"code":90001}`, 429)
//	}))(h)
func RateLimit(limiter rate.RateLimiter, keyFunc func(*http.Request) string, opts ...RateLimitOption) func(http.Handler) http.Handler {
	cfg := &rateLimitConfig{
		onLimit: func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		},
	}
	for _, o := range opts {
		o(cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if keyFunc == nil {
				next.ServeHTTP(w, r)
				return
			}
			key := keyFunc(r)
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if !limiter.Limiter(key).Allow(1) {
				cfg.onLimit(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
