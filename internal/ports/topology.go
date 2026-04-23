package ports

import (
	"sync"
	"time"
)

// TopologySnapshot captures a point-in-time view of which ports are
// co-observed together within a single scan cycle.
type TopologySnapshot struct {
	Timestamp time.Time
	Edges     map[string][]string // port key -> co-observed port keys
}

// TopologyTracker records co-occurrence relationships between open ports
// across scan cycles and exposes the current adjacency view.
type TopologyTracker struct {
	mu      sync.Mutex
	clock   func() time.Time
	window  time.Duration
	buckets []topologyBucket
}

type topologyBucket struct {
	at    time.Time
	pairs [][2]string
}

// DefaultTopologyWindow is the rolling window used to build the graph.
const DefaultTopologyWindow = 5 * time.Minute

// NewTopologyTracker creates a tracker with the given rolling window.
func NewTopologyTracker(window time.Duration) *TopologyTracker {
	return newTopologyTrackerWithClock(window, time.Now)
}

func newTopologyTrackerWithClock(window time.Duration, clock func() time.Time) *TopologyTracker {
	if window <= 0 {
		window = DefaultTopologyWindow
	}
	return &TopologyTracker{clock: clock, window: window}
}

// Record ingests a scan result and stores co-occurrence pairs for this cycle.
func (t *TopologyTracker) Record(entries []PortEntry) {
	if len(entries) < 2 {
		return
	}
	now := t.clock()
	var pairs [][2]string
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			pairs = append(pairs, [2]string{entries[i].Key(), entries[j].Key()})
		}
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buckets = append(t.buckets, topologyBucket{at: now, pairs: pairs})
	t.evict(now)
}

// Snapshot returns the current adjacency map built from the rolling window.
func (t *TopologyTracker) Snapshot() TopologySnapshot {
	now := t.clock()
	t.mu.Lock()
	defer t.mu.Unlock()
	t.evict(now)
	adj := make(map[string][]string)
	seen := make(map[[2]string]struct{})
	for _, b := range t.buckets {
		for _, p := range b.pairs {
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			adj[p[0]] = append(adj[p[0]], p[1])
			adj[p[1]] = append(adj[p[1]], p[0])
		}
	}
	return TopologySnapshot{Timestamp: now, Edges: adj}
}

func (t *TopologyTracker) evict(now time.Time) {
	cutoff := now.Add(-t.window)
	start := 0
	for start < len(t.buckets) && t.buckets[start].at.Before(cutoff) {
		start++
	}
	t.buckets = t.buckets[start:]
}
