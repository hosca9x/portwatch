package alert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeEntry(port int, proto string, pid int) ports.PortEntry {
	return ports.PortEntry{Port: port, Proto: proto, PID: pid}
}

func TestNotify_OpenedPorts(t *testing.T) {
	var buf bytes.Buffer
	n := NewNotifier(&buf)

	d := ports.DiffResult{
		Opened: []ports.PortEntry{makeEntry(8080, "tcp", 1234)},
	}
	n.Notify(d)

	out := buf.String()
	if !strings.Contains(out, "ALERT") {
		t.Errorf("expected ALERT level, got: %s", out)
	}
	if !strings.Contains(out, "8080/tcp") {
		t.Errorf("expected port info in output, got: %s", out)
	}
}

func TestNotify_ClosedPorts(t *testing.T) {
	var buf bytes.Buffer
	n := NewNotifier(&buf)

	d := ports.DiffResult{
		Closed: []ports.PortEntry{makeEntry(443, "tcp", 99)},
	}
	n.Notify(d)

	out := buf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN level, got: %s", out)
	}
	if !strings.Contains(out, "443/tcp") {
		t.Errorf("expected port info in output, got: %s", out)
	}
}

func TestNotify_NoChanges(t *testing.T) {
	var buf bytes.Buffer
	n := NewNotifier(&buf)

	n.Notify(ports.DiffResult{})

	if buf.Len() != 0 {
		t.Errorf("expected no output for empty diff, got: %s", buf.String())
	}
}

func TestNewNotifier_DefaultsToStdout(t *testing.T) {
	n := NewNotifier(nil)
	if n.out == nil {
		t.Error("expected non-nil writer when passed")
	}
}
