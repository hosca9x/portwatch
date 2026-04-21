package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/ports"
)

// Daemon orchestrates periodic port scanning, diffing, alerting, and throttling.
type Daemon struct {
	cfg       config.Config
	watcher   *ports.Watcher
	notifier  *alert.Notifier
	throttler *ports.Throttler
}

// New constructs a Daemon from the provided config.
func New(cfg config.Config) *Daemon {
	watcher := ports.NewWatcher(cfg)
	notifier := alert.NewNotifier(nil)
	throttler := ports.NewThrottler(ports.ThrottleConfig{
		MaxScansPerMinute: cfg.MaxScansPerMinute,
		BurstSize:         cfg.BurstSize,
	}, nil)
	return &Daemon{
		cfg:       cfg,
		watcher:   watcher,
		notifier:  notifier,
		throttler: throttler,
	}
}

// Run starts the daemon loop. It blocks until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Duration(d.cfg.IntervalSeconds) * time.Second)
	defer ticker.Stop()

	log.Println("portwatch daemon started")

	for {
		select {
		case <-ctx.Done():
			log.Println("portwatch daemon stopping")
			return nil
		case <-ticker.C:
			if !d.throttler.Allow() {
				log.Println("scan throttled, skipping interval")
				continue
			}
			results, err := d.watcher.Scan(ctx)
			if err != nil {
				log.Printf("scan error: %v", err)
				continue
			}
			if err := d.notifier.Notify(results); err != nil {
				log.Printf("notify error: %v", err)
			}
		}
	}
}
