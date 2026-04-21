package ports

import (
	"testing"
	"time"
)

func fixedScoreboardClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestScoreboard_FirstRecordSetsScore(t *testing.T) {
	now := time.Unix(1_000, 0)
	sb := newScoreboardWithClock(DefaultScoreboardDecay, 0.5, fixedScoreboardClock(now))
	sb.Record("tcp:80", 1.0)

	e := sb.Get("tcp:80")
	if e == nil {
		t.Fatal("expected entry, got nil")
	}
	if e.Score != 1.0 {
		t.Fatalf("expected score 1.0, got %f", e.Score)
	}
	if e.Hits != 1 {
		t.Fatalf("expected 1 hit, got %d", e.Hits)
	}
}

func TestScoreboard_AccumulatesWithoutDecay(t *testing.T) {
	now := time.Unix(1_000, 0)
	// Use decay factor 1.0 so score never decays
	sb := newScoreboardWithClock(DefaultScoreboardDecay, 1.0, fixedScoreboardClock(now))
	sb.Record("tcp:443", 2.0)
	sb.Record("tcp:443", 3.0)

	e := sb.Get("tcp:443")
	if e == nil {
		t.Fatal("expected entry")
	}
	if e.Score != 5.0 {
		t.Fatalf("expected 5.0, got %f", e.Score)
	}
	if e.Hits != 2 {
		t.Fatalf("expected 2 hits, got %d", e.Hits)
	}
}

func TestScoreboard_DecaysOverTime(t *testing.T) {
	base := time.Unix(0, 0)
	var current time.Time
	clock := func() time.Time { return current }

	// decay factor 0.5 per DefaultScoreboardDecay (5 min)
	sb := newScoreboardWithClock(DefaultScoreboardDecay, 0.5, clock)

	current = base
	sb.Record("udp:53", 8.0)

	// advance one full decay period → score should halve before new weight added
	current = base.Add(DefaultScoreboardDecay)
	sb.Record("udp:53", 0.0)

	e := sb.Get("udp:53")
	if e == nil {
		t.Fatal("expected entry")
	}
	// 8.0 * 0.5^1 + 0.0 = 4.0
	if e.Score < 3.9 || e.Score > 4.1 {
		t.Fatalf("expected ~4.0 after decay, got %f", e.Score)
	}
}

func TestScoreboard_GetMissingKeyReturnsNil(t *testing.T) {
	now := time.Unix(0, 0)
	sb := newScoreboardWithClock(DefaultScoreboardDecay, 0.5, fixedScoreboardClock(now))
	if sb.Get("tcp:9999") != nil {
		t.Fatal("expected nil for missing key")
	}
}

func TestScoreboard_ResetRemovesEntry(t *testing.T) {
	now := time.Unix(0, 0)
	sb := newScoreboardWithClock(DefaultScoreboardDecay, 0.5, fixedScoreboardClock(now))
	sb.Record("tcp:22", 5.0)
	sb.Reset("tcp:22")
	if sb.Get("tcp:22") != nil {
		t.Fatal("expected nil after reset")
	}
}

func TestScoreboard_IndependentKeys(t *testing.T) {
	now := time.Unix(0, 0)
	sb := newScoreboardWithClock(DefaultScoreboardDecay, 1.0, fixedScoreboardClock(now))
	sb.Record("tcp:80", 3.0)
	sb.Record("tcp:443", 7.0)

	a := sb.Get("tcp:80")
	b := sb.Get("tcp:443")
	if a.Score != 3.0 || b.Score != 7.0 {
		t.Fatalf("scores mixed up: a=%f b=%f", a.Score, b.Score)
	}
}
