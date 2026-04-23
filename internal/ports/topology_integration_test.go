package ports

import (
	"testing"
	"time"
)

func TestTopology_RecordThenSnapshot(t *testing.T) {
	now := time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	tr := newTopologyTrackerWithClock(2*time.Minute, clock)

	scan1 := []PortEntry{
		{Port: 22, Proto: "tcp"},
		{Port: 80, Proto: "tcp"},
		{Port: 443, Proto: "tcp"},
	}
	tr.Record(scan1)

	snap := tr.Snapshot()
	if len(snap.Edges) == 0 {
		t.Fatal("expected non-empty topology after recording three ports")
	}
	// Each of 3 ports should have 2 neighbours.
	for k, v := range snap.Edges {
		if len(v) != 2 {
			t.Errorf("port %s expected 2 neighbours, got %d: %v", k, len(v), v)
		}
	}
}

func TestTopology_PartialExpiry(t *testing.T) {
	now := time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	tr := newTopologyTrackerWithClock(time.Minute, clock)

	// First scan: two ports co-observed.
	tr.Record([]PortEntry{{Port: 22, Proto: "tcp"}, {Port: 80, Proto: "tcp"}})

	// Advance past window, record new pair.
	now = now.Add(90 * time.Second)
	tr.Record([]PortEntry{{Port: 443, Proto: "tcp"}, {Port: 8443, Proto: "tcp"}})

	snap := tr.Snapshot()
	// Old pair should be gone; only new pair remains.
	if _, ok := snap.Edges["22/tcp"]; ok {
		t.Error("expected port 22/tcp to be evicted from topology")
	}
	if _, ok := snap.Edges["443/tcp"]; !ok {
		t.Error("expected port 443/tcp to be present in topology")
	}
}

func TestTopology_EmptyRecordIgnored(t *testing.T) {
	now := time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	tr := newTopologyTrackerWithClock(time.Minute, clock)
	tr.Record(nil)
	tr.Record([]PortEntry{{Port: 80, Proto: "tcp"}})
	snap := tr.Snapshot()
	if len(snap.Edges) != 0 {
		t.Fatalf("expected no edges from single-entry records, got %v", snap.Edges)
	}
}
