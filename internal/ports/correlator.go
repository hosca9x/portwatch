package ports

import (
	"sync"
	"time"
)

// CorrelationEvent represents a port event with a timestamp and direction.
type CorrelationEvent struct {
	Key       string
	Port      int
	Proto     string
	Opened    bool
	Timestamp time.Time
}

// CorrelationReport groups related events that occurred within a time window.
type CorrelationReport struct {
	Group     string
	Events    []CorrelationEvent
	StartedAt time.Time
	EndedAt   time.Time
}

// Correlator groups port events that occur within a short window,
// allowing detection of coordinated open/close sequences.
type Correlator struct {
	mu      sync.Mutex
	window  time.Duration
	events  []CorrelationEvent
	clock   func() time.Time
}

// NewCorrelator returns a Correlator with the given correlation window.
func NewCorrelator(window time.Duration) *Correlator {
	return &Correlator{
		window: window,
		clock:  time.Now,
	}
}

// newCorrelatorWithClock is used in tests to inject a custom clock.
func newCorrelatorWithClock(window time.Duration, clock func() time.Time) *Correlator {
	return &Correlator{
		window: window,
		clock:  clock,
	}
}

// Record adds a port event to the correlator.
func (c *Correlator) Record(entry PortEntry, opened bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = append(c.events, CorrelationEvent{
		Key:       entry.Key(),
		Port:      entry.Port,
		Proto:     entry.Proto,
		Opened:    opened,
		Timestamp: c.clock(),
	})
}

// Flush returns a CorrelationReport of all events within the current window
// and clears stale events older than the window.
func (c *Correlator) Flush() *CorrelationReport {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.clock()
	cutoff := now.Add(-c.window)

	var active []CorrelationEvent
	for _, e := range c.events {
		if e.Timestamp.After(cutoff) {
			active = append(active, e)
		}
	}
	c.events = active

	if len(active) == 0 {
		return nil
	}

	report := &CorrelationReport{
		Group:     "correlated",
		Events:    active,
		StartedAt: active[0].Timestamp,
		EndedAt:   active[len(active)-1].Timestamp,
	}
	return report
}

// Reset clears all recorded events.
func (c *Correlator) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = nil
}
