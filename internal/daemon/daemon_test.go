package daemon

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

func defaultCfg() *config.Config {
	cfg := config.Default()
	cfg.IntervalSeconds = 1
	return cfg
}

func TestNew_ReturnsNonNil(t *testing.T) {
	cfg := defaultCfg()
	d, err := New(cfg, "/tmp/portwatch_test_snap.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil daemon")
	}
}

func TestRun_CancelsCleanly(t *testing.T) {
	cfg := defaultCfg()
	cfg.IntervalSeconds = 60 // long interval so only initial tick fires

	tmp, err := os.CreateTemp("", "portwatch-snap-*.json")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	defer os.Remove(tmp.Name())

	d, err := New(cfg, tmp.Name())
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = d.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}
