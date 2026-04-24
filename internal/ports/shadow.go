package ports

import (
	"sync"
	"time"
)

// ShadowEntry records a port that was seen briefly and then disappeared.
type ShadowEntry struct {
	Key       string
	Port      int
	Proto     string
	FirstSeen time.Time
	LastSeen  time.Time
	SeenCount int
}

// ShadowConfig controls shadow-port detection behaviour.
type ShadowConfig struct {
	// MaxLifetime is the window within which brief appearances are tracked.
	MaxLifetime time.Duration
	// MinAppearances is the minimum number of times a port must appear to NOT
	// be considered a shadow port.
	MinAppearances int
}

// DefaultShadowConfig returns sensible defaults.
func DefaultShadowConfig() ShadowConfig {
	return ShadowConfig{
		MaxLifetime:    5 * time.Minute,
		MinAppearances: 3,
	}
}

// ShadowTracker detects ports that appear only transiently.
type ShadowTracker struct {
	mu      sync.Mutex
	cfg     ShadowConfig
	entries map[string]*ShadowEntry
	clock   func() time.Time
}

// NewShadowTracker creates a ShadowTracker with the given config.
func NewShadowTracker(cfg ShadowConfig) *ShadowTracker {
	return newShadowTrackerWithClock(cfg, time.Now)
}

func newShadowTrackerWithClock(cfg ShadowConfig, clock func() time.Time) *ShadowTracker {
	return &ShadowTracker{
		cfg:     cfg,
		entries: make(map[string]*ShadowEntry),
		clock:   clock,
	}
}

// Record registers an observation of a port entry.
func (s *ShadowTracker) Record(e PortEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	if se, ok := s.entries[e.Key()]; ok {
		se.LastSeen = now
		se.SeenCount++
	} else {
		s.entries[e.Key()] = &ShadowEntry{
			Key:       e.Key(),
			Port:      e.Port,
			Proto:     e.Proto,
			FirstSeen: now,
			LastSeen:  now,
			SeenCount: 1,
		}
	}
}

// Shadows returns entries that appeared fewer than MinAppearances times
// within MaxLifetime and have not been seen recently.
func (s *ShadowTracker) Shadows() []ShadowEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	var out []ShadowEntry
	for k, se := range s.entries {
		age := now.Sub(se.FirstSeen)
		if age >= s.cfg.MaxLifetime && se.SeenCount < s.cfg.MinAppearances {
			out = append(out, *se)
			delete(s.entries, k)
		}
	}
	return out
}

// Evict removes stale entries that have exceeded the tracking window.
func (s *ShadowTracker) Evict() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.clock()
	for k, se := range s.entries {
		if now.Sub(se.FirstSeen) >= s.cfg.MaxLifetime {
			delete(s.entries, k)
		}
	}
}
