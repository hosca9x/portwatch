package ports

import (
	"sync"
	"time"
)

// RateLimiter prevents alert flooding by tracking how recently
// a given port key has triggered a notification.
type RateLimiter struct {
	mu       sync.Mutex
	lastSeen map[string]time.Time
	cooldown time.Duration
}

// NewRateLimiter creates a RateLimiter with the given cooldown duration.
// Alerts for the same port key are suppressed until the cooldown has elapsed.
func NewRateLimiter(cooldown time.Duration) *RateLimiter {
	return &RateLimiter{
		lastSeen: make(map[string]time.Time),
		cooldown: cooldown,
	}
}

// Allow returns true if the given key has not been seen within the cooldown
// window. If allowed, the key's timestamp is updated.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	if last, ok := r.lastSeen[key]; ok {
		if now.Sub(last) < r.cooldown {
			return false
		}
	}
	r.lastSeen[key] = now
	return true
}

// Reset clears the rate-limit state for a specific key.
func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lastSeen, key)
}

// Flush clears all tracked keys.
func (r *RateLimiter) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastSeen = make(map[string]time.Time)
}
