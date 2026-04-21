package ports

import (
	"sync"
	"time"
)

// ScoreboardEntry holds the aggregated score for a single port key.
type ScoreboardEntry struct {
	Key       string
	Score     float64
	Hits      int
	LastSeen  time.Time
}

// Scoreboard accumulates weighted scores per port key, decaying over time.
type Scoreboard struct {
	mu       sync.Mutex
	entries  map[string]*ScoreboardEntry
	decayPer time.Duration
	decayFac float64
	clock    func() time.Time
}

// DefaultScoreboardDecay is the half-life window for score decay.
const DefaultScoreboardDecay = 5 * time.Minute

// NewScoreboard creates a Scoreboard with the given decay half-life.
func NewScoreboard(decayPer time.Duration, decayFactor float64) *Scoreboard {
	return newScoreboardWithClock(decayPer, decayFactor, time.Now)
}

func newScoreboardWithClock(decayPer time.Duration, decayFactor float64, clock func() time.Time) *Scoreboard {
	return &Scoreboard{
		entries:  make(map[string]*ScoreboardEntry),
		decayPer: decayPer,
		decayFac: decayFactor,
		clock:    clock,
	}
}

// Record adds weight to the score for the given key, applying time-based decay first.
func (s *Scoreboard) Record(key string, weight float64) {
	now := s.clock()
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[key]
	if !ok {
		s.entries[key] = &ScoreboardEntry{Key: key, Score: weight, Hits: 1, LastSeen: now}
		return
	}

	elapsed := now.Sub(e.LastSeen)
	if s.decayPer > 0 && elapsed > 0 {
		periods := float64(elapsed) / float64(s.decayPer)
		e.Score *= pow(s.decayFac, periods)
	}
	e.Score += weight
	e.Hits++
	e.LastSeen = now
}

// Get returns the current ScoreboardEntry for key, or nil if absent.
func (s *Scoreboard) Get(key string) *ScoreboardEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[key]
	if !ok {
		return nil
	}
	copy := *e
	return &copy
}

// Reset removes the entry for key.
func (s *Scoreboard) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// pow computes base^exp using repeated multiplication for small values.
func pow(base, exp float64) float64 {
	result := 1.0
	// Use logarithm-free approximation via math.Exp/Log is fine, but keep
	// the dependency minimal; use a simple loop for integer-ish periods.
	// For non-integer, fall back to a simple iterative approach.
	for exp >= 1.0 {
		result *= base
		exp -= 1.0
	}
	// fractional remainder: linear interpolation between 1 and base
	result *= 1.0 + exp*(base-1.0)
	return result
}
