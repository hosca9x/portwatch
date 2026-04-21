package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

func makeAnomalyReport(kind ports.AnomalyKind, key string, count int) *ports.AnomalyReport {
	return &ports.AnomalyReport{
		Kind:       kind,
		Key:        key,
		Count:      count,
		Window:     30 * time.Second,
		DetectedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestAnomalyAlerter_NilReportNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewAnomalyAlerter(&buf)
	if err := a.Notify(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for nil report, got: %q", buf.String())
	}
}

func TestAnomalyAlerter_FlappingOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewAnomalyAlerter(&buf)
	report := makeAnomalyReport(ports.AnomalyFlapping, "tcp:8080", 5)
	if err := a.Notify(report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "FLAPPING") {
		t.Errorf("expected FLAPPING in output, got: %q", out)
	}
	if !strings.Contains(out, "tcp:8080") {
		t.Errorf("expected key in output, got: %q", out)
	}
	if !strings.Contains(out, "5") {
		t.Errorf("expected count in output, got: %q", out)
	}
}

func TestAnomalyAlerter_BurstOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewAnomalyAlerter(&buf)
	report := makeAnomalyReport(ports.AnomalyBurst, "batch", 12)
	if err := a.Notify(report); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "BURST") {
		t.Errorf("expected BURST in output, got: %q", out)
	}
	if !strings.Contains(out, "12") {
		t.Errorf("expected count in output, got: %q", out)
	}
}

func TestNewAnomalyAlerter_DefaultsToStdout(t *testing.T) {
	a := NewAnomalyAlerter(nil)
	if a.out == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}
