package ports

import (
	"sync"
	"time"
)

// TrendDirection indicates whether port activity is increasing or decreasing.
type TrendDirection string

const (
	TrendUp     TrendDirection = "up"
	TrendDown   TrendDirection = "down"
	TrendStable TrendDirection = "stable"
)

// TrendSample records the number of open ports at a point in time.
type TrendSample struct {
	At    time.Time
	Count int
}

// TrendReport summarises recent port-count movement.
type TrendReport struct {
	Direction TrendDirection
	Delta     int
	Samples   []TrendSample
}

// TrendTracker maintains a rolling window of port-count samples and
// derives a simple trend direction from them.
type TrendTracker struct {
	mu      sync.Mutex
	window  int
	samples []TrendSample
	clock   func() time.Time
}

// NewTrendTracker creates a TrendTracker that keeps at most windowSize samples.
func NewTrendTracker(windowSize int) *TrendTracker {
	if windowSize < 2 {
		windowSize = 2
	}
	return &TrendTracker{
		window: windowSize,
		clock:  time.Now,
	}
}

// Record adds a new sample with the current port count.
func (t *TrendTracker) Record(count int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.samples = append(t.samples, TrendSample{At: t.clock(), Count: count})
	if len(t.samples) > t.window {
		t.samples = t.samples[len(t.samples)-t.window:]
	}
}

// Report returns the current trend based on the first and last sample.
func (t *TrendTracker) Report() TrendReport {
	t.mu.Lock()
	defer t.mu.Unlock()

	snap := make([]TrendSample, len(t.samples))
	copy(snap, t.samples)

	if len(snap) < 2 {
		return TrendReport{Direction: TrendStable, Samples: snap}
	}

	first := snap[0].Count
	last := snap[len(snap)-1].Count
	delta := last - first

	var dir TrendDirection
	switch {
	case delta > 0:
		dir = TrendUp
	case delta < 0:
		dir = TrendDown
	default:
		dir = TrendStable
	}

	return TrendReport{Direction: dir, Delta: delta, Samples: snap}
}

// Reset clears all recorded samples.
func (t *TrendTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.samples = nil
}
