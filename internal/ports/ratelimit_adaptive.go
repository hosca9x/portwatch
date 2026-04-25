package ports

import (
	"sync"
	"time"
)

// AdaptiveRateLimiterConfig holds tuning parameters for the adaptive limiter.
type AdaptiveRateLimiterConfig struct {
	BaseRate     float64       // events per second at baseline
	MinRate      float64       // floor rate under heavy load
	MaxRate      float64       // ceiling rate under light load
	AdjustPeriod time.Duration // how often to recalculate rate
}

// DefaultAdaptiveRateLimiterConfig returns sensible defaults.
func DefaultAdaptiveRateLimiterConfig() AdaptiveRateLimiterConfig {
	return AdaptiveRateLimiterConfig{
		BaseRate:     10.0,
		MinRate:      1.0,
		MaxRate:      50.0,
		AdjustPeriod: 10 * time.Second,
	}
}

// AdaptiveRateLimiter adjusts its rate limit based on observed error pressure.
type AdaptiveRateLimiter struct {
	mu          sync.Mutex
	cfg         AdaptiveRateLimiterConfig
	currentRate float64
	errorCount  int
	allowCount  int
	lastAdjust  time.Time
	clock       func() time.Time
}

// NewAdaptiveRateLimiter creates a new AdaptiveRateLimiter with the given config.
func NewAdaptiveRateLimiter(cfg AdaptiveRateLimiterConfig) *AdaptiveRateLimiter {
	return newAdaptiveRateLimiterWithClock(cfg, time.Now)
}

func newAdaptiveRateLimiterWithClock(cfg AdaptiveRateLimiterConfig, clock func() time.Time) *AdaptiveRateLimiter {
	return &AdaptiveRateLimiter{
		cfg:         cfg,
		currentRate: cfg.BaseRate,
		lastAdjust:  clock(),
		clock:       clock,
	}
}

// Allow returns true if the event should be allowed under the current rate.
func (a *AdaptiveRateLimiter) Allow() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.maybeAdjust()
	a.allowCount++
	threshold := int(a.currentRate)
	if threshold < 1 {
		threshold = 1
	}
	return a.allowCount%threshold == 1
}

// RecordError signals that a downstream error occurred, reducing the rate.
func (a *AdaptiveRateLimiter) RecordError() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.errorCount++
}

// CurrentRate returns the current effective rate.
func (a *AdaptiveRateLimiter) CurrentRate() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentRate
}

func (a *AdaptiveRateLimiter) maybeAdjust() {
	now := a.clock()
	if now.Sub(a.lastAdjust) < a.cfg.AdjustPeriod {
		return
	}
	if a.errorCount > 0 {
		a.currentRate = a.currentRate * 0.75
	} else {
		a.currentRate = a.currentRate * 1.1
	}
	if a.currentRate < a.cfg.MinRate {
		a.currentRate = a.cfg.MinRate
	}
	if a.currentRate > a.cfg.MaxRate {
		a.currentRate = a.cfg.MaxRate
	}
	a.errorCount = 0
	a.lastAdjust = now
}
