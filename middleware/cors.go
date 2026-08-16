package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOption configures the CORS middleware.
type CORSOption func(*corsConfig)

type corsConfig struct {
	allowedOrigins   []string
	allowedMethods   []string
	allowedHeaders   []string
	allowCredentials bool
	maxAge           int
}

// WithAllowedOrigins sets the allowed origins for CORS.
// Use "*" to allow all origins.
func WithAllowedOrigins(origins ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowedOrigins = origins
	}
}

// WithAllowedMethods sets the allowed HTTP methods for CORS.
func WithAllowedMethods(methods ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowedMethods = methods
	}
}

// WithAllowedHeaders sets the allowed request headers for CORS.
func WithAllowedHeaders(headers ...string) CORSOption {
	return func(c *corsConfig) {
		c.allowedHeaders = headers
	}
}

// WithAllowCredentials sets whether credentials are allowed in CORS requests.
func WithAllowCredentials(v bool) CORSOption {
	return func(c *corsConfig) {
		c.allowCredentials = v
	}
}

// WithMaxAge sets the Access-Control-Max-Age header (in seconds).
func WithMaxAge(seconds int) CORSOption {
	return func(c *corsConfig) {
		c.maxAge = seconds
	}
}

// CORS returns an HTTP middleware that handles cross-origin requests.
func CORS(opts ...CORSOption) func(http.Handler) http.Handler {
	cfg := &corsConfig{}
	for _, o := range opts {
		o(cfg)
	}

	allowedOrigins := make(map[string]bool)
	for _, o := range cfg.allowedOrigins {
		allowedOrigins[strings.ToLower(o)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.ToLower(r.Header.Get("Origin"))

			if _, ok := allowedOrigins[origin]; ok || len(cfg.allowedOrigins) == 1 && cfg.allowedOrigins[0] == "*" {
				if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			if cfg.allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if r.Method == http.MethodOptions {
				if len(cfg.allowedMethods) > 0 {
					w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.allowedMethods, ", "))
				}
				if len(cfg.allowedHeaders) > 0 {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.allowedHeaders, ", "))
				}
				if cfg.maxAge > 0 {
					w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.maxAge))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			if len(cfg.allowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.allowedMethods, ", "))
			}

			next.ServeHTTP(w, r)
		})
	}
}
