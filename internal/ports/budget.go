package ports

import (
	"sync"
	"time"
)

// BudgetPolicy controls how scan budgets are allocated and replenished.
type BudgetPolicy struct {
	MaxScansPerWindow int
	Window            time.Duration
}

// DefaultBudgetPolicy returns sensible defaults.
func DefaultBudgetPolicy() BudgetPolicy {
	return BudgetPolicy{
		MaxScansPerWindow: 100,
		Window:            time.Minute,
	}
}

// BudgetTracker enforces a rolling scan budget per key.
type BudgetTracker struct {
	mu     sync.Mutex
	policy BudgetPolicy
	clock  func() time.Time
	events map[string][]time.Time
}

// NewBudgetTracker creates a BudgetTracker with the given policy.
func NewBudgetTracker(policy BudgetPolicy) *BudgetTracker {
	return newBudgetTrackerWithClock(policy, time.Now)
}

func newBudgetTrackerWithClock(policy BudgetPolicy, clock func() time.Time) *BudgetTracker {
	return &BudgetTracker{
		policy: policy,
		clock:  clock,
		events: make(map[string][]time.Time),
	}
}

// Record attempts to consume one unit of budget for key.
// Returns true if budget is available, false if exhausted.
func (b *BudgetTracker) Record(key string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.clock()
	cutoff := now.Add(-b.policy.Window)

	prev := b.events[key]
	var active []time.Time
	for _, t := range prev {
		if t.After(cutoff) {
			active = append(active, t)
		}
	}

	if len(active) >= b.policy.MaxScansPerWindow {
		b.events[key] = active
		return false
	}

	b.events[key] = append(active, now)
	return true
}

// Remaining returns how many scan units are left for key within the current window.
func (b *BudgetTracker) Remaining(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.clock()
	cutoff := now.Add(-b.policy.Window)

	count := 0
	for _, t := range b.events[key] {
		if t.After(cutoff) {
			count++
		}
	}

	remaining := b.policy.MaxScansPerWindow - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Reset clears all recorded events for key.
func (b *BudgetTracker) Reset(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.events, key)
}
