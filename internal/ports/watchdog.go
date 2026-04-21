package ports

import (
	"sync"
	"time"
)

// WatchdogConfig holds configuration for the Watchdog.
type WatchdogConfig struct {
	// Timeout is how long without a successful tick before the watchdog fires.
	Timeout time.Duration
	// OnStall is called when the watchdog detects a stall.
	OnStall func(stalledFor time.Duration)
}

// DefaultWatchdogConfig returns a sensible default configuration.
func DefaultWatchdogConfig() WatchdogConfig {
	return WatchdogConfig{
		Timeout: 30 * time.Second,
		OnStall: func(_ time.Duration) {},
	}
}

// Watchdog monitors that a scan loop continues to make progress.
// If Tick is not called within Timeout, OnStall is invoked.
type Watchdog struct {
	cfg     WatchdogConfig
	clock   func() time.Time
	mu      sync.Mutex
	lastTick time.Time
	stop    chan struct{}
	wg      sync.WaitGroup
}

// NewWatchdog creates a Watchdog and starts its background monitor.
func NewWatchdog(cfg WatchdogConfig) *Watchdog {
	return newWatchdogWithClock(cfg, time.Now)
}

func newWatchdogWithClock(cfg WatchdogConfig, clock func() time.Time) *Watchdog {
	if cfg.OnStall == nil {
		cfg.OnStall = func(_ time.Duration) {}
	}
	w := &Watchdog{
		cfg:      cfg,
		clock:    clock,
		lastTick: clock(),
		stop:     make(chan struct{}),
	}
	w.wg.Add(1)
	go w.run()
	return w
}

// Tick signals that the scan loop is still alive.
func (w *Watchdog) Tick() {
	w.mu.Lock()
	w.lastTick = w.clock()
	w.mu.Unlock()
}

// Stop shuts down the watchdog monitor.
func (w *Watchdog) Stop() {
	close(w.stop)
	w.wg.Wait()
}

func (w *Watchdog) run() {
	defer w.wg.Done()
	ticker := time.NewTicker(w.cfg.Timeout / 2)
	defer ticker.Stop()
	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			w.mu.Lock()
			stalledFor := w.clock().Sub(w.lastTick)
			w.mu.Unlock()
			if stalledFor >= w.cfg.Timeout {
				w.cfg.OnStall(stalledFor)
			}
		}
	}
}
