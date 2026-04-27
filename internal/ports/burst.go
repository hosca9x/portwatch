package ports

import (
	"sync"
	"time"
)

// BurstPolicy configures burst detection parameters.
type BurstPolicy struct {
	Window    time.Duration
	Threshold int
}

// DefaultBurstPolicy returns sensible defaults.
func DefaultBurstPolicy() BurstPolicy {
	return BurstPolicy{
		Window:    30 * time.Second,
		Threshold: 5,
	}
}

// BurstReport describes a detected burst for a given key.
type BurstReport struct {
	Key       string
	Count     int
	Window    time.Duration
	DetectedAt time.Time
}

type burstEvent struct {
	at time.Time
}

// BurstDetector tracks rapid port-event frequency and reports bursts.
type BurstDetector struct {
	mu     sync.Mutex
	policy BurstPolicy
	events map[string][]burstEvent
	clock  func() time.Time
}

// NewBurstDetector creates a BurstDetector with the given policy.
func NewBurstDetector(p BurstPolicy) *BurstDetector {
	return newBurstDetectorWithClock(p, time.Now)
}

func newBurstDetectorWithClock(p BurstPolicy, clock func() time.Time) *BurstDetector {
	return &BurstDetector{
		policy: p,
		events: make(map[string][]burstEvent),
		clock:  clock,
	}
}

// Record adds an event for key and returns a BurstReport if the threshold is met.
func (b *BurstDetector) Record(key string) *BurstReport {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.clock()
	cutoff := now.Add(-b.policy.Window)

	evs := b.events[key]
	filtered := evs[:0]
	for _, e := range evs {
		if e.at.After(cutoff) {
			filtered = append(filtered, e)
		}
	}
	filtered = append(filtered, burstEvent{at: now})
	b.events[key] = filtered

	if len(filtered) >= b.policy.Threshold {
		return &BurstReport{
			Key:        key,
			Count:      len(filtered),
			Window:     b.policy.Window,
			DetectedAt: now,
		}
	}
	return nil
}

// Reset clears all events for the given key.
func (b *BurstDetector) Reset(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.events, key)
}
