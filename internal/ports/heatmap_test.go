package ports

import (
	"testing"
	"time"
)

func fixedHeatmapClock(t time.Time) heatmapClock {
	return func() time.Time { return t }
}

func TestHeatmap_InitialSnapshotEmpty(t *testing.T) {
	h := newHeatmapWithClock(DefaultHeatmapPolicy(), fixedHeatmapClock(time.Now()))
	if got := h.Snapshot(); len(got) != 0 {
		t.Fatalf("expected empty snapshot, got %d entries", len(got))
	}
}

func TestHeatmap_RecordIncreasesCount(t *testing.T) {
	now := time.Now()
	h := newHeatmapWithClock(DefaultHeatmapPolicy(), fixedHeatmapClock(now))
	h.Record("tcp:80")
	h.Record("tcp:80")
	snap := h.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 report, got %d", len(snap))
	}
	if snap[0].Count != 2 {
		t.Errorf("expected count 2, got %d", snap[0].Count)
	}
}

func TestHeatmap_HeatCalculated(t *testing.T) {
	now := time.Now()
	p := HeatmapPolicy{WindowSize: 10 * time.Minute}
	h := newHeatmapWithClock(p, fixedHeatmapClock(now))
	for i := 0; i < 5; i++ {
		h.Record("udp:53")
	}
	snap := h.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 report, got %d", len(snap))
	}
	want := 5.0 / 10.0 // 5 events / 10 minutes
	if snap[0].Heat != want {
		t.Errorf("expected heat %.4f, got %.4f", want, snap[0].Heat)
	}
}

func TestHeatmap_OldEventsExpire(t *testing.T) {
	base := time.Now()
	p := HeatmapPolicy{WindowSize: 5 * time.Minute}
	var current time.Time
	clk := func() time.Time { return current }
	h := newHeatmapWithClock(p, clk)

	current = base
	h.Record("tcp:443")

	// advance past window
	current = base.Add(6 * time.Minute)
	h.Record("tcp:443") // one fresh event

	snap := h.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 report, got %d", len(snap))
	}
	if snap[0].Count != 1 {
		t.Errorf("expected count 1 after expiry, got %d", snap[0].Count)
	}
}

func TestHeatmap_IndependentKeys(t *testing.T) {
	now := time.Now()
	h := newHeatmapWithClock(DefaultHeatmapPolicy(), fixedHeatmapClock(now))
	h.Record("tcp:22")
	h.Record("tcp:22")
	h.Record("udp:123")
	snap := h.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(snap))
	}
}
