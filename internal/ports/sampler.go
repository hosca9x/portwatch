package ports

import (
	"sync"
	"time"
)

// SamplePolicy controls how the sampler retains scan results.
type SamplePolicy struct {
	MaxSamples int
	MaxAge     time.Duration
}

// DefaultSamplePolicy returns a sensible default sampling policy.
func DefaultSamplePolicy() SamplePolicy {
	return SamplePolicy{
		MaxSamples: 60,
		MaxAge:     10 * time.Minute,
	}
}

// Sample holds a single recorded scan result.
type Sample struct {
	At      time.Time
	Entries []PortEntry
}

// Sampler retains a rolling window of recent scan snapshots.
type Sampler struct {
	mu      sync.Mutex
	policy  SamplePolicy
	samples []Sample
	clock   func() time.Time
}

// NewSampler creates a Sampler with the given policy.
func NewSampler(policy SamplePolicy) *Sampler {
	return &Sampler{
		policy: policy,
		clock:  time.Now,
	}
}

func newSamplerWithClock(policy SamplePolicy, clock func() time.Time) *Sampler {
	return &Sampler{policy: policy, clock: clock}
}

// Record adds a new sample and evicts entries outside the window.
func (s *Sampler) Record(entries []PortEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock()
	s.samples = append(s.samples, Sample{At: now, Entries: entries})
	s.evict(now)
}

// Samples returns a copy of the current sample window.
func (s *Sampler) Samples() []Sample {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Sample, len(s.samples))
	copy(out, s.samples)
	return out
}

// Len returns the number of retained samples.
func (s *Sampler) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.samples)
}

func (s *Sampler) evict(now time.Time) {
	cutoff := now.Add(-s.policy.MaxAge)
	start := 0
	for start < len(s.samples) && s.samples[start].At.Before(cutoff) {
		start++
	}
	s.samples = s.samples[start:]
	if len(s.samples) > s.policy.MaxSamples {
		s.samples = s.samples[len(s.samples)-s.policy.MaxSamples:]
	}
}
