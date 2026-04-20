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

// tempSnapshotFile creates a temporary file for use as a snapshot path in
// tests and returns its name along with a cleanup function.
func tempSnapshotFile(t *testing.T) (string, func()) {
	t.Helper()
	tmp, err := os.CreateTemp("", "portwatch-snap-*.json")
	if err != nil {
		t.Fatal(err)
	}
	tmp.Close()
	return tmp.Name(), func() { os.Remove(tmp.Name()) }
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

	snap, cleanup := tempSnapshotFile(t)
	defer cleanup()

	d, err := New(cfg, snap)
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
