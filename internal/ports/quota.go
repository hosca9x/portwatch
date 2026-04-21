package ports

import (
	"sync"
	"time"
)

// QuotaPolicy defines limits for the quota tracker.
type QuotaPolicy struct {
	MaxEvents int           // maximum events allowed in the window
	Window    time.Duration // rolling window duration
}

// DefaultQuotaPolicy returns a sensible default quota policy.
func DefaultQuotaPolicy() QuotaPolicy {
	return QuotaPolicy{
		MaxEvents: 100,
		Window:    time.Minute,
	}
}

// QuotaTracker tracks per-key event counts within a rolling time window
// and reports whether a key has exceeded its quota.
type QuotaTracker struct {
	mu     sync.Mutex
	policy QuotaPolicy
	events map[string][]time.Time
	clock  func() time.Time
}

// NewQuotaTracker creates a QuotaTracker with the given policy.
func NewQuotaTracker(policy QuotaPolicy) *QuotaTracker {
	return newQuotaTrackerWithClock(policy, time.Now)
}

func newQuotaTrackerWithClock(policy QuotaPolicy, clock func() time.Time) *QuotaTracker {
	return &QuotaTracker{
		policy: policy,
		events: make(map[string][]time.Time),
		clock:  clock,
	}
}

// Record records an event for key and returns true if the quota has been exceeded.
func (q *QuotaTracker) Record(key string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.clock()
	cutoff := now.Add(-q.policy.Window)

	times := q.events[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	filtered = append(filtered, now)
	q.events[key] = filtered

	return len(filtered) > q.policy.MaxEvents
}

// Count returns the current number of events within the window for key.
func (q *QuotaTracker) Count(key string) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := q.clock()
	cutoff := now.Add(-q.policy.Window)
	count := 0
	for _, t := range q.events[key] {
		if t.After(cutoff) {
			count++
		}
	}
	return count
}

// Reset clears recorded events for key.
func (q *QuotaTracker) Reset(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.events, key)
}
