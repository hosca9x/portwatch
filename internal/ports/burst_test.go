package ports

import (
	"testing"
	"time"
)

var fixedBurstClock = func() func() time.Time {
	t := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return func() time.Time { return t }
}()

func defaultBurstPolicy() BurstPolicy {
	return BurstPolicy{
		Window:    30 * time.Second,
		Threshold: 3,
	}
}

func TestBurstDetector_BelowThresholdNoReport(t *testing.T) {
	d := newBurstDetectorWithClock(defaultBurstPolicy(), fixedBurstClock)
	for i := 0; i < 2; i++ {
		if r := d.Record("tcp:8080"); r != nil {
			t.Fatalf("expected nil report, got %+v", r)
		}
	}
}

func TestBurstDetector_AtThresholdEmitsReport(t *testing.T) {
	d := newBurstDetectorWithClock(defaultBurstPolicy(), fixedBurstClock)
	var last *BurstReport
	for i := 0; i < 3; i++ {
		last = d.Record("tcp:8080")
	}
	if last == nil {
		t.Fatal("expected burst report, got nil")
	}
	if last.Count != 3 {
		t.Errorf("expected count 3, got %d", last.Count)
	}
	if last.Key != "tcp:8080" {
		t.Errorf("unexpected key: %s", last.Key)
	}
}

func TestBurstDetector_StaleEventsExpire(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	calls := []time.Time{
		now.Add(-60 * time.Second),
		now.Add(-50 * time.Second),
		now, // only this one is within 30s window
	}
	i := 0
	clock := func() time.Time {
		v := calls[i]
		if i < len(calls)-1 {
			i++
		}
		return v
	}
	d := newBurstDetectorWithClock(defaultBurstPolicy(), clock)
	for range calls {
		d.Record("udp:53")
	}
	// Only 1 event within window, threshold=3 → no report
	if r := d.Record("udp:53"); r != nil && r.Count >= 3 {
		t.Errorf("expected no burst, got count %d", r.Count)
	}
}

func TestBurstDetector_IndependentKeys(t *testing.T) {
	d := newBurstDetectorWithClock(defaultBurstPolicy(), fixedBurstClock)
	for i := 0; i < 3; i++ {
		d.Record("tcp:80")
	}
	// Different key should not be affected
	if r := d.Record("tcp:443"); r != nil {
		t.Errorf("expected nil for independent key, got %+v", r)
	}
}

func TestBurstDetector_ResetClearsState(t *testing.T) {
	d := newBurstDetectorWithClock(defaultBurstPolicy(), fixedBurstClock)
	for i := 0; i < 3; i++ {
		d.Record("tcp:9090")
	}
	d.Reset("tcp:9090")
	if r := d.Record("tcp:9090"); r != nil {
		t.Errorf("expected nil after reset, got %+v", r)
	}
}

func TestDefaultBurstPolicy_Values(t *testing.T) {
	p := DefaultBurstPolicy()
	if p.Threshold <= 0 {
		t.Errorf("expected positive threshold, got %d", p.Threshold)
	}
	if p.Window <= 0 {
		t.Errorf("expected positive window, got %v", p.Window)
	}
}
