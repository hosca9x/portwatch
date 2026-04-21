package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

func makeCircuitEvent(state ports.CircuitState, failures int) CircuitEvent {
	return CircuitEvent{
		Target:    "192.168.1.1:22",
		State:     state,
		Failures:  failures,
		Timestamp: time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
	}
}

func TestCircuitAlert_OpenEmitted(t *testing.T) {
	var buf bytes.Buffer
	ca := NewCircuitAlerter(&buf)
	ca.Notify(makeCircuitEvent(ports.CircuitOpen, 3))
	out := buf.String()
	if !strings.Contains(out, "CIRCUIT OPEN") {
		t.Errorf("expected CIRCUIT OPEN in output, got: %s", out)
	}
	if !strings.Contains(out, "failures=3") {
		t.Errorf("expected failures=3 in output, got: %s", out)
	}
}

func TestCircuitAlert_HalfOpenEmitted(t *testing.T) {
	var buf bytes.Buffer
	ca := NewCircuitAlerter(&buf)
	ca.Notify(makeCircuitEvent(ports.CircuitHalfOpen, 0))
	out := buf.String()
	if !strings.Contains(out, "CIRCUIT PROBE") {
		t.Errorf("expected CIRCUIT PROBE in output, got: %s", out)
	}
}

func TestCircuitAlert_ClosedEmitted(t *testing.T) {
	var buf bytes.Buffer
	ca := NewCircuitAlerter(&buf)
	ca.Notify(makeCircuitEvent(ports.CircuitClosed, 0))
	out := buf.String()
	if !strings.Contains(out, "CIRCUIT OK") {
		t.Errorf("expected CIRCUIT OK in output, got: %s", out)
	}
}

func TestNewCircuitAlerter_DefaultsToStdout(t *testing.T) {
	ca := NewCircuitAlerter(nil)
	if ca.out == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestCircuitAlert_TargetInOutput(t *testing.T) {
	var buf bytes.Buffer
	ca := NewCircuitAlerter(&buf)
	ca.Notify(makeCircuitEvent(ports.CircuitOpen, 5))
	if !strings.Contains(buf.String(), "192.168.1.1:22") {
		t.Errorf("expected target in output, got: %s", buf.String())
	}
}
