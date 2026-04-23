package ports

import (
	"testing"
	"time"
)

var fixedTopologyClock = func() func() time.Time {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return func() time.Time { return now }
}()

func makeTopologyEntries(keys ...string) []PortEntry {
	var out []PortEntry
	for _, k := range keys {
		out = append(out, PortEntry{Port: 80, Proto: k})
	}
	return out
}

func TestTopologyTracker_EmptyOnNoRecord(t *testing.T) {
	tr := newTopologyTrackerWithClock(time.Minute, fixedTopologyClock)
	snap := tr.Snapshot()
	if len(snap.Edges) != 0 {
		t.Fatalf("expected empty edges, got %d", len(snap.Edges))
	}
}

func TestTopologyTracker_SingleEntryNoEdges(t *testing.T) {
	tr := newTopologyTrackerWithClock(time.Minute, fixedTopologyClock)
	tr.Record([]PortEntry{{Port: 80, Proto: "tcp"}})
	snap := tr.Snapshot()
	if len(snap.Edges) != 0 {
		t.Fatalf("single entry should produce no edges, got %v", snap.Edges)
	}
}

func TestTopologyTracker_TwoEntriesOneEdge(t *testing.T) {
	tr := newTopologyTrackerWithClock(time.Minute, fixedTopologyClock)
	tr.Record([]PortEntry{{Port: 80, Proto: "tcp"}, {Port: 443, Proto: "tcp"}})
	snap := tr.Snapshot()
	if len(snap.Edges) != 2 {
		t.Fatalf("expected 2 adjacency entries, got %d", len(snap.Edges))
	}
}

func TestTopologyTracker_EdgesExpireAfterWindow(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	tr := newTopologyTrackerWithClock(time.Minute, clock)
	tr.Record([]PortEntry{{Port: 80, Proto: "tcp"}, {Port: 443, Proto: "tcp"}})

	// advance past window
	now = now.Add(2 * time.Minute)
	snap := tr.Snapshot()
	if len(snap.Edges) != 0 {
		t.Fatalf("expected edges to expire, got %v", snap.Edges)
	}
}

func TestTopologyTracker_DeduplicatesEdgesAcrossBuckets(t *testing.T) {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	tr := newTopologyTrackerWithClock(time.Minute, clock)
	entries := []PortEntry{{Port: 80, Proto: "tcp"}, {Port: 443, Proto: "tcp"}}
	tr.Record(entries)
	now = now.Add(10 * time.Second)
	tr.Record(entries)
	snap := tr.Snapshot()
	// dedup means still only one edge per direction
	for k, v := range snap.Edges {
		if len(v) != 1 {
			t.Fatalf("key %s should have 1 neighbour, got %d", k, len(v))
		}
	}
}

func TestTopologyTracker_SnapshotTimestamp(t *testing.T) {
	now := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	tr := newTopologyTrackerWithClock(time.Minute, clock)
	snap := tr.Snapshot()
	if !snap.Timestamp.Equal(now) {
		t.Fatalf("expected timestamp %v, got %v", now, snap.Timestamp)
	}
}
