package ports

import (
	"testing"
	"time"
)

func TestBurst_TriggerThenReset(t *testing.T) {
	policy := BurstPolicy{Window: 10 * time.Second, Threshold: 3}
	now := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }

	d := newBurstDetectorWithClock(policy, clock)

	var report *BurstReport
	for i := 0; i < 3; i++ {
		report = d.Record("tcp:3000")
	}
	if report == nil {
		t.Fatal("expected burst report after threshold")
	}
	if report.Count != 3 {
		t.Errorf("want count=3, got %d", report.Count)
	}

	d.Reset("tcp:3000")

	for i := 0; i < 2; i++ {
		if r := d.Record("tcp:3000"); r != nil {
			t.Errorf("expected nil after reset, got report with count=%d", r.Count)
		}
	}
}

func TestBurst_WindowSliding(t *testing.T) {
	policy := BurstPolicy{Window: 5 * time.Second, Threshold: 3}
	times := []time.Time{
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 1, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 10, 0, time.UTC), // outside window for first two
		time.Date(2024, 1, 1, 0, 0, 11, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 12, 0, time.UTC),
	}
	i := 0
	clock := func() time.Time {
		v := times[i]
		if i < len(times)-1 {
			i++
		}
		return v
	}

	d := newBurstDetectorWithClock(policy, clock)

	// First two are outside window by the time we reach t=10s
	d.Record("udp:161")
	d.Record("udp:161")

	// Now three rapid events within 5s window
	var last *BurstReport
	for j := 0; j < 3; j++ {
		last = d.Record("udp:161")
	}
	if last == nil {
		t.Fatal("expected burst report for rapid events")
	}
}
