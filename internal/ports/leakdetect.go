package ports

import (
	"sync"
	"time"
)

// LeakReport describes a port that has been open continuously beyond a threshold.
type LeakReport struct {
	Key       string
	Port      int
	Proto     string
	FirstSeen time.Time
	Duration  time.Duration
}

// LeakDetectorConfig controls leak detection behaviour.
type LeakDetectorConfig struct {
	Threshold time.Duration // how long a port must stay open to be considered a leak
}

// DefaultLeakDetectorConfig returns sensible defaults.
func DefaultLeakDetectorConfig() LeakDetectorConfig {
	return LeakDetectorConfig{
		Threshold: 24 * time.Hour,
	}
}

type leakEntry struct {
	port      int
	proto     string
	firstSeen time.Time
}

// LeakDetector tracks how long ports stay open and reports those exceeding the threshold.
type LeakDetector struct {
	mu    sync.Mutex
	cfg   LeakDetectorConfig
	seen  map[string]leakEntry
	clock func() time.Time
}

// NewLeakDetector creates a LeakDetector with the given config.
func NewLeakDetector(cfg LeakDetectorConfig) *LeakDetector {
	return newLeakDetectorWithClock(cfg, time.Now)
}

func newLeakDetectorWithClock(cfg LeakDetectorConfig, clock func() time.Time) *LeakDetector {
	return &LeakDetector{
		cfg:   cfg,
		seen:  make(map[string]leakEntry),
		clock: clock,
	}
}

// Observe records the current open ports. Returns any that have exceeded the threshold.
func (d *LeakDetector) Observe(entries []PortEntry) []LeakReport {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.clock()
	current := make(map[string]struct{}, len(entries))

	for _, e := range entries {
		current[e.Key()] = struct{}{}
		if _, ok := d.seen[e.Key()]; !ok {
			d.seen[e.Key()] = leakEntry{
				port:      e.Port,
				proto:     e.Proto,
				firstSeen: now,
			}
		}
	}

	// Evict ports no longer open.
	for k := range d.seen {
		if _, ok := current[k]; !ok {
			delete(d.seen, k)
		}
	}

	var reports []LeakReport
	for k, le := range d.seen {
		dur := now.Sub(le.firstSeen)
		if dur >= d.cfg.Threshold {
			reports = append(reports, LeakReport{
				Key:       k,
				Port:      le.port,
				Proto:     le.proto,
				FirstSeen: le.firstSeen,
				Duration:  dur,
			})
		}
	}
	return reports
}

// Reset clears all tracked state.
func (d *LeakDetector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seen = make(map[string]leakEntry)
}
