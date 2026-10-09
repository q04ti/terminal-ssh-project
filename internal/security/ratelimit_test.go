package security

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(3, 100*time.Millisecond)

	key := "client_test_1"

	// First 3 should succeed (capacity is 3)
	if !rl.Allow(key) {
		t.Fatal("1st request should be allowed")
	}
	if !rl.Allow(key) {
		t.Fatal("2nd request should be allowed")
	}
	if !rl.Allow(key) {
		t.Fatal("3rd request should be allowed")
	}

	// 4th request exceeds burst limit
	if rl.Allow(key) {
		t.Fatal("4th request should be blocked by rate limit")
	}

	// Wait for replenishment
	time.Sleep(150 * time.Millisecond)
	if !rl.Allow(key) {
		t.Fatal("Request after replenishment should be allowed")
	}
}
