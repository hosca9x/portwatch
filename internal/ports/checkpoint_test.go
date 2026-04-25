package ports

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

var fixedCheckpointClock = func() time.Time {
	return time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
}

func makeCheckpointEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto}
}

func TestCheckpoint_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "checkpoint.json")
	policy := DefaultCheckpointPolicy()

	ct := newCheckpointTrackerWithClock(path, policy, fixedCheckpointClock)
	entries := []PortEntry{
		makeCheckpointEntry(80, "tcp"),
		makeCheckpointEntry(443, "tcp"),
	}

	if err := ct.Save(entries); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := ct.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(got) != len(entries) {
		t.Fatalf("expected %d entries, got %d", len(entries), len(got))
	}
	if got[0].Port != 80 || got[1].Port != 443 {
		t.Errorf("unexpected entries: %+v", got)
	}
}

func TestCheckpoint_MissingFileReturnsNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no_such_file.json")
	ct := newCheckpointTrackerWithClock(path, DefaultCheckpointPolicy(), fixedCheckpointClock)

	got, err := ct.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestCheckpoint_StaleRecordReturnsNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "checkpoint.json")
	policy := CheckpointPolicy{MaxAge: time.Hour}

	// Save using a clock set 2 hours in the past.
	pastClock := func() time.Time { return fixedCheckpointClock().Add(-2 * time.Hour) }
	ct := newCheckpointTrackerWithClock(path, policy, pastClock)
	if err := ct.Save([]PortEntry{makeCheckpointEntry(22, "tcp")}); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Load using the fixed (current) clock — record is stale.
	ct2 := newCheckpointTrackerWithClock(path, policy, fixedCheckpointClock)
	got, err := ct2.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for stale record, got %+v", got)
	}
}

func TestCheckpoint_SaveCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "checkpoint.json")
	// Ensure parent exists so WriteFile can create the file.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	ct := newCheckpointTrackerWithClock(path, DefaultCheckpointPolicy(), fixedCheckpointClock)
	if err := ct.Save(nil); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist: %v", err)
	}
}
