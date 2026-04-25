package ports

import (
	"sync"
	"time"
)

// HeatmapPolicy controls heatmap behaviour.
type HeatmapPolicy struct {
	// WindowSize is how far back observations are counted.
	WindowSize time.Duration
}

// DefaultHeatmapPolicy returns sensible defaults.
func DefaultHeatmapPolicy() HeatmapPolicy {
	return HeatmapPolicy{
		WindowSize: 10 * time.Minute,
	}
}

// HeatmapEntry records a single observation timestamp for a port key.
type HeatmapEntry struct {
	Key string
	At  time.Time
}

// HeatmapReport is returned by Snapshot.
type HeatmapReport struct {
	Key   string
	Count int
	Heat  float64 // observations per minute over the window
}

type heatmapClock func() time.Time

// Heatmap tracks observation frequency per port key within a sliding window.
type Heatmap struct {
	mu     sync.Mutex
	policy HeatmapPolicy
	clock  heatmapClock
	events map[string][]time.Time
}

// NewHeatmap creates a Heatmap with the given policy.
func NewHeatmap(p HeatmapPolicy) *Heatmap {
	return newHeatmapWithClock(p, time.Now)
}

func newHeatmapWithClock(p HeatmapPolicy, clk heatmapClock) *Heatmap {
	return &Heatmap{
		policy: p,
		clock:  clk,
		events: make(map[string][]time.Time),
	}
}

// Record registers an observation for the given key.
func (h *Heatmap) Record(key string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := h.clock()
	h.events[key] = append(h.prune(key, now), now)
}

// Snapshot returns a HeatmapReport for every key that has observations in the window.
func (h *Heatmap) Snapshot() []HeatmapReport {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := h.clock()
	minutes := h.policy.WindowSize.Minutes()
	var out []HeatmapReport
	for key := range h.events {
		times := h.prune(key, now)
		h.events[key] = times
		if len(times) == 0 {
			continue
		}
		heat := float64(len(times)) / minutes
		out = append(out, HeatmapReport{Key: key, Count: len(times), Heat: heat})
	}
	return out
}

// prune removes events outside the window; caller must hold the lock.
func (h *Heatmap) prune(key string, now time.Time) []time.Time {
	cutoff := now.Add(-h.policy.WindowSize)
	old := h.events[key]
	var fresh []time.Time
	for _, t := range old {
		if !t.Before(cutoff) {
			fresh = append(fresh, t)
		}
	}
	return fresh
}
