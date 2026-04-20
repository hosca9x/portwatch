package ports

import (
	"os"
	"path/filepath"
	"testing"
)

func makeEntry(proto, addr string, port int) PortEntry {
	return PortEntry{Protocol: proto, Address: addr, Port: port}
}

func TestHistory_AddEntry(t *testing.T) {
	h := &History{}
	h.AddEntry("opened", makeEntry("tcp", "0.0.0.0", 8080))
	h.AddEntry("closed", makeEntry("udp", "0.0.0.0", 53))

	if len(h.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(h.Entries))
	}
	if h.Entries[0].Event != "opened" {
		t.Errorf("expected event 'opened', got %s", h.Entries[0].Event)
	}
	if h.Entries[1].Port.Port != 53 {
		t.Errorf("expected port 53, got %d", h.Entries[1].Port.Port)
	}
}

func TestHistory_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history.json")

	h := &History{}
	h.AddEntry("opened", makeEntry("tcp", "127.0.0.1", 443))

	if err := SaveHistory(path, h); err != nil {
		t.Fatalf("SaveHistory error: %v", err)
	}

	loaded, err := LoadHistory(path)
	if err != nil {
		t.Fatalf("LoadHistory error: %v", err)
	}
	if len(loaded.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(loaded.Entries))
	}
	if loaded.Entries[0].Port.Port != 443 {
		t.Errorf("expected port 443, got %d", loaded.Entries[0].Port.Port)
	}
}

func TestLoadHistory_MissingFile(t *testing.T) {
	h, err := LoadHistory("/nonexistent/path/history.json")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(h.Entries) != 0 {
		t.Errorf("expected empty history, got %d entries", len(h.Entries))
	}
}

func TestHistory_TimestampSet(t *testing.T) {
	h := &History{}
	h.AddEntry("opened", makeEntry("tcp", "0.0.0.0", 22))
	if h.Entries[0].Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
	_ = os.Getenv("") // suppress unused import warning
}
