package ports

import (
	"encoding/json"
	"os"
	"time"
)

// HistoryEntry records a change event for a port.
type HistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Event     string    `json:"event"` // "opened" or "closed"
	Port      PortEntry `json:"port"`
}

// History holds a list of change events.
type History struct {
	Entries []HistoryEntry `json:"entries"`
}

// AddEntry appends a new event to the history.
func (h *History) AddEntry(event string, port PortEntry) {
	h.Entries = append(h.Entries, HistoryEntry{
		Timestamp: time.Now().UTC(),
		Event:     event,
		Port:      port,
	})
}

// SaveHistory writes the history to a JSON file at path.
func SaveHistory(path string, h *History) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(h)
}

// LoadHistory reads a history file from path.
// Returns an empty History if the file does not exist.
func LoadHistory(path string) (*History, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return &History{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var h History
	if err := json.NewDecoder(f).Decode(&h); err != nil {
		return nil, err
	}
	return &h, nil
}
