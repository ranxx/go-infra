package rate

// Limiter defines the interface for a rate limiter.
type Limiter interface {
	Allow(n int) bool
	Limit() float64
}

// RateLimiter defines the interface for a rate limiter that can provide limiters based on keys.
type RateLimiter interface {
	Limiter(key string) Limiter
}
