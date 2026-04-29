package ports

import (
	"testing"
	"time"
)

func fixedCensusClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeCensusEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto}
}

func TestCensus_RecordIncreasesCount(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	ct := newCensusTrackerWithClock(DefaultCensusPolicy(), fixedCensusClock(now))
	e := makeCensusEntry(80, "tcp")
	ct.Record(e)
	ct.Record(e)
	snap := ct.Snapshot()
	if len(snap.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap.Entries))
	}
	if snap.Entries[0].Count != 2 {
		t.Errorf("expected count 2, got %d", snap.Entries[0].Count)
	}
}

func TestCensus_FirstAndLastSeenSet(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	var tick int64
	clock := func() time.Time {
		t := base.Add(time.Duration(tick) * time.Second)
		tick++
		return t
	}
	policy := DefaultCensusPolicy()
	ct := newCensusTrackerWithClock(policy, clock)
	e := makeCensusEntry(443, "tcp")
	ct.Record(e)
	ct.Record(e)
	snap := ct.Snapshot()
	if len(snap.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap.Entries))
	}
	ce := snap.Entries[0]
	if !ce.FirstSeen.Before(ce.LastSeen) {
		t.Errorf("expected FirstSeen < LastSeen, got first=%v last=%v", ce.FirstSeen, ce.LastSeen)
	}
}

func TestCensus_StaleEntryEvicted(t *testing.T) {
	base := time.Unix(1_000_000, 0)
	policy := CensusPolicy{MaxAge: 5 * time.Minute}
	ct := newCensusTrackerWithClock(policy, fixedCensusClock(base))
	ct.Record(makeCensusEntry(22, "tcp"))
	// advance clock beyond MaxAge
	ct.clock = fixedCensusClock(base.Add(10 * time.Minute))
	snap := ct.Snapshot()
	if len(snap.Entries) != 0 {
		t.Errorf("expected 0 entries after expiry, got %d", len(snap.Entries))
	}
}

func TestCensus_IndependentKeys(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	ct := newCensusTrackerWithClock(DefaultCensusPolicy(), fixedCensusClock(now))
	ct.Record(makeCensusEntry(80, "tcp"))
	ct.Record(makeCensusEntry(443, "tcp"))
	ct.Record(makeCensusEntry(80, "tcp"))
	snap := ct.Snapshot()
	if len(snap.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap.Entries))
	}
}

func TestCensus_SnapshotTimestamp(t *testing.T) {
	now := time.Unix(9_999_999, 0)
	ct := newCensusTrackerWithClock(DefaultCensusPolicy(), fixedCensusClock(now))
	snap := ct.Snapshot()
	if !snap.TakenAt.Equal(now) {
		t.Errorf("expected TakenAt=%v, got %v", now, snap.TakenAt)
	}
}
