package ports

import (
	"testing"
	"time"
)

func fixedCircuitClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func defaultCircuitCfg() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Threshold:    3,
		RecoveryWait: 10 * time.Second,
	}
}

func TestCircuit_InitiallyAllows(t *testing.T) {
	cb := NewCircuitBreaker(defaultCircuitCfg())
	if !cb.Allow() {
		t.Fatal("expected Allow() == true in closed state")
	}
}

func TestCircuit_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(defaultCircuitCfg())
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen, got %v", cb.State())
	}
	if cb.Allow() {
		t.Fatal("expected Allow() == false when circuit is open")
	}
}

func TestCircuit_HalfOpenAfterRecoveryWait(t *testing.T) {
	now := time.Now()
	cb := newCircuitBreakerWithClock(defaultCircuitCfg(), fixedCircuitClock(now))
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	// advance clock past recovery wait
	cb.clock = fixedCircuitClock(now.Add(15 * time.Second))
	if !cb.Allow() {
		t.Fatal("expected Allow() == true in half-open state")
	}
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected CircuitHalfOpen, got %v", cb.State())
	}
}

func TestCircuit_SuccessCloses(t *testing.T) {
	cb := NewCircuitBreaker(defaultCircuitCfg())
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected CircuitClosed after success, got %v", cb.State())
	}
	if !cb.Allow() {
		t.Fatal("expected Allow() == true after recovery")
	}
}

func TestCircuit_HalfOpenFailureReopens(t *testing.T) {
	now := time.Now()
	cb := newCircuitBreakerWithClock(defaultCircuitCfg(), fixedCircuitClock(now))
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	cb.clock = fixedCircuitClock(now.Add(15 * time.Second))
	cb.Allow() // transitions to half-open
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatalf("expected CircuitOpen after half-open failure, got %v", cb.State())
	}
}

func TestCircuit_Reset(t *testing.T) {
	cb := NewCircuitBreaker(defaultCircuitCfg())
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	cb.Reset()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected CircuitClosed after Reset, got %v", cb.State())
	}
}

func TestCircuit_DefaultThreshold(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{})
	for i := 0; i < 3; i++ {
		cb.RecordFailure()
	}
	if cb.State() != CircuitOpen {
		t.Fatal("expected default threshold of 3 to open circuit")
	}
}
