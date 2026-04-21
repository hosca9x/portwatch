package ports

import (
	"math/rand"
	"time"
)

// JitterPolicy controls how randomised delay is applied to scan intervals.
type JitterPolicy struct {
	// MaxFraction is the maximum fraction of the base duration to add as jitter.
	// For example, 0.25 means up to 25% of the base duration is added.
	MaxFraction float64
}

// DefaultJitterPolicy returns a JitterPolicy with sensible defaults.
func DefaultJitterPolicy() JitterPolicy {
	return JitterPolicy{
		MaxFraction: 0.20,
	}
}

// Jitterer applies randomised jitter to durations so that concurrent daemons
// do not all scan at exactly the same instant (thundering-herd mitigation).
type Jitterer struct {
	policy JitterPolicy
	rng    *rand.Rand
}

// NewJitterer creates a Jitterer using the given policy.
func NewJitterer(policy JitterPolicy) *Jitterer {
	return &Jitterer{
		policy: policy,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Apply returns base plus a random additional duration in
// [0, base * MaxFraction).
func (j *Jitterer) Apply(base time.Duration) time.Duration {
	if j.policy.MaxFraction <= 0 || base <= 0 {
		return base
	}
	maxExtra := float64(base) * j.policy.MaxFraction
	extra := time.Duration(j.rng.Float64() * maxExtra)
	return base + extra
}

// Reset re-seeds the internal RNG. Useful in tests or after a fork.
func (j *Jitterer) Reset(seed int64) {
	j.rng = rand.New(rand.NewSource(seed))
}
