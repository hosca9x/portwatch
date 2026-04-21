package ports

import (
	"testing"
	"time"
)

func fixedQuotaClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestQuotaTracker_BelowLimitNotExceeded(t *testing.T) {
	base := time.Now()
	q := newQuotaTrackerWithClock(QuotaPolicy{MaxEvents: 3, Window: time.Minute}, fixedQuotaClock(base))

	for i := 0; i < 3; i++ {
		if q.Record("tcp:80") {
			t.Fatalf("expected quota not exceeded on call %d", i+1)
		}
	}
}

func TestQuotaTracker_ExceedsOnNextRecord(t *testing.T) {
	base := time.Now()
	q := newQuotaTrackerWithClock(QuotaPolicy{MaxEvents: 3, Window: time.Minute}, fixedQuotaClock(base))

	for i := 0; i < 3; i++ {
		q.Record("tcp:80")
	}
	if !q.Record("tcp:80") {
		t.Fatal("expected quota exceeded on 4th call")
	}
}

func TestQuotaTracker_OldEventsExpire(t *testing.T) {
	now := time.Now()
	var current time.Time
	clockFn := func() time.Time { return current }

	q := newQuotaTrackerWithClock(QuotaPolicy{MaxEvents: 2, Window: time.Minute}, clockFn)

	current = now
	q.Record("tcp:443")
	q.Record("tcp:443")

	// Advance past the window so old events expire.
	current = now.Add(2 * time.Minute)
	if q.Record("tcp:443") {
		t.Fatal("expected quota not exceeded after old events expired")
	}
}

func TestQuotaTracker_IndependentKeys(t *testing.T) {
	base := time.Now()
	q := newQuotaTrackerWithClock(QuotaPolicy{MaxEvents: 1, Window: time.Minute}, fixedQuotaClock(base))

	q.Record("tcp:80")
	if q.Record("tcp:80") == false {
		t.Fatal("expected tcp:80 to be exceeded")
	}
	if q.Record("tcp:443") {
		t.Fatal("expected tcp:443 to be independent and not exceeded")
	}
}

func TestQuotaTracker_ResetClearsCount(t *testing.T) {
	base := time.Now()
	q := newQuotaTrackerWithClock(QuotaPolicy{MaxEvents: 1, Window: time.Minute}, fixedQuotaClock(base))

	q.Record("tcp:22")
	q.Reset("tcp:22")

	if q.Count("tcp:22") != 0 {
		t.Fatalf("expected count 0 after reset, got %d", q.Count("tcp:22"))
	}
	if q.Record("tcp:22") {
		t.Fatal("expected quota not exceeded after reset")
	}
}

func TestQuotaTracker_CountReflectsWindow(t *testing.T) {
	now := time.Now()
	var current time.Time
	clockFn := func() time.Time { return current }

	q := newQuotaTrackerWithClock(QuotaPolicy{MaxEvents: 10, Window: time.Minute}, clockFn)

	current = now
	q.Record("udp:53")
	q.Record("udp:53")

	current = now.Add(2 * time.Minute)
	if got := q.Count("udp:53"); got != 0 {
		t.Fatalf("expected 0 after window expiry, got %d", got)
	}
}
