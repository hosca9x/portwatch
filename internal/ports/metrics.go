package ports

import (
	"sync"
	"time"
)

// ScanMetrics holds statistics collected across daemon scan cycles.
type ScanMetrics struct {
	mu sync.Mutex

	TotalScans    int
	TotalNew      int
	TotalClosed   int
	LastScanAt    time.Time
	LastScanPorts int
	Errors        int
}

// Collector accumulates scan metrics over time.
type Collector struct {
	metrics ScanMetrics
	clock   func() time.Time
}

// NewCollector returns a Collector using the real wall clock.
func NewCollector() *Collector {
	return &Collector{clock: time.Now}
}

// RecordScan records the outcome of a single scan cycle.
func (c *Collector) RecordScan(portCount, newPorts, closedPorts int, err error) {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	c.metrics.TotalScans++
	c.metrics.TotalNew += newPorts
	c.metrics.TotalClosed += closedPorts
	c.metrics.LastScanAt = c.clock()
	c.metrics.LastScanPorts = portCount
	if err != nil {
		c.metrics.Errors++
	}
}

// Snapshot returns a copy of the current metrics.
func (c *Collector) Snapshot() ScanMetrics {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	return ScanMetrics{
		TotalScans:    c.metrics.TotalScans,
		TotalNew:      c.metrics.TotalNew,
		TotalClosed:   c.metrics.TotalClosed,
		LastScanAt:    c.metrics.LastScanAt,
		LastScanPorts: c.metrics.LastScanPorts,
		Errors:        c.metrics.Errors,
	}
}

// Reset zeroes all accumulated metrics.
func (c *Collector) Reset() {
	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()
	c.metrics = ScanMetrics{}
}
