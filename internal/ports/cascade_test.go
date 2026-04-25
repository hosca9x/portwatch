package ports

import (
	"testing"
	"time"
)

func fixedCascadeClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeCascadeEntry(port int) PortEntry {
	return PortEntry{Port: port, Proto: "tcp"}
}

func TestCascade_BelowThresholdNoReport(t *testing.T) {
	now := time.Now()
	cd := newCascadeDetectorWithClock(CascadePolicy{MinPorts: 3, Window: 10 * time.Second}, fixedCascadeClock(now))

	for i := 0; i < 2; i++ {
		if r := cd.Record(makeCascadeEntry(8000 + i)); r != nil {
			t.Fatalf("expected nil report below threshold, got %+v", r)
		}
	}
}

func TestCascade_AtThresholdEmitsReport(t *testing.T) {
	now := time.Now()
	cd := newCascadeDetectorWithClock(CascadePolicy{MinPorts: 3, Window: 10 * time.Second}, fixedCascadeClock(now))

	var report *CascadeReport
	for i := 0; i < 3; i++ {
		report = cd.Record(makeCascadeEntry(9000 + i))
	}
	if report == nil {
		t.Fatal("expected cascade report at threshold")
	}
	if len(report.Ports) != 3 {
		t.Fatalf("expected 3 ports in report, got %d", len(report.Ports))
	}
}

func TestCascade_StaleEventsExpire(t *testing.T) {
	base := time.Now()
	clock := base
	cd := newCascadeDetectorWithClock(CascadePolicy{MinPorts: 3, Window: 5 * time.Second}, func() time.Time { return clock })

	// record 2 events then advance past window
	cd.Record(makeCascadeEntry(7001))
	cd.Record(makeCascadeEntry(7002))

	clock = base.Add(6 * time.Second)

	// this third event should NOT trigger because the first two are stale
	if r := cd.Record(makeCascadeEntry(7003)); r != nil {
		t.Fatalf("expected nil after stale expiry, got %+v", r)
	}
}

func TestCascade_ResetClearsState(t *testing.T) {
	now := time.Now()
	cd := newCascadeDetectorWithClock(CascadePolicy{MinPorts: 2, Window: 10 * time.Second}, fixedCascadeClock(now))

	cd.Record(makeCascadeEntry(6001))
	cd.Reset()

	if r := cd.Record(makeCascadeEntry(6002)); r != nil {
		t.Fatalf("expected nil after reset, got %+v", r)
	}
}

func TestDefaultCascadePolicy_Values(t *testing.T) {
	p := DefaultCascadePolicy()
	if p.MinPorts <= 0 {
		t.Errorf("MinPorts should be positive, got %d", p.MinPorts)
	}
	if p.Window <= 0 {
		t.Errorf("Window should be positive, got %v", p.Window)
	}
}
