package ports

import "time"

// CascadePolicy controls cascade detection behaviour.
type CascadePolicy struct {
	// MinPorts is the minimum number of distinct ports that must open within
	// Window to consider the event a cascade.
	MinPorts int
	// Window is the time range examined when counting distinct ports.
	Window time.Duration
}

// DefaultCascadePolicy returns sensible defaults.
func DefaultCascadePolicy() CascadePolicy {
	return CascadePolicy{
		MinPorts: 5,
		Window:   30 * time.Second,
	}
}

// CascadeReport describes a detected cascade event.
type CascadeReport struct {
	Ports     []PortEntry
	DetectedAt time.Time
}

// CascadeDetector watches for rapid bursts of newly-opened ports.
type CascadeDetector struct {
	policy CascadePolicy
	events []cascadeEvent
	clock  func() time.Time
}

type cascadeEvent struct {
	entry PortEntry
	at    time.Time
}

// NewCascadeDetector creates a CascadeDetector with the given policy.
func NewCascadeDetector(p CascadePolicy) *CascadeDetector {
	return newCascadeDetectorWithClock(p, time.Now)
}

func newCascadeDetectorWithClock(p CascadePolicy, clock func() time.Time) *CascadeDetector {
	return &CascadeDetector{policy: p, clock: clock}
}

// Record adds a newly-opened port to the detector's sliding window.
// If the number of distinct ports within the window meets the threshold a
// CascadeReport is returned; otherwise nil is returned.
func (c *CascadeDetector) Record(e PortEntry) *CascadeReport {
	now := c.clock()
	cutoff := now.Add(-c.policy.Window)

	// evict stale events
	filtered := c.events[:0]
	for _, ev := range c.events {
		if ev.at.After(cutoff) {
			filtered = append(filtered, ev)
		}
	}
	filtered = append(filtered, cascadeEvent{entry: e, at: now})
	c.events = filtered

	if len(c.events) < c.policy.MinPorts {
		return nil
	}

	ports := make([]PortEntry, len(c.events))
	for i, ev := range c.events {
		ports[i] = ev.entry
	}
	return &CascadeReport{Ports: ports, DetectedAt: now}
}

// Reset clears all recorded events.
func (c *CascadeDetector) Reset() {
	c.events = c.events[:0]
}
