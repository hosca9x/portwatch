package ports

import (
	"sync"
	"time"
)

// WindowPolicy defines the configuration for a sliding window counter.
type WindowPolicy struct {
	Size     time.Duration
	MaxCount int
}

// DefaultWindowPolicy returns sensible defaults for a sliding window.
func DefaultWindowPolicy() WindowPolicy {
	return WindowPolicy{
		Size:     time.Minute,
		MaxCount: 20,
	}
}

// WindowCounter tracks event counts within a sliding time window per key.
type WindowCounter struct {
	mu     sync.Mutex
	policy WindowPolicy
	events map[string][]time.Time
	clock  func() time.Time
}

// NewWindowCounter creates a WindowCounter with the given policy.
func NewWindowCounter(policy WindowPolicy) *WindowCounter {
	return newWindowCounterWithClock(policy, time.Now)
}

func newWindowCounterWithClock(policy WindowPolicy, clock func() time.Time) *WindowCounter {
	return &WindowCounter{
		policy: policy,
		events: make(map[string][]time.Time),
		clock:  clock,
	}
}

// Add records an event for the given key and returns the current count within the window.
func (w *WindowCounter) Add(key string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.clock()
	cutoff := now.Add(-w.policy.Size)

	filtered := w.events[key][:0]
	for _, t := range w.events[key] {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	filtered = append(filtered, now)
	w.events[key] = filtered
	return len(filtered)
}

// Count returns the number of events for key within the current window.
func (w *WindowCounter) Count(key string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.clock()
	cutoff := now.Add(-w.policy.Size)
	count := 0
	for _, t := range w.events[key] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Exceeded reports whether the count for key is at or above MaxCount.
func (w *WindowCounter) Exceeded(key string) bool {
	return w.Count(key) >= w.policy.MaxCount
}

// Reset clears all recorded events for the given key.
func (w *WindowCounter) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.events, key)
}
