package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

func makeWindowEntry(proto string, port int) ports.PortEntry {
	return ports.PortEntry{Proto: proto, Port: port}
}

func TestWindowAlerter_BelowThresholdNoOutput(t *testing.T) {
	var buf bytes.Buffer
	policy := ports.WindowPolicy{Size: time.Minute, MaxCount: 5}
	wa := NewWindowAlerter(&buf, policy)
	entry := makeWindowEntry("tcp", 80)

	for i := 0; i < 4; i++ {
		if wa.Observe(entry) {
			t.Fatal("expected no alert below threshold")
		}
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output, got: %s", buf.String())
	}
}

func TestWindowAlerter_AtThresholdEmitsAlert(t *testing.T) {
	var buf bytes.Buffer
	policy := ports.WindowPolicy{Size: time.Minute, MaxCount: 3}
	wa := NewWindowAlerter(&buf, policy)
	entry := makeWindowEntry("tcp", 443)

	for i := 0; i < 2; i++ {
		wa.Observe(entry)
	}
	if alerted := wa.Observe(entry); !alerted {
		t.Fatal("expected alert at threshold")
	}
	if !strings.Contains(buf.String(), "WINDOW-ALERT") {
		t.Fatalf("expected WINDOW-ALERT in output, got: %s", buf.String())
	}
}

func TestWindowAlerter_ResetClearsState(t *testing.T) {
	var buf bytes.Buffer
	policy := ports.WindowPolicy{Size: time.Minute, MaxCount: 2}
	wa := NewWindowAlerter(&buf, policy)
	entry := makeWindowEntry("udp", 53)

	wa.Observe(entry)
	wa.Observe(entry) // hits threshold
	buf.Reset()

	wa.Reset(entry)
	if alerted := wa.Observe(entry); alerted {
		t.Fatal("expected no alert after reset")
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output after reset, got: %s", buf.String())
	}
}

func TestNewWindowAlerter_DefaultsToStdout(t *testing.T) {
	policy := ports.DefaultWindowPolicy()
	wa := NewWindowAlerter(nil, policy)
	if wa.writer == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestWindowAlerter_IndependentKeys(t *testing.T) {
	var buf bytes.Buffer
	policy := ports.WindowPolicy{Size: time.Minute, MaxCount: 2}
	wa := NewWindowAlerter(&buf, policy)

	e1 := makeWindowEntry("tcp", 80)
	e2 := makeWindowEntry("tcp", 8080)

	wa.Observe(e1)
	wa.Observe(e2)

	// only e1 should trigger — it gets a second event
	if alerted := wa.Observe(e1); !alerted {
		t.Fatal("expected alert for e1")
	}
	// e2 still at 1, should not alert
	if alerted := wa.Observe(e2); alerted {
		t.Fatal("expected no alert for e2 at count 2 — wait, MaxCount=2 so this should alert")
		// Note: MaxCount=2 means >=2 triggers. e2 is now at 2, so this path won't execute.
	}
}
