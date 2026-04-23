package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

func makeTopologySnap(edges map[string][]string) ports.TopologySnapshot {
	return ports.TopologySnapshot{
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Edges:     edges,
	}
}

func TestTopologyAlert_EmptyEdgesNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewTopologyAlerter(&buf)
	a.Notify(makeTopologySnap(nil))
	if buf.Len() != 0 {
		t.Fatalf("expected no output, got: %q", buf.String())
	}
}

func TestTopologyAlert_SingleEdgeEmitted(t *testing.T) {
	var buf bytes.Buffer
	a := NewTopologyAlerter(&buf)
	a.Notify(makeTopologySnap(map[string][]string{
		"80/tcp":  {"443/tcp"},
		"443/tcp": {"80/tcp"},
	}))
	out := buf.String()
	if !strings.Contains(out, "co-observed") {
		t.Fatalf("expected co-observed in output, got: %q", out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 unique edge line, got %d: %v", len(lines), lines)
	}
}

func TestTopologyAlert_MultipleEdgesDeduped(t *testing.T) {
	var buf bytes.Buffer
	a := NewTopologyAlerter(&buf)
	a.Notify(makeTopologySnap(map[string][]string{
		"80/tcp":   {"443/tcp", "8080/tcp"},
		"443/tcp":  {"80/tcp"},
		"8080/tcp": {"80/tcp"},
	}))
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 unique edge lines, got %d: %v", len(lines), lines)
	}
}

func TestTopologyAlert_LineContainsTimestamp(t *testing.T) {
	var buf bytes.Buffer
	a := NewTopologyAlerter(&buf)
	a.Notify(makeTopologySnap(map[string][]string{
		"22/tcp": {"80/tcp"},
		"80/tcp": {"22/tcp"},
	}))
	if !strings.Contains(buf.String(), "12:00:00") {
		t.Fatalf("expected timestamp in output, got: %q", buf.String())
	}
}

func TestNewTopologyAlerter_DefaultsToStdout(t *testing.T) {
	a := NewTopologyAlerter(nil)
	if a.out == nil {
		t.Fatal("expected non-nil writer")
	}
}
