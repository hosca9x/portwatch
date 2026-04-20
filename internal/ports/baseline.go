package ports

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

// Baseline represents a trusted set of ports that are considered "known good".
// It is used to suppress alerts for ports that were present when the baseline was established.
type Baseline struct {
	CreatedAt time.Time   `json:"created_at"`
	Ports     []PortEntry `json:"ports"`
}

// NewBaseline creates a Baseline from the given snapshot entries.
func NewBaseline(entries []PortEntry) *Baseline {
	snap := make([]PortEntry, len(entries))
	copy(snap, entries)
	return &Baseline{
		CreatedAt: time.Now().UTC(),
		Ports:     snap,
	}
}

// Contains reports whether the given PortEntry is present in the baseline.
func (b *Baseline) Contains(e PortEntry) bool {
	for _, p := range b.Ports {
		if p.Key() == e.Key() {
			return true
		}
	}
	return false
}

// SaveBaseline writes the baseline to the given file path as JSON.
func SaveBaseline(path string, b *Baseline) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// LoadBaseline reads a baseline from the given file path.
// Returns nil and no error if the file does not exist.
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &b, nil
}
