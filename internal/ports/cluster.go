package ports

import (
	"sync"
	"time"
)

// ClusterPolicy controls how ports are grouped into clusters.
type ClusterPolicy struct {
	// MaxGap is the maximum port number gap to still consider two ports in the same cluster.
	MaxGap int
	// MinSize is the minimum number of ports required to emit a cluster report.
	MinSize int
	// TTL is how long a cluster observation is retained.
	TTL time.Duration
}

// DefaultClusterPolicy returns sensible defaults.
func DefaultClusterPolicy() ClusterPolicy {
	return ClusterPolicy{
		MaxGap:  10,
		MinSize: 3,
		TTL:     5 * time.Minute,
	}
}

// ClusterReport describes a group of adjacent open ports.
type ClusterReport struct {
	Proto    string
	Start    int
	End      int
	Size     int
	SeenAt   time.Time
}

// ClusterTracker groups observed ports into contiguous clusters.
type ClusterTracker struct {
	mu     sync.Mutex
	policy ClusterPolicy
	clock  func() time.Time
	// proto -> port -> observedAt
	observed map[string]map[int]time.Time
}

// NewClusterTracker creates a ClusterTracker with the given policy.
func NewClusterTracker(p ClusterPolicy) *ClusterTracker {
	return newClusterTrackerWithClock(p, time.Now)
}

func newClusterTrackerWithClock(p ClusterPolicy, clock func() time.Time) *ClusterTracker {
	return &ClusterTracker{
		policy:   p,
		clock:    clock,
		observed: make(map[string]map[int]time.Time),
	}
}

// Record notes that the given port/proto was observed.
func (c *ClusterTracker) Record(e PortEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.observed[e.Proto]; !ok {
		c.observed[e.Proto] = make(map[int]time.Time)
	}
	c.observed[e.Proto][e.Port] = c.clock()
}

// Clusters returns all clusters that meet the MinSize threshold, evicting stale entries first.
func (c *ClusterTracker) Clusters() []ClusterReport {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock()
	var reports []ClusterReport
	for proto, ports := range c.observed {
		// evict stale
		for port, seen := range ports {
			if now.Sub(seen) > c.policy.TTL {
				delete(ports, port)
			}
		}
		reports = append(reports, buildClusters(proto, ports, c.policy, now)...)
	}
	return reports
}

func buildClusters(proto string, ports map[int]time.Time, p ClusterPolicy, now time.Time) []ClusterReport {
	if len(ports) == 0 {
		return nil
	}
	sorted := sortedKeys(ports)
	var reports []ClusterReport
	start := sorted[0]
	end := sorted[0]
	var latest time.Time
	for _, port := range sorted {
		if port-end > p.MaxGap {
			if end-start+1 >= p.MinSize {
				reports = append(reports, ClusterReport{Proto: proto, Start: start, End: end, Size: end - start + 1, SeenAt: latest})
			}
			start = port
			latest = time.Time{}
		}
		end = port
		if t := ports[port]; t.After(latest) {
			latest = t
		}
	}
	if end-start+1 >= p.MinSize {
		reports = append(reports, ClusterReport{Proto: proto, Start: start, End: end, Size: end - start + 1, SeenAt: latest})
	}
	return reports
}

func sortedKeys(m map[int]time.Time) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
