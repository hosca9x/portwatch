package ports

import (
	"testing"
	"time"
)

func fixedTrendClock(base time.Time) func() time.Time {
	var i int
	return func() time.Time {
		t := base.Add(time.Duration(i) * time.Second)
		i++
		return t
	}
}

func TestTrendTracker_StableOnSingleSample(t *testing.T) {
	tr := NewTrendTracker(5)
	tr.clock = fixedTrendClock(time.Now())
	tr.Record(10)

	r := tr.Report()
	if r.Direction != TrendStable {
		t.Fatalf("expected stable, got %s", r.Direction)
	}
}

func TestTrendTracker_DetectsUp(t *testing.T) {
	tr := NewTrendTracker(5)
	tr.clock = fixedTrendClock(time.Now())
	tr.Record(5)
	tr.Record(10)

	r := tr.Report()
	if r.Direction != TrendUp {
		t.Fatalf("expected up, got %s", r.Direction)
	}
	if r.Delta != 5 {
		t.Fatalf("expected delta 5, got %d", r.Delta)
	}
}

func TestTrendTracker_DetectsDown(t *testing.T) {
	tr := NewTrendTracker(5)
	tr.clock = fixedTrendClock(time.Now())
	tr.Record(20)
	tr.Record(8)

	r := tr.Report()
	if r.Direction != TrendDown {
		t.Fatalf("expected down, got %s", r.Direction)
	}
	if r.Delta != -12 {
		t.Fatalf("expected delta -12, got %d", r.Delta)
	}
}

func TestTrendTracker_WindowCapsOldSamples(t *testing.T) {
	tr := NewTrendTracker(3)
	tr.clock = fixedTrendClock(time.Now())
	for _, c := range []int{1, 2, 3, 4, 5} {
		tr.Record(c)
	}

	r := tr.Report()
	if len(r.Samples) != 3 {
		t.Fatalf("expected 3 samples, got %d", len(r.Samples))
	}
	if r.Samples[0].Count != 3 {
		t.Fatalf("expected oldest kept sample count=3, got %d", r.Samples[0].Count)
	}
}

func TestTrendTracker_Reset(t *testing.T) {
	tr := NewTrendTracker(5)
	tr.clock = fixedTrendClock(time.Now())
	tr.Record(10)
	tr.Record(20)
	tr.Reset()

	r := tr.Report()
	if len(r.Samples) != 0 {
		t.Fatalf("expected 0 samples after reset, got %d", len(r.Samples))
	}
	if r.Direction != TrendStable {
		t.Fatalf("expected stable after reset, got %s", r.Direction)
	}
}
