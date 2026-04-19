package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

func writeTmp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "portwatch-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatal(err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestDefault(t *testing.T) {
	cfg := config.Default()
	if cfg.Interval != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.Interval)
	}
	if len(cfg.Protocols) != 2 {
		t.Errorf("expected 2 protocols, got %d", len(cfg.Protocols))
	}
}

func TestLoad_ValidFile(t *testing.T) {
	path := writeTmp(t, "interval: 1m\nalert_file: /tmp/alerts.log\nignore: [22, 80]\nprotocols: [tcp]\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != time.Minute {
		t.Errorf("expected 1m, got %v", cfg.Interval)
	}
	if cfg.AlertFile != "/tmp/alerts.log" {
		t.Errorf("unexpected alert_file: %s", cfg.AlertFile)
	}
	if len(cfg.Ignore) != 2 {
		t.Errorf("expected 2 ignored ports, got %d", len(cfg.Ignore))
	}
	if len(cfg.Protocols) != 1 || cfg.Protocols[0] != "tcp" {
		t.Errorf("unexpected protocols: %v", cfg.Protocols)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/portwatch.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoad_EmptyProtocolsFallback(t *testing.T) {
	path := writeTmp(t, "interval: 10s\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Protocols) != 2 {
		t.Errorf("expected fallback protocols, got %v", cfg.Protocols)
	}
}
