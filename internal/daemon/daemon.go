package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/ports"
)

// Daemon orchestrates periodic port scanning, diffing, alerting, and metrics.
type Daemon struct {
	cfg       *config.Config
	watcher   *ports.Watcher
	notifier  *alert.Notifier
	collector *ports.Collector
}

// New creates a Daemon from the provided configuration.
func New(cfg *config.Config) *Daemon {
	return &Daemon{
		cfg:      cfg,
		watcher:  ports.NewWatcher(cfg),
		notifier: alert.NewNotifier(cfg),
		collector: ports.NewCollector(),
	}
}

// Metrics returns a snapshot of the current scan metrics.
func (d *Daemon) Metrics() ports.ScanMetrics {
	return d.collector.Snapshot()
}

// Run starts the daemon loop, blocking until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	log.Printf("portwatch: starting — interval %s", d.cfg.Interval)

	results := d.watcher.Watch(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("portwatch: shutting down")
			return nil

		case res, ok := <-results:
			if !ok {
				return nil
			}
			if res.Err != nil {
				log.Printf("portwatch: scan error: %v", res.Err)
				d.collector.RecordScan(0, 0, 0, res.Err)
				continue
			}

			newCount := len(res.Diff.Opened)
			closedCount := len(res.Diff.Closed)

			d.collector.RecordScan(len(res.Ports), newCount, closedCount, nil)

			if newCount > 0 || closedCount > 0 {
				if err := d.notifier.Notify(res.Diff); err != nil {
					log.Printf("portwatch: alert error: %v", err)
				}
			}

			log.Printf("portwatch: scan complete at %s — ports=%d new=%d closed=%d",
				time.Now().Format(time.RFC3339),
				len(res.Ports), newCount, closedCount)
		}
	}
}
