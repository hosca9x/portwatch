package ports

import (
	"sync"
	"time"
)

// AnomalyKind classifies the type of anomaly detected.
type AnomalyKind string

const (
	AnomalyFlapping AnomalyKind = "flapping" // port opens/closes repeatedly
	AnomalyBurst    AnomalyKind = "burst"    // many new ports opened in short window
)

// AnomalyReport describes a detected anomaly.
type AnomalyReport struct {
	Kind      AnomalyKind
	Key       string
	Count     int
	Window    time.Duration
	DetectedAt time.Time
}

// AnomalyDetector watches for flapping ports and burst openings.
type AnomalyDetector struct {
	mu          sync.Mutex
	clock       func() time.Time
	window      time.Duration
	flapThresh  int
	burstThresh int
	events      map[string][]time.Time // key -> timestamps of open/close events
}

// NewAnomalyDetector creates a detector with the given window and thresholds.
// flapThresh: number of open/close transitions within window to flag as flapping.
// burstThresh: number of distinct new ports within window to flag as burst.
func NewAnomalyDetector(window time.Duration, flapThresh, burstThresh int) *AnomalyDetector {
	return &AnomalyDetector{
		clock:       time.Now,
		window:      window,
		flapThresh:  flapThresh,
		burstThresh: burstThresh,
		events:      make(map[string][]time.Time),
	}
}

// Record registers a change event for the given key and returns any anomaly detected.
func (d *AnomalyDetector) Record(key string) *AnomalyReport {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.clock()
	cutoff := now.Add(-d.window)

	times := d.events[key]
	// prune old events
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	filtered = append(filtered, now)
	d.events[key] = filtered

	if len(filtered) >= d.flapThresh {
		return &AnomalyReport{
			Kind:       AnomalyFlapping,
			Key:        key,
			Count:      len(filtered),
			Window:     d.window,
			DetectedAt: now,
		}
	}
	return nil
}

// RecordBurst registers a batch of new port keys and returns a burst anomaly if threshold exceeded.
func (d *AnomalyDetector) RecordBurst(keys []string) *AnomalyReport {
	if len(keys) >= d.burstThresh {
		return &AnomalyReport{
			Kind:       AnomalyBurst,
			Key:        "batch",
			Count:      len(keys),
			Window:     d.window,
			DetectedAt: d.clock(),
		}
	}
	return nil
}

// Reset clears all recorded events for a key.
func (d *AnomalyDetector) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.events, key)
}
