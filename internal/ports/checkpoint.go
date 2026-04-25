package ports

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// CheckpointPolicy controls checkpoint behaviour.
type CheckpointPolicy struct {
	// MaxAge is how long a checkpoint is considered valid.
	MaxAge time.Duration
}

// DefaultCheckpointPolicy returns sensible defaults.
func DefaultCheckpointPolicy() CheckpointPolicy {
	return CheckpointPolicy{
		MaxAge: 24 * time.Hour,
	}
}

// CheckpointRecord is persisted to disk.
type CheckpointRecord struct {
	SavedAt  time.Time  `json:"saved_at"`
	Entries  []PortEntry `json:"entries"`
}

// CheckpointTracker manages periodic on-disk checkpoints of port state.
type CheckpointTracker struct {
	mu     sync.Mutex
	policy CheckpointPolicy
	clock  func() time.Time
	path   string
}

// NewCheckpointTracker creates a CheckpointTracker backed by the given file path.
func NewCheckpointTracker(path string, policy CheckpointPolicy) *CheckpointTracker {
	return newCheckpointTrackerWithClock(path, policy, time.Now)
}

func newCheckpointTrackerWithClock(path string, policy CheckpointPolicy, clock func() time.Time) *CheckpointTracker {
	return &CheckpointTracker{path: path, policy: policy, clock: clock}
}

// Save writes the current entries to the checkpoint file.
func (c *CheckpointTracker) Save(entries []PortEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	rec := CheckpointRecord{
		SavedAt: c.clock(),
		Entries: entries,
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o644)
}

// Load reads the checkpoint file and returns entries if the record is still
// within MaxAge. Returns nil, nil when the file is missing or stale.
func (c *CheckpointTracker) Load() ([]PortEntry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var rec CheckpointRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, err
	}

	if c.clock().Sub(rec.SavedAt) > c.policy.MaxAge {
		return nil, nil
	}
	return rec.Entries, nil
}
