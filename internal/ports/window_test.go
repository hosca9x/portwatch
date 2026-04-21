package ports

import (
	"testing"
	"time"
)

var fixedWindowBase = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedWindowClock(offset time.Duration) func() time.Time {
	return func() time.Time { return fixedWindowBase.Add(offset) }
}

func TestWindowCounter_InitialCountIsZero(t *testing.T) {
	w := newWindowCounterWithClock(DefaultWindowPolicy(), fixedWindowClock(0))
	if got := w.Count("tcp:80"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestWindowCounter_AddIncrements(t *testing.T) {
	w := newWindowCounterWithClock(DefaultWindowPolicy(), fixedWindowClock(0))
	w.Add("tcp:80")
	w.Add("tcp:80")
	if got := w.Count("tcp:80"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestWindowCounter_ExceedReportsTrue(t *testing.T) {
	policy := WindowPolicy{Size: time.Minute, MaxCount: 3}
	w := newWindowCounterWithClock(policy, fixedWindowClock(0))
	w.Add("tcp:443")
	w.Add("tcp:443")
	w.Add("tcp:443")
	if !w.Exceeded("tcp:443") {
		t.Fatal("expected Exceeded to be true")
	}
}

func TestWindowCounter_OldEventsExpire(t *testing.T) {
	policy := WindowPolicy{Size: 30 * time.Second, MaxCount: 10}
	now := fixedWindowBase
	calls := 0
	clock := func() time.Time {
		calls++
		if calls <= 2 {
			return now
		}
		return now.Add(60 * time.Second)
	}
	w := newWindowCounterWithClock(policy, clock)
	w.Add("tcp:22")
	w.Add("tcp:22")
	// third call is 60s later — previous two should be evicted
	if got := w.Add("tcp:22"); got != 1 {
		t.Fatalf("expected 1 after expiry, got %d", got)
	}
}

func TestWindowCounter_Reset(t *testing.T) {
	w := newWindowCounterWithClock(DefaultWindowPolicy(), fixedWindowClock(0))
	w.Add("tcp:8080")
	w.Add("tcp:8080")
	w.Reset("tcp:8080")
	if got := w.Count("tcp:8080"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestWindowCounter_IndependentKeys(t *testing.T) {
	w := newWindowCounterWithClock(DefaultWindowPolicy(), fixedWindowClock(0))
	w.Add("tcp:80")
	w.Add("tcp:80")
	w.Add("udp:53")
	if got := w.Count("tcp:80"); got != 2 {
		t.Fatalf("tcp:80 expected 2, got %d", got)
	}
	if got := w.Count("udp:53"); got != 1 {
		t.Fatalf("udp:53 expected 1, got %d", got)
	}
}
