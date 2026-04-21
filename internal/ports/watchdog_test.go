package ports

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestWatchdog_NoStallWhenTickedFrequently(t *testing.T) {
	var stallCount int32
	cfg := WatchdogConfig{
		Timeout: 100 * time.Millisecond,
		OnStall: func(_ time.Duration) { atomic.AddInt32(&stallCount, 1) },
	}
	now := time.Now()
	clock := func() time.Time { return now }
	w := newWatchdogWithClock(cfg, clock)
	defer w.Stop()

	// Tick immediately — no stall should fire.
	w.Tick()
	time.Sleep(20 * time.Millisecond)

	if atomic.LoadInt32(&stallCount) != 0 {
		t.Errorf("expected no stall, got %d", stallCount)
	}
}

func TestWatchdog_FiresOnStall(t *testing.T) {
	var stallCount int32
	var stalledDur time.Duration

	timeout := 40 * time.Millisecond
	cfg := WatchdogConfig{
		Timeout: timeout,
		OnStall: func(d time.Duration) {
			atomic.AddInt32(&stallCount, 1)
			stalledDur = d
		},
	}

	base := time.Now()
	// Advance clock past the timeout immediately so the monitor fires.
	offset := timeout + 10*time.Millisecond
	clock := func() time.Time { return base.Add(offset) }

	w := newWatchdogWithClock(cfg, clock)
	defer w.Stop()

	// Wait long enough for the background ticker (timeout/2) to fire at least once.
	time.Sleep(timeout)

	if atomic.LoadInt32(&stallCount) == 0 {
		t.Error("expected stall callback to fire")
	}
	if stalledDur < timeout {
		t.Errorf("expected stalledDur >= %v, got %v", timeout, stalledDur)
	}
}

func TestWatchdog_TickResetsStall(t *testing.T) {
	var stallCount int32
	cfg := WatchdogConfig{
		Timeout: 40 * time.Millisecond,
		OnStall: func(_ time.Duration) { atomic.AddInt32(&stallCount, 1) },
	}
	w := NewWatchdog(cfg)
	defer w.Stop()

	// Keep ticking to prevent stall.
	for i := 0; i < 5; i++ {
		w.Tick()
		time.Sleep(5 * time.Millisecond)
	}

	if atomic.LoadInt32(&stallCount) != 0 {
		t.Errorf("expected no stall while ticking, got %d", stallCount)
	}
}

func TestDefaultWatchdogConfig_Timeout(t *testing.T) {
	cfg := DefaultWatchdogConfig()
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", cfg.Timeout)
	}
	if cfg.OnStall == nil {
		t.Error("expected non-nil OnStall")
	}
}
