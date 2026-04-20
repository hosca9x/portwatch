package ports

import (
	"os"
	"path/filepath"
	"testing"
)

func baselineEntry(proto string, port int) PortEntry {
	return PortEntry{Protocol: proto, Port: port, PID: 1, Process: "test"}
}

func TestNewBaseline_StoresEntries(t *testing.T) {
	entries := []PortEntry{
		baselineEntry("tcp", 80),
		baselineEntry("tcp", 443),
	}
	b := NewBaseline(entries)
	if len(b.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(b.Ports))
	}
	if b.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestBaseline_Contains(t *testing.T) {
	entries := []PortEntry{
		baselineEntry("tcp", 8080),
	}
	b := NewBaseline(entries)

	if !b.Contains(baselineEntry("tcp", 8080)) {
		t.Error("expected baseline to contain tcp:8080")
	}
	if b.Contains(baselineEntry("tcp", 9090)) {
		t.Error("expected baseline NOT to contain tcp:9090")
	}
	if b.Contains(baselineEntry("udp", 8080)) {
		t.Error("expected baseline NOT to contain udp:8080 (different proto)")
	}
}

func TestBaseline_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "baseline.json")

	orig := NewBaseline([]PortEntry{
		baselineEntry("tcp", 22),
		baselineEntry("udp", 53),
	})

	if err := SaveBaseline(path, orig); err != nil {
		t.Fatalf("SaveBaseline: %v", err)
	}

	loaded, err := LoadBaseline(path)
	if err != nil {
		t.Fatalf("LoadBaseline: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil baseline")
	}
	if len(loaded.Ports) != 2 {
		t.Fatalf("expected 2 ports, got %d", len(loaded.Ports))
	}
	if !loaded.Contains(baselineEntry("tcp", 22)) {
		t.Error("loaded baseline missing tcp:22")
	}
}

func TestLoadBaseline_MissingFile(t *testing.T) {
	b, err := LoadBaseline("/nonexistent/path/baseline.json")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if b != nil {
		t.Error("expected nil baseline for missing file")
	}
}

func TestNewBaseline_IsolatesCopy(t *testing.T) {
	entries := []PortEntry{baselineEntry("tcp", 80)}
	b := NewBaseline(entries)
	entries[0].Port = 9999
	if b.Ports[0].Port == 9999 {
		t.Error("baseline should not share backing array with input slice")
	}
}

func TestBaseline_SaveCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new_baseline.json")
	b := NewBaseline(nil)
	if err := SaveBaseline(path, b); err != nil {
		t.Fatalf("SaveBaseline: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist: %v", err)
	}
}
