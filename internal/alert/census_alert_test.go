package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/portwatch/internal/ports"
)

func makeCensusSnap(entries []ports.CensusEntry) ports.CensusSnapshot {
	return ports.CensusSnapshot{
		TakenAt: time.Unix(1_700_000_000, 0).UTC(),
		Entries: entries,
	}
}

func TestCensusAlert_EmptyNoOutput(t *testing.T) {
	var buf bytes.Buffer
	a := NewCensusAlerter(&buf, 1)
	a.Notify(makeCensusSnap(nil))
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty snapshot, got: %q", buf.String())
	}
}

func TestCensusAlert_SingleEntryEmitted(t *testing.T) {
	var buf bytes.Buffer
	a := NewCensusAlerter(&buf, 1)
	now := time.Unix(1_700_000_000, 0).UTC()
	snap := makeCensusSnap([]ports.CensusEntry{
		{Key: "80/tcp", FirstSeen: now, LastSeen: now, Count: 3},
	})
	a.Notify(snap)
	out := buf.String()
	if !strings.Contains(out, "80/tcp") {
		t.Errorf("expected key in output, got: %q", out)
	}
	if !strings.Contains(out, "seen=3") {
		t.Errorf("expected count in output, got: %q", out)
	}
}

func TestCensusAlert_BelowMinCountFiltered(t *testing.T) {
	var buf bytes.Buffer
	a := NewCensusAlerter(&buf, 5)
	now := time.Unix(1_700_000_000, 0).UTC()
	snap := makeCensusSnap([]ports.CensusEntry{
		{Key: "22/tcp", FirstSeen: now, LastSeen: now, Count: 2},
	})
	a.Notify(snap)
	if buf.Len() != 0 {
		t.Errorf("expected no output below minCount, got: %q", buf.String())
	}
}

func TestCensusAlert_MultipleEntriesSorted(t *testing.T) {
	var buf bytes.Buffer
	a := NewCensusAlerter(&buf, 1)
	now := time.Unix(1_700_000_000, 0).UTC()
	snap := makeCensusSnap([]ports.CensusEntry{
		{Key: "443/tcp", FirstSeen: now, LastSeen: now, Count: 1},
		{Key: "22/tcp", FirstSeen: now, LastSeen: now, Count: 1},
		{Key: "80/tcp", FirstSeen: now, LastSeen: now, Count: 1},
	})
	a.Notify(snap)
	out := buf.String()
	pos22 := strings.Index(out, "22/tcp")
	pos80 := strings.Index(out, "80/tcp")
	pos443 := strings.Index(out, "443/tcp")
	if pos22 > pos80 || pos80 > pos443 {
		t.Errorf("entries not sorted: %q", out)
	}
}

func TestNewCensusAlerter_DefaultsToStdout(t *testing.T) {
	a := NewCensusAlerter(nil, 0)
	if a.out == nil {
		t.Error("expected non-nil writer")
	}
	if a.minCount < 1 {
		t.Errorf("expected minCount >= 1, got %d", a.minCount)
	}
}
