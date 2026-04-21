package alert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makePipelineChan(entries []ports.PortEntry) <-chan ports.PortEntry {
	ch := make(chan ports.PortEntry, len(entries))
	for _, e := range entries {
		ch <- e
	}
	close(ch)
	return ch
}

func TestPipelineAlerter_EmitsOneLinePerEntry(t *testing.T) {
	var buf bytes.Buffer
	a := NewPipelineAlerter(&buf)

	entries := []ports.PortEntry{
		{Port: 80, Proto: "tcp"},
		{Port: 443, Proto: "tcp"},
	}
	a.Consume(makePipelineChan(entries))

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 alert lines, got %d: %q", len(lines), buf.String())
	}
}

func TestPipelineAlerter_LineContainsPortAndProto(t *testing.T) {
	var buf bytes.Buffer
	a := NewPipelineAlerter(&buf)

	a.Consume(makePipelineChan([]ports.PortEntry{{Port: 8080, Proto: "udp"}}))

	got := buf.String()
	if !strings.Contains(got, "port=8080") {
		t.Errorf("expected port=8080 in output, got: %q", got)
	}
	if !strings.Contains(got, "proto=udp") {
		t.Errorf("expected proto=udp in output, got: %q", got)
	}
}

func TestPipelineAlerter_EmptyChannelNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewPipelineAlerter(&buf)

	a.Consume(makePipelineChan(nil))

	if buf.Len() != 0 {
		t.Errorf("expected no output for empty channel, got: %q", buf.String())
	}
}

func TestNewPipelineAlerter_DefaultsToStdout(t *testing.T) {
	a := NewPipelineAlerter(nil)
	if a.out == nil {
		t.Fatal("expected non-nil writer when nil passed")
	}
}

func TestPipelineAlerter_LineContainsPipelineAlertPrefix(t *testing.T) {
	var buf bytes.Buffer
	a := NewPipelineAlerter(&buf)

	a.Consume(makePipelineChan([]ports.PortEntry{{Port: 22, Proto: "tcp"}}))

	if !strings.Contains(buf.String(), "[pipeline-alert]") {
		t.Errorf("expected [pipeline-alert] prefix, got: %q", buf.String())
	}
}
