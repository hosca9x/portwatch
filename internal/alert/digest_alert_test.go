package alert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeDigestAlerter(w *bytes.Buffer) *DigestAlert {
	return NewDigestAlerter(w, ports.DefaultDigestPolicy())
}

func makeDigestPortEntries(ps ...int) []ports.PortEntry {
	out := make([]ports.PortEntry, len(ps))
	for i, p := range ps {
		out[i] = ports.PortEntry{Port: p, Proto: "tcp"}
	}
	return out
}

func TestDigestAlert_NoOutputWhenUnchanged(t *testing.T) {
	var buf bytes.Buffer
	a := makeDigestAlerter(&buf)
	a.Notify("scan", makeDigestPortEntries(80, 443))
	buf.Reset()
	a.Notify("scan", makeDigestPortEntries(80, 443))
	if buf.Len() != 0 {
		t.Fatalf("expected no output for unchanged digest, got: %s", buf.String())
	}
}

func TestDigestAlert_EmitsOnChange(t *testing.T) {
	var buf bytes.Buffer
	a := makeDigestAlerter(&buf)
	a.Notify("scan", makeDigestPortEntries(80))
	buf.Reset()
	a.Notify("scan", makeDigestPortEntries(8080))
	if buf.Len() == 0 {
		t.Fatal("expected output after digest change")
	}
	line := buf.String()
	if !strings.Contains(line, "digest changed") {
		t.Fatalf("expected 'digest changed' in output, got: %s", line)
	}
}

func TestDigestAlert_LineContainsLabel(t *testing.T) {
	var buf bytes.Buffer
	a := makeDigestAlerter(&buf)
	a.Notify("myscan", makeDigestPortEntries(80))
	if !strings.Contains(buf.String(), "label=myscan") {
		t.Fatalf("expected label in output, got: %s", buf.String())
	}
}

func TestDigestAlert_FirstCallAlwaysEmits(t *testing.T) {
	var buf bytes.Buffer
	a := makeDigestAlerter(&buf)
	a.Notify("scan", makeDigestPortEntries(22))
	if buf.Len() == 0 {
		t.Fatal("expected output on first call")
	}
}

func TestDigestAlert_ResetCausesReemit(t *testing.T) {
	var buf bytes.Buffer
	a := makeDigestAlerter(&buf)
	a.Notify("scan", makeDigestPortEntries(80))
	buf.Reset()
	a.Notify("scan", makeDigestPortEntries(80)) // same — no output
	if buf.Len() != 0 {
		t.Fatal("expected no output before reset")
	}
	a.Reset()
	a.Notify("scan", makeDigestPortEntries(80))
	if buf.Len() == 0 {
		t.Fatal("expected output after reset")
	}
}

func TestNewDigestAlerter_DefaultsToStdout(t *testing.T) {
	a := NewDigestAlerter(nil, ports.DefaultDigestPolicy())
	if a.w == nil {
		t.Fatal("expected non-nil writer")
	}
}
