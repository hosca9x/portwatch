package ports

import (
	"sync"
	"time"
)

// SequencePolicy controls how the sequence tracker behaves.
type SequencePolicy struct {
	// MaxGap is the maximum time between events before the sequence resets.
	MaxGap time.Duration
	// MinLength is the minimum number of consecutive events to form a sequence.
	MinLength int
}

// DefaultSequencePolicy returns a sensible default policy.
func DefaultSequencePolicy() SequencePolicy {
	return SequencePolicy{
		MaxGap:    30 * time.Second,
		MinLength: 3,
	}
}

// SequenceReport describes a detected port-scan sequence.
type SequenceReport struct {
	Key    string
	Ports  []int
	Length int
	First  time.Time
	Last   time.Time
}

type seqState struct {
	ports []int
	times []time.Time
}

// SequenceTracker detects sequential port-access patterns that may indicate a scan.
type SequenceTracker struct {
	mu     sync.Mutex
	policy SequencePolicy
	state  map[string]*seqState
	clock  func() time.Time
}

// NewSequenceTracker creates a SequenceTracker with the given policy.
func NewSequenceTracker(p SequencePolicy) *SequenceTracker {
	return newSequenceTrackerWithClock(p, time.Now)
}

func newSequenceTrackerWithClock(p SequencePolicy, clock func() time.Time) *SequenceTracker {
	return &SequenceTracker{
		policy: p,
		state:  make(map[string]*seqState),
		clock:  clock,
	}
}

// Record registers a port access for the given key (e.g. source IP).
// Returns a SequenceReport if a sequence of MinLength is detected, otherwise nil.
func (s *SequenceTracker) Record(key string, port int) *SequenceReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	st, ok := s.state[key]
	if !ok {
		st = &seqState{}
		s.state[key] = st
	}

	// Reset if gap exceeded
	if len(st.times) > 0 && now.Sub(st.times[len(st.times)-1]) > s.policy.MaxGap {
		st.ports = nil
		st.times = nil
	}

	st.ports = append(st.ports, port)
	st.times = append(st.times, now)

	if len(st.ports) >= s.policy.MinLength {
		report := &SequenceReport{
			Key:    key,
			Ports:  append([]int(nil), st.ports...),
			Length: len(st.ports),
			First:  st.times[0],
			Last:   now,
		}
		// Reset after reporting
		st.ports = nil
		st.times = nil
		return report
	}
	return nil
}

// Reset clears the tracked state for a key.
func (s *SequenceTracker) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.state, key)
}
