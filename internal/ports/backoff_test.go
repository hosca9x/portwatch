package ports

import (
	"testing"
	"time"
)

func defaultBackoffPolicy() BackoffPolicy {
	return BackoffPolicy{
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Multiplier: 2.0,
	}
}

func TestBackoff_FirstCallReturnsBaseDelay(t *testing.T) {
	b := NewBackoff(defaultBackoffPolicy())
	delay := b.Next("host:22")
	if delay != 100*time.Millisecond {
		t.Errorf("expected 100ms, got %v", delay)
	}
}

func TestBackoff_SecondCallDoubles(t *testing.T) {
	b := NewBackoff(defaultBackoffPolicy())
	b.Next("host:22") // attempt 0 → 100ms
	delay := b.Next("host:22") // attempt 1 → 200ms
	if delay != 200*time.Millisecond {
		t.Errorf("expected 200ms, got %v", delay)
	}
}

func TestBackoff_CapsAtMaxDelay(t *testing.T) {
	b := NewBackoff(defaultBackoffPolicy())
	var last time.Duration
	for i := 0; i < 20; i++ {
		last = b.Next("host:22")
	}
	if last != 1*time.Second {
		t.Errorf("expected max delay 1s, got %v", last)
	}
}

func TestBackoff_ResetClearsAttempts(t *testing.T) {
	b := NewBackoff(defaultBackoffPolicy())
	b.Next("host:22")
	b.Next("host:22")
	b.Reset("host:22")

	if attempts := b.Attempts("host:22"); attempts != 0 {
		t.Errorf("expected 0 attempts after reset, got %d", attempts)
	}
	delay := b.Next("host:22")
	if delay != 100*time.Millisecond {
		t.Errorf("expected base delay after reset, got %v", delay)
	}
}

func TestBackoff_IndependentKeys(t *testing.T) {
	b := NewBackoff(defaultBackoffPolicy())
	b.Next("host:22")
	b.Next("host:22")

	delay := b.Next("host:80") // fresh key
	if delay != 100*time.Millisecond {
		t.Errorf("expected base delay for independent key, got %v", delay)
	}
}

func TestBackoff_AttemptsTracked(t *testing.T) {
	b := NewBackoff(defaultBackoffPolicy())
	b.Next("host:443")
	b.Next("host:443")
	b.Next("host:443")

	if a := b.Attempts("host:443"); a != 3 {
		t.Errorf("expected 3 attempts, got %d", a)
	}
}

func TestDefaultBackoffPolicy_SaneValues(t *testing.T) {
	p := DefaultBackoffPolicy()
	if p.BaseDelay <= 0 {
		t.Error("BaseDelay must be positive")
	}
	if p.MaxDelay <= p.BaseDelay {
		t.Error("MaxDelay must exceed BaseDelay")
	}
	if p.Multiplier <= 1.0 {
		t.Error("Multiplier must be greater than 1")
	}
}
