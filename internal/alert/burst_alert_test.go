package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

func makeBurstReport(key string, count int) *ports.BurstReport {
	return &ports.BurstReport{
		Key:        key,
		Count:      count,
		Window:     30 * time.Second,
		DetectedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestBurstAlert_NilReportNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewBurstAlerter(&buf)
	a.Notify(nil)
	if buf.Len() != 0 {
		t.Errorf("expected no output for nil report, got: %q", buf.String())
	}
}

func TestBurstAlert_EmitsLine(t *testing.T) {
	var buf bytes.Buffer
	a := NewBurstAlerter(&buf)
	a.Notify(makeBurstReport("tcp:8080", 7))
	out := buf.String()
	if !strings.Contains(out, "BURST") {
		t.Errorf("expected BURST tag in output: %q", out)
	}
	if !strings.Contains(out, "tcp:8080") {
		t.Errorf("expected key in output: %q", out)
	}
	if !strings.Contains(out, "7") {
		t.Errorf("expected count in output: %q", out)
	}
}

func TestBurstAlert_NotifyAll_MultipleLines(t *testing.T) {
	var buf bytes.Buffer
	a := NewBurstAlerter(&buf)
	a.NotifyAll([]*ports.BurstReport{
		makeBurstReport("tcp:80", 5),
		makeBurstReport("udp:53", 6),
	})
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}
}

func TestBurstAlert_NotifyAll_SkipsNil(t *testing.T) {
	var buf bytes.Buffer
	a := NewBurstAlerter(&buf)
	a.NotifyAll([]*ports.BurstReport{nil, makeBurstReport("tcp:443", 4), nil})
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line, got %d", len(lines))
	}
}

func TestNewBurstAlerter_DefaultsToStdout(t *testing.T) {
	a := NewBurstAlerter(nil)
	if a.out == nil {
		t.Error("expected non-nil writer")
	}
}

func TestBurstAlert_LineContainsWindow(t *testing.T) {
	var buf bytes.Buffer
	a := NewBurstAlerter(&buf)
	a.Notify(makeBurstReport("tcp:22", 3))
	if !strings.Contains(buf.String(), "30s") {
		t.Errorf("expected window in output: %q", buf.String())
	}
}
