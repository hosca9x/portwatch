package ports_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/ports"
)

func defaultWatcherCfg() config.Config {
	cfg := config.Default()
	cfg.Protocols = []string{"tcp"}
	return cfg
}

func TestWatcher_EmitsResults(t *testing.T) {
	snap := tmpFile(t)

	cfg := defaultWatcherCfg()
	scanner := ports.NewScanner(cfg)
	filter := ports.NewFilter(cfg)
	watcher := ports.NewWatcher(scanner, filter, 50*time.Millisecond, snap)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	ch := watcher.Watch(ctx)

	var results []ports.WatchResult
	for r := range ch {
		results = append(results, r)
	}

	if len(results) == 0 {
		t.Fatal("expected at least one WatchResult, got none")
	}
	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result[%d] unexpected error: %v", i, r.Err)
		}
		if r.Diff == nil {
			t.Errorf("result[%d] Diff is nil", i)
		}
	}
}

func TestWatcher_ChannelClosedOnCancel(t *testing.T) {
	snap := tmpFile(t)

	cfg := defaultWatcherCfg()
	watcher := ports.NewWatcher(ports.NewScanner(cfg), ports.NewFilter(cfg), 30*time.Millisecond, snap)

	ctx, cancel := context.WithCancel(context.Background())
	ch := watcher.Watch(ctx)
	cancel()

	select {
	case _, ok := <-ch:
		if ok {
			// drain remaining
			for range ch {
			}
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("channel was not closed after context cancellation")
	}
}

func tmpFile(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "snap-*.json")
	if err != nil {
		t.Fatalf("tmpFile: %v", err)
	}
	f.Close()
	return f.Name()
}
