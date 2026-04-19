package ports

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSnapshot_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")

	orig := &Snapshot{
		CapturedAt: time.Now().Truncate(time.Second),
		Entries: []Entry{
			{Protocol: "tcp", Port: 80, Process: "nginx"},
			{Protocol: "tcp", Port: 443, Process: "nginx"},
		},
	}

	if err := orig.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadSnapshot(path)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if len(loaded.Entries) != len(orig.Entries) {
		t.Fatalf("entry count: want %d, got %d", len(orig.Entries), len(loaded.Entries))
	}
	for i, e := range orig.Entries {
		if loaded.Entries[i] != e {
			t.Errorf("entry[%d]: want %+v, got %+v", i, e, loaded.Entries[i])
		}
	}
}

func TestLoadSnapshot_MissingFile(t *testing.T) {
	s, err := LoadSnapshot("/nonexistent/path/snap.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != nil {
		t.Fatal("expected nil snapshot for missing file")
	}
}

func TestSnapshot_SaveCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.json")

	s := &Snapshot{CapturedAt: time.Now(), Entries: []Entry{}}
	if err := s.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}
