package sources

import (
	"context"
	"sync"
	"time"
)

// RateLimiter provides per-source rate limiting using a token bucket.
type RateLimiter struct {
	interval time.Duration
	mu       sync.Mutex
	last     time.Time
}

// NewRateLimiter creates a rate limiter that allows one request per interval.
func NewRateLimiter(interval time.Duration) *RateLimiter {
	return &RateLimiter{interval: interval}
}

// NewRateLimiterPerSecond creates a rate limiter from a requests-per-second rate.
func NewRateLimiterPerSecond(rps float64) *RateLimiter {
	if rps <= 0 {
		rps = 1
	}
	return NewRateLimiter(time.Duration(float64(time.Second) / rps))
}

// Wait blocks until a request is allowed or the context is cancelled.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	elapsed := time.Since(rl.last)
	if wait := rl.interval - elapsed; wait > 0 {
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	rl.last = time.Now()
	return nil
}

// Interval returns the minimum time between requests.
func (rl *RateLimiter) Interval() time.Duration {
	return rl.interval
}
