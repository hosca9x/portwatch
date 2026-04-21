package ports

import (
	"sync"
	"time"
)

// RollupPolicy configures how events are aggregated.
type RollupPolicy struct {
	Window   time.Duration
	MaxItems int
}

// DefaultRollupPolicy returns sensible defaults.
func DefaultRollupPolicy() RollupPolicy {
	return RollupPolicy{
		Window:   30 * time.Second,
		MaxItems: 100,
	}
}

// RollupBucket holds aggregated entries within a time window.
type RollupBucket struct {
	Key     string
	Entries []PortEntry
	First   time.Time
	Last    time.Time
}

// Rollup aggregates PortEntry events by key over a sliding window.
type Rollup struct {
	mu      sync.Mutex
	policy  RollupPolicy
	buckets map[string]*RollupBucket
	clock   func() time.Time
}

// NewRollup creates a Rollup with the given policy.
func NewRollup(policy RollupPolicy) *Rollup {
	return newRollupWithClock(policy, time.Now)
}

func newRollupWithClock(policy RollupPolicy, clock func() time.Time) *Rollup {
	if policy.MaxItems <= 0 {
		policy.MaxItems = DefaultRollupPolicy().MaxItems
	}
	if policy.Window <= 0 {
		policy.Window = DefaultRollupPolicy().Window
	}
	return &Rollup{
		policy:  policy,
		buckets: make(map[string]*RollupBucket),
		clock:   clock,
	}
}

// Add records a PortEntry under the given key.
func (r *Rollup) Add(key string, entry PortEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.clock()
	b, ok := r.buckets[key]
	if !ok || now.Sub(b.First) > r.policy.Window {
		b = &RollupBucket{Key: key, First: now}
		r.buckets[key] = b
	}
	if len(b.Entries) < r.policy.MaxItems {
		b.Entries = append(b.Entries, entry)
	}
	b.Last = now
}

// Flush returns and clears the bucket for the given key.
// Returns nil if no bucket exists.
func (r *Rollup) Flush(key string) *RollupBucket {
	r.mu.Lock()
	defer r.mu.Unlock()
	b, ok := r.buckets[key]
	if !ok {
		return nil
	}
	delete(r.buckets, key)
	return b
}

// Keys returns all active bucket keys.
func (r *Rollup) Keys() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	keys := make([]string, 0, len(r.buckets))
	for k := range r.buckets {
		keys = append(keys, k)
	}
	return keys
}
