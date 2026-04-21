package alert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeTagEntry(port int, proto string) ports.PortEntry {
	return ports.PortEntry{Port: port, Proto: proto}
}

func TestTagAlert_MatchedTagEmitsLine(t *testing.T) {
	tm := ports.NewTagMap()
	e := makeTagEntry(80, "tcp")
	tm.Add(e.Key(), "web")

	var buf bytes.Buffer
	a := NewTagAlerter(&buf, tm, "web")
	a.Notify([]ports.PortEntry{e})

	if !strings.Contains(buf.String(), "web") {
		t.Errorf("expected 'web' in output, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "80") {
		t.Errorf("expected port 80 in output, got: %s", buf.String())
	}
}

func TestTagAlert_NoMatchNoOutput(t *testing.T) {
	tm := ports.NewTagMap()
	e := makeTagEntry(443, "tcp")
	tm.Add(e.Key(), "tls")

	var buf bytes.Buffer
	a := NewTagAlerter(&buf, tm, "web") // watching "web", entry has "tls"
	a.Notify([]ports.PortEntry{e})

	if buf.Len() != 0 {
		t.Errorf("expected no output, got: %s", buf.String())
	}
}

func TestTagAlert_UntaggedEntryNoOutput(t *testing.T) {
	tm := ports.NewTagMap()
	e := makeTagEntry(22, "tcp")

	var buf bytes.Buffer
	a := NewTagAlerter(&buf, tm, "ssh")
	a.Notify([]ports.PortEntry{e})

	if buf.Len() != 0 {
		t.Errorf("expected no output for untagged entry, got: %s", buf.String())
	}
}

func TestTagAlert_MultipleTagsReported(t *testing.T) {
	tm := ports.NewTagMap()
	e := makeTagEntry(8080, "tcp")
	tm.Add(e.Key(), "proxy", "web")

	var buf bytes.Buffer
	a := NewTagAlerter(&buf, tm, "proxy", "web")
	a.Notify([]ports.PortEntry{e})

	out := buf.String()
	if !strings.Contains(out, "proxy") || !strings.Contains(out, "web") {
		t.Errorf("expected both tags in output, got: %s", out)
	}
}

func TestNewTagAlerter_DefaultsToStdout(t *testing.T) {
	tm := ports.NewTagMap()
	a := NewTagAlerter(nil, tm, "any")
	if a.w == nil {
		t.Error("expected non-nil writer")
	}
}
