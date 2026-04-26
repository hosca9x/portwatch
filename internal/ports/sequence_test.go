package ports

import (
	"testing"
	"time"
)

func fixedSeqClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSequenceTracker_BelowMinLengthNoReport(t *testing.T) {
	p := DefaultSequencePolicy()
	p.MinLength = 3
	tr := newSequenceTrackerWithClock(p, fixedSeqClock(time.Now()))

	if r := tr.Record("host1", 80); r != nil {
		t.Fatalf("expected nil on first record, got %+v", r)
	}
	if r := tr.Record("host1", 81); r != nil {
		t.Fatalf("expected nil on second record, got %+v", r)
	}
}

func TestSequenceTracker_AtMinLengthEmitsReport(t *testing.T) {
	p := DefaultSequencePolicy()
	p.MinLength = 3
	now := time.Now()
	tr := newSequenceTrackerWithClock(p, fixedSeqClock(now))

	tr.Record("host1", 22)
	tr.Record("host1", 80)
	report := tr.Record("host1", 443)

	if report == nil {
		t.Fatal("expected report on third record")
	}
	if report.Length != 3 {
		t.Errorf("expected length 3, got %d", report.Length)
	}
	if report.Key != "host1" {
		t.Errorf("expected key host1, got %s", report.Key)
	}
	if len(report.Ports) != 3 {
		t.Errorf("expected 3 ports, got %d", len(report.Ports))
	}
}

func TestSequenceTracker_ResetsAfterReport(t *testing.T) {
	p := DefaultSequencePolicy()
	p.MinLength = 2
	tr := newSequenceTrackerWithClock(p, fixedSeqClock(time.Now()))

	tr.Record("h", 1)
	tr.Record("h", 2) // triggers report

	// next single record should not trigger
	if r := tr.Record("h", 3); r != nil {
		t.Fatalf("expected nil after reset, got %+v", r)
	}
}

func TestSequenceTracker_GapResetsState(t *testing.T) {
	p := DefaultSequencePolicy()
	p.MinLength = 2
	p.MaxGap = 5 * time.Second

	base := time.Now()
	calls := []time.Time{base, base.Add(10 * time.Second)}
	idx := 0
	clock := func() time.Time {
		t := calls[idx]
		if idx < len(calls)-1 {
			idx++
		}
		return t
	}

	tr := newSequenceTrackerWithClock(p, clock)
	tr.Record("h", 80)
	// 10s gap exceeds MaxGap of 5s — should reset
	if r := tr.Record("h", 443); r != nil {
		t.Fatalf("expected nil after gap reset, got %+v", r)
	}
}

func TestSequenceTracker_IndependentKeys(t *testing.T) {
	p := DefaultSequencePolicy()
	p.MinLength = 2
	tr := newSequenceTrackerWithClock(p, fixedSeqClock(time.Now()))

	tr.Record("a", 1)
	tr.Record("b", 1)

	// neither key should have reached MinLength=2 independently via the other
	if r := tr.Record("a", 2); r == nil {
		t.Fatal("expected report for key a")
	}
	if r := tr.Record("b", 2); r == nil {
		t.Fatal("expected report for key b")
	}
}

func TestSequenceTracker_Reset(t *testing.T) {
	p := DefaultSequencePolicy()
	p.MinLength = 2
	tr := newSequenceTrackerWithClock(p, fixedSeqClock(time.Now()))

	tr.Record("x", 10)
	tr.Reset("x")

	if r := tr.Record("x", 20); r != nil {
		t.Fatalf("expected nil after explicit reset, got %+v", r)
	}
}
