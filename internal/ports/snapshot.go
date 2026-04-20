package ports

import (
	"encoding/json"
	"errors"
	"os"
)

// SaveSnapshot persists entries to path as JSON, creating the file if needed.
func SaveSnapshot(path string, entries []Entry) error {
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadSnapshot reads a previously saved snapshot from path.
// Returns nil slice (no error) when the file does not exist.
func LoadSnapshot(path string) ([]Entry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}
