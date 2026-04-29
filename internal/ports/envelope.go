package ports

import "time"

// EnvelopePolicy controls envelope wrapping behaviour.
type EnvelopePolicy struct {
	// MaxAge is the maximum age of an envelope before it is considered stale.
	MaxAge time.Duration
	// Source is an optional label identifying the originating scanner instance.
	Source string
}

// DefaultEnvelopePolicy returns sensible defaults.
func DefaultEnvelopePolicy() EnvelopePolicy {
	return EnvelopePolicy{
		MaxAge: 5 * time.Minute,
		Source: "portwatch",
	}
}

// Envelope wraps a slice of PortEntry values with metadata produced at scan
// time so downstream consumers can make staleness decisions without needing
// access to the scanner itself.
type Envelope struct {
	Source    string
	ScannedAt time.Time
	Entries   []PortEntry
}

// IsStale reports whether the envelope was produced longer ago than MaxAge.
func (e *Envelope) IsStale(policy EnvelopePolicy, now time.Time) bool {
	return now.Sub(e.ScannedAt) > policy.MaxAge
}

// EnvelopeWrapper creates Envelope values from raw scan results.
type EnvelopeWrapper struct {
	policy EnvelopePolicy
	clock  func() time.Time
}

// NewEnvelopeWrapper returns an EnvelopeWrapper using the supplied policy.
func NewEnvelopeWrapper(policy EnvelopePolicy) *EnvelopeWrapper {
	return newEnvelopeWrapperWithClock(policy, time.Now)
}

func newEnvelopeWrapperWithClock(policy EnvelopePolicy, clock func() time.Time) *EnvelopeWrapper {
	return &EnvelopeWrapper{policy: policy, clock: clock}
}

// Wrap bundles entries into an Envelope stamped with the current time.
func (w *EnvelopeWrapper) Wrap(entries []PortEntry) *Envelope {
	return &Envelope{
		Source:    w.policy.Source,
		ScannedAt: w.clock(),
		Entries:   entries,
	}
}

// Stale reports whether the given envelope has exceeded its maximum age.
func (w *EnvelopeWrapper) Stale(e *Envelope) bool {
	return e.IsStale(w.policy, w.clock())
}
