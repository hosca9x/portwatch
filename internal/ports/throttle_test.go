package ports

import (
	"testing"
	"time"
)

func fixedThrottleClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestThrottler_AllowsUpToBurst(t *testing.T) {
	now := time.Now()
	th := NewThrottler(ThrottleConfig{MaxScansPerMinute: 60, BurstSize: 3}, fixedThrottleClock(now))

	for i := 0; i < 3; i++ {
		if !th.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
	if th.Allow() {
		t.Fatal("expected Allow()=false after burst exhausted")
	}
}

func TestThrottler_RefillsOverTime(t *testing.T) {
	base := time.Now()
	current := base
	clock := func() time.Time { return current }

	th := NewThrottler(ThrottleConfig{MaxScansPerMinute: 60, BurstSize: 1}, clock)

	if !th.Allow() {
		t.Fatal("expected first Allow()=true")
	}
	if th.Allow() {
		t.Fatal("expected Allow()=false after token consumed")
	}

	// Advance time by 2 seconds — should refill 2 tokens, capped at burst=1.
	current = base.Add(2 * time.Second)
	if !th.Allow() {
		t.Fatal("expected Allow()=true after time advance")
	}
}

func TestThrottler_Reset(t *testing.T) {
	now := time.Now()
	th := NewThrottler(ThrottleConfig{MaxScansPerMinute: 60, BurstSize: 2}, fixedThrottleClock(now))

	th.Allow()
	th.Allow()
	if th.Allow() {
		t.Fatal("expected throttled after burst")
	}

	th.Reset()
	if !th.Allow() {
		t.Fatal("expected Allow()=true after Reset")
	}
}

func TestThrottler_DefaultsApplied(t *testing.T) {
	// Zero-value config should not panic and apply defaults.
	th := NewThrottler(ThrottleConfig{}, nil)
	if th.cfg.MaxScansPerMinute != 60 {
		t.Errorf("expected default MaxScansPerMinute=60, got %d", th.cfg.MaxScansPerMinute)
	}
	if th.cfg.BurstSize != 1 {
		t.Errorf("expected default BurstSize=1, got %d", th.cfg.BurstSize)
	}
	if !th.Allow() {
		t.Fatal("expected first Allow()=true with defaults")
	}
}

func TestThrottler_IndependentInstances(t *testing.T) {
	now := time.Now()
	cfg := ThrottleConfig{MaxScansPerMinute: 60, BurstSize: 1}
	th1 := NewThrottler(cfg, fixedThrottleClock(now))
	th2 := NewThrottler(cfg, fixedThrottleClock(now))

	th1.Allow()
	// th2 should still have its token.
	if !th2.Allow() {
		t.Fatal("expected th2 to be independent from th1")
	}
}
