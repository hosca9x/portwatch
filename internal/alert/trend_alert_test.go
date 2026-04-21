package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

func makeTrendReport(dir ports.TrendDirection, delta int, n int) ports.TrendReport {
	samples := make([]ports.TrendSample, n)
	for i := range samples {
		samples[i] = ports.TrendSample{At: time.Now(), Count: 10 + i}
	}
	return ports.TrendReport{Direction: dir, Delta: delta, Samples: samples}
}

func TestTrendAlerter_StableNoOutput(t *testing.T) {
	var buf bytes.Buffer
	ta := NewTrendAlerter(&buf)

	emitted := ta.Notify(makeTrendReport(ports.TrendStable, 0, 3))
	if emitted {
		t.Fatal("expected no alert for stable trend")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected empty output, got: %s", buf.String())
	}
}

func TestTrendAlerter_UpEmitsAlert(t *testing.T) {
	var buf bytes.Buffer
	ta := NewTrendAlerter(&buf)
	ta.clock = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }

	emitted := ta.Notify(makeTrendReport(ports.TrendUp, 4, 5))
	if !emitted {
		t.Fatal("expected alert for upward trend")
	}
	out := buf.String()
	if !strings.Contains(out, "TREND up") {
		t.Fatalf("expected 'TREND up' in output, got: %s", out)
	}
	if !strings.Contains(out, "+4") {
		t.Fatalf("expected '+4' delta in output, got: %s", out)
	}
}

func TestTrendAlerter_DownEmitsAlert(t *testing.T) {
	var buf bytes.Buffer
	ta := NewTrendAlerter(&buf)
	ta.clock = func() time.Time { return time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) }

	emitted := ta.Notify(makeTrendReport(ports.TrendDown, -7, 4))
	if !emitted {
		t.Fatal("expected alert for downward trend")
	}
	out := buf.String()
	if !strings.Contains(out, "TREND down") {
		t.Fatalf("expected 'TREND down' in output, got: %s", out)
	}
	if !strings.Contains(out, "-7") {
		t.Fatalf("expected '-7' delta in output, got: %s", out)
	}
}

func TestNewTrendAlerter_DefaultsToStdout(t *testing.T) {
	ta := NewTrendAlerter(nil)
	if ta.out == nil {
		t.Fatal("expected non-nil writer when nil passed")
	}
}
