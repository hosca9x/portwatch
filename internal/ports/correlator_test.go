package ports

import (
	"testing"
	"time"
)

var baseTime = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedCorrelatorClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeCorrelatorEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto, Addr: "127.0.0.1"}
}

func TestCorrelator_FlushNilWhenEmpty(t *testing.T) {
	c := newCorrelatorWithClock(5*time.Second, fixedCorrelatorClock(baseTime))
	if r := c.Flush(); r != nil {
		t.Errorf("expected nil report on empty correlator, got %+v", r)
	}
}

func TestCorrelator_FlushReturnsRecordedEvents(t *testing.T) {
	c := newCorrelatorWithClock(5*time.Second, fixedCorrelatorClock(baseTime))
	c.Record(makeCorrelatorEntry(80, "tcp"), true)
	c.Record(makeCorrelatorEntry(443, "tcp"), true)

	r := c.Flush()
	if r == nil {
		t.Fatal("expected non-nil report")
	}
	if len(r.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(r.Events))
	}
}

func TestCorrelator_StaleEventsExpired(t *testing.T) {
	now := baseTime
	clock := func() time.Time { return now }
	c := newCorrelatorWithClock(5*time.Second, clock)

	// Record an event in the past
	now = baseTime
	c.Record(makeCorrelatorEntry(22, "tcp"), true)

	// Advance time beyond the window
	now = baseTime.Add(10 * time.Second)
	r := c.Flush()
	if r != nil {
		t.Errorf("expected nil after expiry, got %+v", r)
	}
}

func TestCorrelator_MixedFreshAndStale(t *testing.T) {
	now := baseTime
	clock := func() time.Time { return now }
	c := newCorrelatorWithClock(5*time.Second, clock)

	now = baseTime
	c.Record(makeCorrelatorEntry(22, "tcp"), true)

	now = baseTime.Add(10 * time.Second)
	c.Record(makeCorrelatorEntry(80, "tcp"), true)

	r := c.Flush()
	if r == nil {
		t.Fatal("expected non-nil report")
	}
	if len(r.Events) != 1 {
		t.Errorf("expected 1 fresh event, got %d", len(r.Events))
	}
	if r.Events[0].Port != 80 {
		t.Errorf("expected port 80, got %d", r.Events[0].Port)
	}
}

func TestCorrelator_Reset(t *testing.T) {
	c := newCorrelatorWithClock(5*time.Second, fixedCorrelatorClock(baseTime))
	c.Record(makeCorrelatorEntry(8080, "tcp"), false)
	c.Reset()
	if r := c.Flush(); r != nil {
		t.Errorf("expected nil after reset, got %+v", r)
	}
}

func TestCorrelator_ReportTimestamps(t *testing.T) {
	now := baseTime
	clock := func() time.Time { return now }
	c := newCorrelatorWithClock(10*time.Second, clock)

	now = baseTime
	c.Record(makeCorrelatorEntry(80, "tcp"), true)
	now = baseTime.Add(3 * time.Second)
	c.Record(makeCorrelatorEntry(443, "tcp"), true)

	r := c.Flush()
	if r == nil {
		t.Fatal("expected report")
	}
	if !r.StartedAt.Equal(baseTime) {
		t.Errorf("wrong StartedAt: %v", r.StartedAt)
	}
	if !r.EndedAt.Equal(baseTime.Add(3 * time.Second)) {
		t.Errorf("wrong EndedAt: %v", r.EndedAt)
	}
}
