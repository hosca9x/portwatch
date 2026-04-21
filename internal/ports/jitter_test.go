package ports

import (
	"testing"
	"time"
)

func TestJitterer_ZeroFractionReturnsBase(t *testing.T) {
	j := NewJitterer(JitterPolicy{MaxFraction: 0})
	base := 10 * time.Second
	got := j.Apply(base)
	if got != base {
		t.Fatalf("expected %v, got %v", base, got)
	}
}

func TestJitterer_NegativeFractionReturnsBase(t *testing.T) {
	j := NewJitterer(JitterPolicy{MaxFraction: -0.5})
	base := 5 * time.Second
	if got := j.Apply(base); got != base {
		t.Fatalf("expected %v, got %v", base, got)
	}
}

func TestJitterer_ZeroBaseReturnsZero(t *testing.T) {
	j := NewJitterer(DefaultJitterPolicy())
	if got := j.Apply(0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestJitterer_ApplyStaysWithinBounds(t *testing.T) {
	policy := JitterPolicy{MaxFraction: 0.25}
	j := NewJitterer(policy)
	j.Reset(42)

	base := 100 * time.Millisecond
	max := base + time.Duration(float64(base)*policy.MaxFraction)

	for i := 0; i < 200; i++ {
		got := j.Apply(base)
		if got < base {
			t.Fatalf("iteration %d: jittered value %v is less than base %v", i, got, base)
		}
		if got >= max+time.Nanosecond {
			t.Fatalf("iteration %d: jittered value %v exceeds max %v", i, got, max)
		}
	}
}

func TestJitterer_DefaultPolicy(t *testing.T) {
	p := DefaultJitterPolicy()
	if p.MaxFraction != 0.20 {
		t.Fatalf("expected MaxFraction 0.20, got %v", p.MaxFraction)
	}
}

func TestJitterer_ResetChangesSeed(t *testing.T) {
	j := NewJitterer(DefaultJitterPolicy())
	base := 1 * time.Second

	j.Reset(1)
	a := j.Apply(base)

	j.Reset(1)
	b := j.Apply(base)

	if a != b {
		t.Fatalf("same seed should produce same value: %v vs %v", a, b)
	}
}
