package ports_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/ports"
)

// TestAnomalyDetector_FlapThenAlert verifies the full pipeline:
// detector records events, and the alerter emits output when flapping is detected.
func TestAnomalyDetector_FlapThenAlert(t *testing.T) {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	detector := ports.NewAnomalyDetector(time.Minute, 3, 10)
	detector.SetClock(func() time.Time { return base })

	var buf bytes.Buffer
	alerter := alert.NewAnomalyAlerter(&buf)

	key := "tcp:443"
	var lastReport *ports.AnomalyReport
	for i := 0; i < 3; i++ {
		lastReport = detector.Record(key)
	}

	if lastReport == nil {
		t.Fatal("expected anomaly report after threshold reached")
	}

	if err := alerter.Notify(lastReport); err != nil {
		t.Fatalf("alerter.Notify returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected alert output, got none")
	}
}

// TestAnomalyDetector_BurstThenAlert verifies burst detection feeds into alerter.
func TestAnomalyDetector_BurstThenAlert(t *testing.T) {
	detector := ports.NewAnomalyDetector(time.Minute, 10, 4)

	var buf bytes.Buffer
	alerter := alert.NewAnomalyAlerter(&buf)

	keys := []string{"tcp:3000", "tcp:3001", "tcp:3002", "tcp:3003"}
	report := detector.RecordBurst(keys)
	if report == nil {
		t.Fatal("expected burst report")
	}

	if err := alerter.Notify(report); err != nil {
		t.Fatalf("alerter.Notify returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected burst alert output, got none")
	}
}
