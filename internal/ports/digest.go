package ports

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
	"time"
)

// DigestPolicy controls digest tracker behaviour.
type DigestPolicy struct {
	TTL time.Duration
}

// DefaultDigestPolicy returns sensible defaults.
func DefaultDigestPolicy() DigestPolicy {
	return DigestPolicy{
		TTL: 10 * time.Minute,
	}
}

// DigestEntry records a computed digest alongside its timestamp.
type DigestEntry struct {
	Digest    string
	ComputedAt time.Time
}

// DigestTracker computes and caches rolling digests of port-entry sets.
type DigestTracker struct {
	mu      sync.Mutex
	policy  DigestPolicy
	clock   func() time.Time
	cache   map[string]DigestEntry
}

// NewDigestTracker returns a DigestTracker with the given policy.
func NewDigestTracker(p DigestPolicy) *DigestTracker {
	return newDigestTrackerWithClock(p, time.Now)
}

func newDigestTrackerWithClock(p DigestPolicy, clock func() time.Time) *DigestTracker {
	return &DigestTracker{
		policy: p,
		clock:  clock,
		cache:  make(map[string]DigestEntry),
	}
}

// Compute returns a stable hex digest for the given entries, keyed by label.
// Results are cached until TTL expires.
func (d *DigestTracker) Compute(label string, entries []PortEntry) string {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.clock()
	if e, ok := d.cache[label]; ok && now.Sub(e.ComputedAt) < d.policy.TTL {
		return e.Digest
	}

	keys := make([]string, len(entries))
	for i, e := range entries {
		keys[i] = fmt.Sprintf("%s:%d", e.Proto, e.Port)
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		h.Write([]byte(k))
	}
	digest := hex.EncodeToString(h.Sum(nil))

	d.cache[label] = DigestEntry{Digest: digest, ComputedAt: now}
	return digest
}

// Invalidate removes a cached digest entry by label.
func (d *DigestTracker) Invalidate(label string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.cache, label)
}

// Cached returns the cached DigestEntry for label, if present and not expired.
func (d *DigestTracker) Cached(label string) (DigestEntry, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.clock()
	e, ok := d.cache[label]
	if !ok || now.Sub(e.ComputedAt) >= d.policy.TTL {
		return DigestEntry{}, false
	}
	return e, true
}
