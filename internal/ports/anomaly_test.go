package ports

import (
	"testing"
	"time"
)

func fixedAnomalyClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAnomalyDetector_NoAnomalyBelowThreshold(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d := NewAnomalyDetector(time.Minute, 4, 5)
	d.clock = fixedAnomalyClock(base)

	for i := 0; i < 3; i++ {
		report := d.Record("tcp:8080")
		if report != nil {
			t.Fatalf("expected no anomaly below threshold, got %+v", report)
		}
	}
}

func TestAnomalyDetector_DetectsFlapping(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d := NewAnomalyDetector(time.Minute, 4, 5)
	d.clock = fixedAnomalyClock(base)

	var report *AnomalyReport
	for i := 0; i < 4; i++ {
		report = d.Record("tcp:8080")
	}
	if report == nil {
		t.Fatal("expected flapping anomaly, got nil")
	}
	if report.Kind != AnomalyFlapping {
		t.Errorf("expected kind %q, got %q", AnomalyFlapping, report.Kind)
	}
	if report.Key != "tcp:8080" {
		t.Errorf("unexpected key: %s", report.Key)
	}
}

func TestAnomalyDetector_EventsExpireOutsideWindow(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d := NewAnomalyDetector(30*time.Second, 3, 5)
	d.clock = fixedAnomalyClock(base)

	// Record 2 events at base time
	d.Record("tcp:9090")
	d.Record("tcp:9090")

	// Advance clock past the window
	d.clock = fixedAnomalyClock(base.Add(60 * time.Second))
	// This event should start fresh; only 1 in window
	report := d.Record("tcp:9090")
	if report != nil {
		t.Fatalf("old events should have expired, got anomaly: %+v", report)
	}
}

func TestAnomalyDetector_DetectsBurst(t *testing.T) {
	d := NewAnomalyDetector(time.Minute, 10, 3)
	keys := []string{"tcp:1000", "tcp:1001", "tcp:1002"}
	report := d.RecordBurst(keys)
	if report == nil {
		t.Fatal("expected burst anomaly, got nil")
	}
	if report.Kind != AnomalyBurst {
		t.Errorf("expected kind %q, got %q", AnomalyBurst, report.Kind)
	}
	if report.Count != 3 {
		t.Errorf("expected count 3, got %d", report.Count)
	}
}

func TestAnomalyDetector_BurstBelowThreshold(t *testing.T) {
	d := NewAnomalyDetector(time.Minute, 10, 5)
	report := d.RecordBurst([]string{"tcp:1000", "tcp:1001"})
	if report != nil {
		t.Fatalf("expected no burst anomaly, got %+v", report)
	}
}

func TestAnomalyDetector_Reset(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	d := NewAnomalyDetector(time.Minute, 3, 5)
	d.clock = fixedAnomalyClock(base)

	d.Record("tcp:7070")
	d.Record("tcp:7070")
	d.Reset("tcp:7070")

	// After reset, two more records should not trigger threshold
	d.Record("tcp:7070")
	report := d.Record("tcp:7070")
	if report != nil {
		t.Fatalf("expected no anomaly after reset, got %+v", report)
	}
}
