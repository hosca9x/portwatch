package ports

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota // normal operation
	CircuitOpen                        // failing, requests blocked
	CircuitHalfOpen                    // probing for recovery
)

// CircuitBreaker prevents repeated scan attempts when a target is
// consistently failing, backing off until a recovery probe succeeds.
type CircuitBreaker struct {
	mu           sync.Mutex
	state        CircuitState
	failures     int
	threshold    int
	recoveryWait time.Duration
	openedAt     time.Time
	clock        func() time.Time
}

// CircuitBreakerConfig holds tuning parameters for the breaker.
type CircuitBreakerConfig struct {
	Threshold    int           // consecutive failures before opening
	RecoveryWait time.Duration // how long to stay open before half-open probe
}

// NewCircuitBreaker creates a CircuitBreaker with the given config.
func NewCircuitBreaker(cfg CircuitBreakerConfig) *CircuitBreaker {
	return newCircuitBreakerWithClock(cfg, time.Now)
}

func newCircuitBreakerWithClock(cfg CircuitBreakerConfig, clock func() time.Time) *CircuitBreaker {
	if cfg.Threshold <= 0 {
		cfg.Threshold = 3
	}
	if cfg.RecoveryWait <= 0 {
		cfg.RecoveryWait = 30 * time.Second
	}
	return &CircuitBreaker{
		state:        CircuitClosed,
		threshold:    cfg.Threshold,
		recoveryWait: cfg.RecoveryWait,
		clock:        clock,
	}
}

// Allow returns true if the caller is permitted to attempt an operation.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if cb.clock().Sub(cb.openedAt) >= cb.recoveryWait {
			cb.state = CircuitHalfOpen
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return false
}

// RecordSuccess resets the breaker to closed state.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = CircuitClosed
}

// RecordFailure increments the failure count and may open the circuit.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	if cb.state == CircuitHalfOpen || cb.failures >= cb.threshold {
		cb.state = CircuitOpen
		cb.openedAt = cb.clock()
	}
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset forces the breaker back to closed state.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = CircuitClosed
}
