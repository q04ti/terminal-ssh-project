package security

import (
	"sync"
	"time"
)

// RateLimiter manages rate limits per client/session key.
type RateLimiter struct {
	mu       sync.Mutex
	limits   map[string]*clientBucket
	capacity int
	interval time.Duration
}

type clientBucket struct {
	tokens    int
	lastCheck time.Time
}

// NewRateLimiter creates a token bucket rate limiter.
// capacity is the burst capacity, interval is how often a token replenishes.
func NewRateLimiter(capacity int, interval time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limits:   make(map[string]*clientBucket),
		capacity: capacity,
		interval: interval,
	}

	// Periodically cleanup stale buckets
	go rl.cleanupRoutine()
	return rl
}

// Allow returns true if an action is allowed for the given key, and consumes a token.
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.limits[key]
	if !exists {
		rl.limits[key] = &clientBucket{
			tokens:    rl.capacity - 1,
			lastCheck: now,
		}
		return true
	}

	// Calculate tokens replenished
	elapsed := now.Sub(bucket.lastCheck)
	replenish := int(elapsed / rl.interval)
	if replenish > 0 {
		bucket.tokens += replenish
		if bucket.tokens > rl.capacity {
			bucket.tokens = rl.capacity
		}
		bucket.lastCheck = now
	}

	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

// Reset clears the bucket for a key.
func (rl *RateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.limits, key)
}

func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for k, b := range rl.limits {
			if now.Sub(b.lastCheck) > 10*time.Minute {
				delete(rl.limits, k)
			}
		}
		rl.mu.Unlock()
	}
}
