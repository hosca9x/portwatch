package ports

import (
	"sync"
	"time"
)

// CensusPolicy controls census tracker behaviour.
type CensusPolicy struct {
	// MaxAge is how long a port entry remains in the census before expiry.
	MaxAge time.Duration
}

// DefaultCensusPolicy returns sensible defaults.
func DefaultCensusPolicy() CensusPolicy {
	return CensusPolicy{
		MaxAge: 10 * time.Minute,
	}
}

// CensusEntry records the first and last seen timestamps for a port key.
type CensusEntry struct {
	Key       string
	FirstSeen time.Time
	LastSeen  time.Time
	Count     int
}

// CensusSnapshot is a point-in-time view of all tracked entries.
type CensusSnapshot struct {
	TakenAt time.Time
	Entries []CensusEntry
}

// CensusTracker records how often each port key has been observed and
// when it was first and last seen, evicting stale entries.
type CensusTracker struct {
	mu      sync.Mutex
	policy  CensusPolicy
	entries map[string]*CensusEntry
	clock   func() time.Time
}

// NewCensusTracker creates a CensusTracker with the given policy.
func NewCensusTracker(policy CensusPolicy) *CensusTracker {
	return newCensusTrackerWithClock(policy, time.Now)
}

func newCensusTrackerWithClock(policy CensusPolicy, clock func() time.Time) *CensusTracker {
	return &CensusTracker{
		policy:  policy,
		entries: make(map[string]*CensusEntry),
		clock:   clock,
	}
}

// Record observes a port entry, updating counts and timestamps.
func (c *CensusTracker) Record(e PortEntry) {
	now := c.clock()
	c.mu.Lock()
	defer c.mu.Unlock()
	key := e.Key()
	if ce, ok := c.entries[key]; ok {
		ce.LastSeen = now
		ce.Count++
	} else {
		c.entries[key] = &CensusEntry{
			Key:       key,
			FirstSeen: now,
			LastSeen:  now,
			Count:     1,
		}
	}
}

// Snapshot returns all non-expired entries as a CensusSnapshot.
func (c *CensusTracker) Snapshot() CensusSnapshot {
	now := c.clock()
	c.mu.Lock()
	defer c.mu.Unlock()
	cutoff := now.Add(-c.policy.MaxAge)
	var out []CensusEntry
	for key, ce := range c.entries {
		if ce.LastSeen.Before(cutoff) {
			delete(c.entries, key)
			continue
		}
		out = append(out, *ce)
	}
	return CensusSnapshot{TakenAt: now, Entries: out}
}
