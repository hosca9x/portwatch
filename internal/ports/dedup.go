package ports

import (
	"sync"
	"time"
)

// Deduplicator suppresses duplicate port change events within a rolling window.
// It is safe for concurrent use.
type Deduplicator struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	window  time.Duration
	clock   func() time.Time
}

// NewDeduplicator creates a Deduplicator that considers events with the same
// key as duplicates if they occur within window duration of each other.
func NewDeduplicator(window time.Duration) *Deduplicator {
	return &Deduplicator{
		seen:   make(map[string]time.Time),
		window: window,
		clock:  time.Now,
	}
}

// IsDuplicate reports whether key has been seen within the dedup window.
// If it is not a duplicate, the key's timestamp is updated.
func (d *Deduplicator) IsDuplicate(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.clock()
	if last, ok := d.seen[key]; ok && now.Sub(last) < d.window {
		return true
	}
	d.seen[key] = now
	return false
}

// Reset clears the recorded state for a specific key.
func (d *Deduplicator) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.seen, key)
}

// Flush removes all expired entries from the internal map.
func (d *Deduplicator) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.clock()
	for k, t := range d.seen {
		if now.Sub(t) >= d.window {
			delete(d.seen, k)
		}
	}
}

// Len returns the number of keys currently tracked.
func (d *Deduplicator) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}
