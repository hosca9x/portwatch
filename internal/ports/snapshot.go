package ports

import (
	"encoding/json"
	"os"
	"time"
)

// Entry represents a single open port at a point in time.
type Entry struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Process  string `json:"process,omitempty"`
}

// Snapshot holds a collection of open ports captured at a specific time.
type Snapshot struct {
	CapturedAt time.Time `json:"captured_at"`
	Entries    []Entry   `json:"entries"`
}

// Save writes the snapshot to the given file path as JSON.
func (s *Snapshot) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

// LoadSnapshot reads a snapshot from the given file path.
// Returns nil, nil if the file does not exist.
func LoadSnapshot(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	var s Snapshot
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}
