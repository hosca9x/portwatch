package ports

import (
	"testing"
	"time"
)

func fixedLeakClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeLeakEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto}
}

func TestLeakDetector_BelowThresholdNoReport(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := LeakDetectorConfig{Threshold: 24 * time.Hour}
	d := newLeakDetectorWithClock(cfg, fixedLeakClock(now))

	entries := []PortEntry{makeLeakEntry(8080, "tcp")}
	reports := d.Observe(entries)
	if len(reports) != 0 {
		t.Fatalf("expected no reports, got %d", len(reports))
	}
}

func TestLeakDetector_AtThresholdEmitsReport(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := LeakDetectorConfig{Threshold: 1 * time.Hour}

	var now time.Time
	clock := func() time.Time { return now }
	d := newLeakDetectorWithClock(cfg, clock)

	now = start
	entries := []PortEntry{makeLeakEntry(443, "tcp")}
	d.Observe(entries)

	now = start.Add(2 * time.Hour)
	reports := d.Observe(entries)
	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}
	if reports[0].Port != 443 {
		t.Errorf("expected port 443, got %d", reports[0].Port)
	}
	if reports[0].Duration < cfg.Threshold {
		t.Errorf("duration %v should be >= threshold %v", reports[0].Duration, cfg.Threshold)
	}
}

func TestLeakDetector_ClosedPortEvicted(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := LeakDetectorConfig{Threshold: 1 * time.Minute}
	d := newLeakDetectorWithClock(cfg, fixedLeakClock(now))

	entries := []PortEntry{makeLeakEntry(9090, "tcp")}
	d.Observe(entries)

	// Next observation without that port.
	reports := d.Observe([]PortEntry{})
	if len(reports) != 0 {
		t.Errorf("expected no reports after port closed, got %d", len(reports))
	}
	if len(d.seen) != 0 {
		t.Errorf("expected seen map to be empty after eviction")
	}
}

func TestLeakDetector_ResetClearsState(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := LeakDetectorConfig{Threshold: 1 * time.Minute}
	d := newLeakDetectorWithClock(cfg, fixedLeakClock(now))

	entries := []PortEntry{makeLeakEntry(22, "tcp"), makeLeakEntry(80, "tcp")}
	d.Observe(entries)
	d.Reset()

	if len(d.seen) != 0 {
		t.Errorf("expected empty seen map after Reset, got %d entries", len(d.seen))
	}
}

func TestLeakDetector_MultiplePortsIndependent(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cfg := LeakDetectorConfig{Threshold: 1 * time.Hour}

	var now time.Time
	clock := func() time.Time { return now }
	d := newLeakDetectorWithClock(cfg, clock)

	now = start
	d.Observe([]PortEntry{makeLeakEntry(80, "tcp")})

	now = start.Add(30 * time.Minute)
	d.Observe([]PortEntry{makeLeakEntry(80, "tcp"), makeLeakEntry(443, "tcp")})

	now = start.Add(90 * time.Minute)
	reports := d.Observe([]PortEntry{makeLeakEntry(80, "tcp"), makeLeakEntry(443, "tcp")})

	if len(reports) != 1 {
		t.Fatalf("expected 1 report (only port 80), got %d", len(reports))
	}
	if reports[0].Port != 80 {
		t.Errorf("expected port 80 in report, got %d", reports[0].Port)
	}
}
