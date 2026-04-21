package ports

import (
	"math"
	"sync"
	"time"
)

// BackoffPolicy defines the configuration for exponential backoff.
type BackoffPolicy struct {
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
}

// DefaultBackoffPolicy returns a sensible default exponential backoff policy.
func DefaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   30 * time.Second,
		Multiplier: 2.0,
	}
}

// Backoff tracks per-key retry attempts and computes the next wait duration
// using exponential backoff with an optional jitter-free cap.
type Backoff struct {
	policy  BackoffPolicy
	attempt map[string]int
	mu      sync.Mutex
	clock   func() time.Time
}

// NewBackoff creates a Backoff using the provided policy.
func NewBackoff(policy BackoffPolicy) *Backoff {
	return &Backoff{
		policy:  policy,
		attempt: make(map[string]int),
		clock:   time.Now,
	}
}

// Next returns the delay for the current attempt and increments the counter.
func (b *Backoff) Next(key string) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := b.attempt[key]
	b.attempt[key] = n + 1

	delay := float64(b.policy.BaseDelay) * math.Pow(b.policy.Multiplier, float64(n))
	if delay > float64(b.policy.MaxDelay) {
		delay = float64(b.policy.MaxDelay)
	}
	return time.Duration(delay)
}

// Attempts returns the number of consecutive failures recorded for a key.
func (b *Backoff) Attempts(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.attempt[key]
}

// Reset clears the attempt counter for a key (e.g. on success).
func (b *Backoff) Reset(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.attempt, key)
}
