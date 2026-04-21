package alert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makePriorityAlerter(w *bytes.Buffer, minSev ports.Severity) *PriorityAlerter {
	p := ports.NewPrioritizer([]int{443}, []int{80})
	return NewPriorityAlerter(w, p, minSev)
}

func TestPriorityAlert_HighPortEmitted(t *testing.T) {
	var buf bytes.Buffer
	a := makePriorityAlerter(&buf, ports.SeverityLow)
	a.Notify([]ports.PortEntry{{Port: 443, Proto: "tcp", Addr: "0.0.0.0"}}, nil)
	out := buf.String()
	if !strings.Contains(out, "[HIGH]") {
		t.Errorf("expected [HIGH] in output, got: %s", out)
	}
	if !strings.Contains(out, "OPENED") {
		t.Errorf("expected OPENED in output, got: %s", out)
	}
}

func TestPriorityAlert_FiltersBelowMinSeverity(t *testing.T) {
	var buf bytes.Buffer
	a := makePriorityAlerter(&buf, ports.SeverityHigh)
	a.Notify([]ports.PortEntry{{Port: 9000, Proto: "tcp", Addr: "0.0.0.0"}}, nil)
	if buf.Len() != 0 {
		t.Errorf("expected no output for low severity port, got: %s", buf.String())
	}
}

func TestPriorityAlert_ClosedPortEmitted(t *testing.T) {
	var buf bytes.Buffer
	a := makePriorityAlerter(&buf, ports.SeverityLow)
	a.Notify(nil, []ports.PortEntry{{Port: 80, Proto: "tcp", Addr: "127.0.0.1"}})
	out := buf.String()
	if !strings.Contains(out, "CLOSED") {
		t.Errorf("expected CLOSED in output, got: %s", out)
	}
	if !strings.Contains(out, "[MEDIUM]") {
		t.Errorf("expected [MEDIUM] in output, got: %s", out)
	}
}

func TestPriorityAlert_NoEntriesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := makePriorityAlerter(&buf, ports.SeverityLow)
	a.Notify(nil, nil)
	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %s", buf.String())
	}
}

func TestNewPriorityAlerter_DefaultsToStdout(t *testing.T) {
	p := ports.NewPrioritizer(nil, nil)
	a := NewPriorityAlerter(nil, p, ports.SeverityLow)
	if a.w == nil {
		t.Error("expected non-nil writer")
	}
}
