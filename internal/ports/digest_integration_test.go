package ports

import (
	"testing"
	"time"
)

func TestDigest_StableAcrossScans(t *testing.T) {
	dt := NewDigestTracker(DefaultDigestPolicy())

	scan1 := []PortEntry{
		{Port: 22, Proto: "tcp"},
		{Port: 80, Proto: "tcp"},
		{Port: 443, Proto: "tcp"},
	}
	scan2 := []PortEntry{
		{Port: 443, Proto: "tcp"},
		{Port: 22, Proto: "tcp"},
		{Port: 80, Proto: "tcp"},
	}

	d1 := dt.Compute("live", scan1)
	dt.Invalidate("live")
	d2 := dt.Compute("live", scan2)

	if d1 != d2 {
		t.Fatalf("reordered scan should produce same digest: %s vs %s", d1, d2)
	}
}

func TestDigest_ChangedAfterNewPort(t *testing.T) {
	dt := NewDigestTracker(DefaultDigestPolicy())

	before := []PortEntry{{Port: 80, Proto: "tcp"}}
	after := []PortEntry{{Port: 80, Proto: "tcp"}, {Port: 8080, Proto: "tcp"}}

	d1 := dt.Compute("live", before)
	dt.Invalidate("live")
	d2 := dt.Compute("live", after)

	if d1 == d2 {
		t.Fatal("expected digest to differ after new port added")
	}
}

func TestDigest_TTLExpiry(t *testing.T) {
	now := time.Now()
	offset := 0 * time.Second
	clock := func() time.Time { return now.Add(offset) }

	policy := DigestPolicy{TTL: 1 * time.Second}
	dt := newDigestTrackerWithClock(policy, clock)

	entries := []PortEntry{{Port: 443, Proto: "tcp"}}
	d1 := dt.Compute("live", entries)

	// Advance past TTL — cache should be stale.
	offset = 2 * time.Second

	newEntries := []PortEntry{{Port: 9090, Proto: "tcp"}}
	d2 := dt.Compute("live", newEntries)

	if d1 == d2 {
		t.Fatal("expected digest to differ after TTL expiry and new entries")
	}
}
