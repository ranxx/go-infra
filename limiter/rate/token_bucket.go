package rate

import (
	"sync"
	"time"
)

// tokenBucket implements Limiter for a single key.
type tokenBucket struct {
	parent     *TokenBucketRateLimiter
	key        string
	tokens     int
	lastRefill time.Time
}

// Allow reports whether n events may happen at this moment.
func (b *tokenBucket) Allow(n int) bool {
	b.parent.mu.Lock()
	defer b.parent.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	refill := int(float64(b.parent.rate) * float64(elapsed) / float64(b.parent.window))
	if refill > 0 {
		b.tokens = min(b.parent.rate, b.tokens+refill)
		b.lastRefill = now
	}

	if b.tokens >= n {
		b.tokens -= n
		return true
	}

	return false
}

// Limit returns the maximum sustained rate (tokens per second).
func (b *tokenBucket) Limit() float64 {
	return float64(b.parent.rate) / b.parent.window.Seconds()
}

// TokenBucketRateLimiter implements RateLimiter using the token bucket algorithm.
// It is safe for concurrent use.
type TokenBucketRateLimiter struct {
	mu              sync.Mutex
	buckets         map[string]*tokenBucket
	rate            int
	window          time.Duration
	cleanupInterval time.Duration
}

// TokenBucketOption configures a TokenBucketRateLimiter.
type TokenBucketOption func(*TokenBucketRateLimiter)

// WithCleanupInterval sets the interval at which stale buckets are removed.
func WithCleanupInterval(d time.Duration) TokenBucketOption {
	return func(rl *TokenBucketRateLimiter) {
		rl.cleanupInterval = d
	}
}

// NewTokenBucket creates a new token bucket rate limiter.
// rate is the maximum number of tokens per window.
func NewTokenBucket(rate int, window time.Duration, opts ...TokenBucketOption) *TokenBucketRateLimiter {
	rl := &TokenBucketRateLimiter{
		buckets:         make(map[string]*tokenBucket),
		rate:            rate,
		window:          window,
		cleanupInterval: 10 * time.Minute,
	}
	for _, o := range opts {
		o(rl)
	}

	go rl.cleanup()
	return rl
}

// Limiter returns a Limiter for the given key. It always returns a valid
// limiter; if no bucket exists yet a new one is created.
func (rl *TokenBucketRateLimiter) Limiter(key string) Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &tokenBucket{
			parent:     rl,
			key:        key,
			tokens:     rl.rate,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = b
	}
	return b
}

func (rl *TokenBucketRateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastRefill) > rl.window*2 {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}
