package ports

import (
	"sync"
	"time"
)

// DecayTracker tracks how long a port has been continuously open and
// assigns a decayed weight that decreases over time, useful for
// distinguishing long-lived stable ports from newly opened ones.
type DecayTracker struct {
	mu      sync.Mutex
	firstSeen map[string]time.Time
	clock     func() time.Time
	halfLife  time.Duration
}

// DecayReport holds the result of a decay calculation for a port entry.
type DecayReport struct {
	Key      string
	Age      time.Duration
	Weight   float64 // 1.0 = brand new, approaches 0 as port ages
	IsStable bool    // true when weight drops below stability threshold
}

const defaultHalfLife = 30 * time.Minute
const stabilityThreshold = 0.25

// NewDecayTracker returns a DecayTracker with the given half-life duration.
// If halfLife is zero, a default of 30 minutes is used.
func NewDecayTracker(halfLife time.Duration) *DecayTracker {
	if halfLife == 0 {
		halfLife = defaultHalfLife
	}
	return &DecayTracker{
		firstSeen: make(map[string]time.Time),
		clock:     time.Now,
		halfLife:  halfLife,
	}
}

// Observe records the first time a port key is seen. Subsequent calls
// for the same key do not reset the first-seen timestamp.
func (d *DecayTracker) Observe(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.firstSeen[key]; !exists {
		d.firstSeen[key] = d.clock()
	}
}

// Report returns a DecayReport for the given key. If the key has never
// been observed, weight is 1.0 and age is 0.
func (d *DecayTracker) Report(key string) DecayReport {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.clock()
	fs, ok := d.firstSeen[key]
	if !ok {
		return DecayReport{Key: key, Age: 0, Weight: 1.0, IsStable: false}
	}

	age := now.Sub(fs)
	// Exponential decay: weight = 2^(-age/halfLife)
	halves := float64(age) / float64(d.halfLife)
	weight := 1.0
	for i := 0.0; i < halves; i++ {
		weight *= 0.5
	}
	return DecayReport{
		Key:      key,
		Age:      age,
		Weight:   weight,
		IsStable: weight < stabilityThreshold,
	}
}

// Forget removes a key from the tracker, e.g. when a port is closed.
func (d *DecayTracker) Forget(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.firstSeen, key)
}
