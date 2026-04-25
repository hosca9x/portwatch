package ports

import (
	"testing"
	"time"
)

var fixedDigestNow = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedDigestClock(offset time.Duration) func() time.Time {
	return func() time.Time { return fixedDigestNow.Add(offset) }
}

func makeDigestEntries(ports ...int) []PortEntry {
	out := make([]PortEntry, len(ports))
	for i, p := range ports {
		out[i] = PortEntry{Port: p, Proto: "tcp"}
	}
	return out
}

func TestDigestTracker_DeterministicOutput(t *testing.T) {
	dt := newDigestTrackerWithClock(DefaultDigestPolicy(), fixedDigestClock(0))
	a := dt.Compute("scan", makeDigestEntries(80, 443, 22))
	dt.Invalidate("scan")
	b := dt.Compute("scan", makeDigestEntries(443, 22, 80)) // different order
	if a != b {
		t.Fatalf("expected same digest for same port set, got %s vs %s", a, b)
	}
}

func TestDigestTracker_DiffersOnChange(t *testing.T) {
	dt := newDigestTrackerWithClock(DefaultDigestPolicy(), fixedDigestClock(0))
	a := dt.Compute("scan", makeDigestEntries(80, 443))
	dt.Invalidate("scan")
	b := dt.Compute("scan", makeDigestEntries(80, 8080))
	if a == b {
		t.Fatal("expected different digest after port change")
	}
}

func TestDigestTracker_CacheHit(t *testing.T) {
	var calls int
	clock := func() time.Time {
		calls++
		return fixedDigestNow
	}
	dt := newDigestTrackerWithClock(DefaultDigestPolicy(), clock)
	dt.Compute("scan", makeDigestEntries(80))
	d1 := dt.Compute("scan", makeDigestEntries(9999)) // different entries, but cache hit
	dt.Invalidate("scan")
	d2 := dt.Compute("scan", makeDigestEntries(9999))
	if d1 == d2 {
		t.Fatal("expected different digest after invalidation")
	}
}

func TestDigestTracker_ExpiredCacheRecomputed(t *testing.T) {
	offset := 0 * time.Second
	clock := func() time.Time { return fixedDigestNow.Add(offset) }
	policy := DigestPolicy{TTL: 5 * time.Minute}
	dt := newDigestTrackerWithClock(policy, clock)

	d1 := dt.Compute("scan", makeDigestEntries(80))
	offset = 6 * time.Minute // advance past TTL
	d2 := dt.Compute("scan", makeDigestEntries(443))
	if d1 == d2 {
		t.Fatal("expected recomputed digest after TTL expiry")
	}
}

func TestDigestTracker_CachedReturnsFalseWhenMissing(t *testing.T) {
	dt := newDigestTrackerWithClock(DefaultDigestPolicy(), fixedDigestClock(0))
	_, ok := dt.Cached("missing")
	if ok {
		t.Fatal("expected false for missing label")
	}
}

func TestDigestTracker_CachedReturnsTrueAfterCompute(t *testing.T) {
	dt := newDigestTrackerWithClock(DefaultDigestPolicy(), fixedDigestClock(0))
	dt.Compute("scan", makeDigestEntries(80))
	e, ok := dt.Cached("scan")
	if !ok {
		t.Fatal("expected cached entry to be found")
	}
	if e.Digest == "" {
		t.Fatal("expected non-empty digest")
	}
}
