package ports

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"sync"
)

// Fingerprint represents a SHA-256 digest of a port snapshot.
type Fingerprint string

// FingerprintTracker computes and compares snapshot fingerprints to detect
// whether the set of open ports has changed between scans.
type FingerprintTracker struct {
	mu   sync.Mutex
	last Fingerprint
}

// NewFingerprintTracker returns a new FingerprintTracker with no prior state.
func NewFingerprintTracker() *FingerprintTracker {
	return &FingerprintTracker{}
}

// Compute derives a deterministic fingerprint from a slice of PortEntry values.
// Entries are sorted by key before hashing so order does not matter.
func Compute(entries []PortEntry) Fingerprint {
	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		keys = append(keys, fmt.Sprintf("%s:%d", e.Proto, e.Port))
	}
	sort.Strings(keys)

	h := sha256.New()
	for _, k := range keys {
		_, _ = fmt.Fprintln(h, k)
	}
	return Fingerprint(hex.EncodeToString(h.Sum(nil)))
}

// Changed returns true if the fingerprint of entries differs from the last
// recorded fingerprint, and updates the stored fingerprint accordingly.
func (ft *FingerprintTracker) Changed(entries []PortEntry) bool {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	current := Compute(entries)
	if current == ft.last {
		return false
	}
	ft.last = current
	return true
}

// Current returns the most recently recorded fingerprint.
func (ft *FingerprintTracker) Current() Fingerprint {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	return ft.last
}

// Reset clears the stored fingerprint, causing the next call to Changed to
// always return true.
func (ft *FingerprintTracker) Reset() {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.last = ""
}
