package ports

import (
	"sync"
	"time"
)

// EvictMapPolicy controls expiry behaviour for EvictMap.
type EvictMapPolicy struct {
	TTL time.Duration
}

// DefaultEvictMapPolicy returns a sensible default TTL.
func DefaultEvictMapPolicy() EvictMapPolicy {
	return EvictMapPolicy{TTL: 5 * time.Minute}
}

type evictEntry struct {
	value     interface{}
	insertedAt time.Time
}

// EvictMap is a map that automatically expires entries after a TTL.
type EvictMap struct {
	mu     sync.Mutex
	policy EvictMapPolicy
	clock  func() time.Time
	items  map[string]evictEntry
}

// NewEvictMap creates an EvictMap with the given policy.
func NewEvictMap(policy EvictMapPolicy) *EvictMap {
	return newEvictMapWithClock(policy, time.Now)
}

func newEvictMapWithClock(policy EvictMapPolicy, clock func() time.Time) *EvictMap {
	return &EvictMap{
		policy: policy,
		clock:  clock,
		items:  make(map[string]evictEntry),
	}
}

// Set stores a value under key, resetting its TTL.
func (e *EvictMap) Set(key string, value interface{}) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.items[key] = evictEntry{value: value, insertedAt: e.clock()}
}

// Get retrieves a value if it exists and has not expired.
// Returns (value, true) on hit, (nil, false) on miss or expiry.
func (e *EvictMap) Get(key string) (interface{}, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	ent, ok := e.items[key]
	if !ok {
		return nil, false
	}
	if e.clock().Sub(ent.insertedAt) > e.policy.TTL {
		delete(e.items, key)
		return nil, false
	}
	return ent.value, true
}

// Delete removes a key unconditionally.
func (e *EvictMap) Delete(key string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.items, key)
}

// Purge removes all expired entries and returns the count removed.
func (e *EvictMap) Purge() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := e.clock()
	removed := 0
	for k, ent := range e.items {
		if now.Sub(ent.insertedAt) > e.policy.TTL {
			delete(e.items, k)
			removed++
		}
	}
	return removed
}

// Len returns the number of entries currently held (including potentially expired ones).
func (e *EvictMap) Len() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.items)
}
