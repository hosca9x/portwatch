package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

var shadowEpoch = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func makeShadowEntries(n int) []ports.ShadowEntry {
	out := make([]ports.ShadowEntry, n)
	for i := range out {
		out[i] = ports.ShadowEntry{
			Key:       fmt.Sprintf("tcp:%d", 8000+i),
			Port:      8000 + i,
			Proto:     "tcp",
			FirstSeen: shadowEpoch,
			LastSeen:  shadowEpoch.Add(10 * time.Second),
			SeenCount: 1,
		}
	}
	return out
}

func TestShadowAlert_EmptyNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewShadowAlerter(&buf)
	a.Notify(nil)
	if buf.Len() != 0 {
		t.Errorf("expected no output, got %q", buf.String())
	}
}

func TestShadowAlert_SingleEntryEmitted(t *testing.T) {
	var buf bytes.Buffer
	a := NewShadowAlerter(&buf)
	a.Notify(makeShadowEntries(1))
	out := buf.String()
	if !strings.Contains(out, "[SHADOW]") {
		t.Errorf("expected [SHADOW] tag, got %q", out)
	}
	if !strings.Contains(out, "port=8000") {
		t.Errorf("expected port=8000 in output, got %q", out)
	}
}

func TestShadowAlert_MultipleEntriesMultipleLines(t *testing.T) {
	var buf bytes.Buffer
	a := NewShadowAlerter(&buf)
	a.Notify(makeShadowEntries(3))
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestShadowAlert_LineContainsAppearances(t *testing.T) {
	var buf bytes.Buffer
	a := NewShadowAlerter(&buf)
	entries := makeShadowEntries(1)
	entries[0].SeenCount = 2
	a.Notify(entries)
	if !strings.Contains(buf.String(), "appearances=2") {
		t.Errorf("expected appearances=2 in output, got %q", buf.String())
	}
}

func TestNewShadowAlerter_DefaultsToStdout(t *testing.T) {
	a := NewShadowAlerter(nil)
	if a.out == nil {
		t.Error("expected non-nil writer")
	}
}
