package ports

import (
	"sync"
	"time"
)

// ThrottleConfig holds configuration for the scan throttler.
type ThrottleConfig struct {
	// MaxScansPerMinute limits how many scans can be triggered per minute.
	MaxScansPerMinute int
	// BurstSize allows a short burst of scans before throttling kicks in.
	BurstSize int
}

// Throttler limits the rate of port scans to avoid overwhelming the host.
type Throttler struct {
	mu        sync.Mutex
	cfg       ThrottleConfig
	tokens    int
	lastRefil time.Time
	clock     func() time.Time
}

// NewThrottler creates a Throttler with the given config.
// If clock is nil, time.Now is used.
func NewThrottler(cfg ThrottleConfig, clock func() time.Time) *Throttler {
	if clock == nil {
		clock = time.Now
	}
	if cfg.BurstSize <= 0 {
		cfg.BurstSize = 1
	}
	if cfg.MaxScansPerMinute <= 0 {
		cfg.MaxScansPerMinute = 60
	}
	return &Throttler{
		cfg:       cfg,
		tokens:    cfg.BurstSize,
		lastRefil: clock(),
		clock:     clock,
	}
}

// Allow returns true if a scan is permitted right now, consuming one token.
// Tokens are refilled based on MaxScansPerMinute over elapsed time.
func (t *Throttler) Allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.clock()
	elapsed := now.Sub(t.lastRefil)

	// Refill tokens proportional to elapsed time.
	if elapsed > 0 {
		rate := float64(t.cfg.MaxScansPerMinute) / 60.0 // tokens per second
		newTokens := int(elapsed.Seconds() * rate)
		if newTokens > 0 {
			t.tokens += newTokens
			if t.tokens > t.cfg.BurstSize {
				t.tokens = t.cfg.BurstSize
			}
			t.lastRefil = now
		}
	}

	if t.tokens <= 0 {
		return false
	}
	t.tokens--
	return true
}

// Reset clears the token bucket back to full burst capacity.
func (t *Throttler) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tokens = t.cfg.BurstSize
	t.lastRefil = t.clock()
}
