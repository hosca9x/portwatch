package ports

import (
	"testing"
	"time"
)

// TestCircuit_FullLifecycle exercises the complete open → half-open → closed
// transition sequence with a controlled clock.
func TestCircuit_FullLifecycle(t *testing.T) {
	now := time.Now()
	clock := fixedCircuitClock(now)
	cfg := CircuitBreakerConfig{
		Threshold:    2,
		RecoveryWait: 5 * time.Second,
	}
	cb := newCircuitBreakerWithClock(cfg, clock)

	// Phase 1: closed — operations allowed.
	if !cb.Allow() {
		t.Fatal("phase 1: expected Allow in closed state")
	}

	// Phase 2: record enough failures to open.
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("phase 2: expected CircuitOpen, got %v", cb.State())
	}
	if cb.Allow() {
		t.Fatal("phase 2: expected block while open")
	}

	// Phase 3: advance clock; circuit should transition to half-open.
	cb.clock = fixedCircuitClock(now.Add(6 * time.Second))
	if !cb.Allow() {
		t.Fatal("phase 3: expected Allow in half-open state")
	}
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("phase 3: expected CircuitHalfOpen, got %v", cb.State())
	}

	// Phase 4: successful probe closes the circuit.
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("phase 4: expected CircuitClosed after success, got %v", cb.State())
	}
	if !cb.Allow() {
		t.Fatal("phase 4: expected Allow after recovery")
	}
}

// TestCircuit_MultipleBreakers ensures independent instances do not share state.
func TestCircuit_MultipleBreakers(t *testing.T) {
	cfg := CircuitBreakerConfig{Threshold: 2, RecoveryWait: 10 * time.Second}
	a := NewCircuitBreaker(cfg)
	b := NewCircuitBreaker(cfg)

	a.RecordFailure()
	a.RecordFailure()

	if b.State() != CircuitClosed {
		t.Fatal("breaker b should remain closed when only a fails")
	}
	if !b.Allow() {
		t.Fatal("breaker b should still allow operations")
	}
}
