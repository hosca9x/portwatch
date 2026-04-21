package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/ports"
)

// Daemon orchestrates scanning, diffing, alerting, and debouncing.
type Daemon struct {
	cfg       *config.Config
	watcher   *ports.Watcher
	notifier  *alert.Notifier
	debouncer *ports.Debouncer
}

// New constructs a Daemon wired with the provided config.
func New(cfg *config.Config) *Daemon {
	notifier := alert.NewNotifier(nil)

	d := &Daemon{
		cfg:      cfg,
		notifier: notifier,
		watcher:  ports.NewWatcher(cfg),
	}

	d.debouncer = ports.NewDebouncer(
		time.Duration(cfg.DebounceSecs)*time.Second,
		func(key string) {
			log.Printf("[debounce] stable change confirmed for key: %s", key)
		},
	)

	return d
}

// Run starts the daemon loop, blocking until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	results := d.watcher.Watch(ctx)

	for {
		select {
		case res, ok := <-results:
			if !ok {
				return nil
			}
			if res.Err != nil {
				log.Printf("[daemon] scan error: %v", res.Err)
				continue
			}
			for _, e := range res.Diff.Opened {
				d.debouncer.Trigger(e.Key())
			}
			for _, e := range res.Diff.Closed {
				d.debouncer.Trigger(e.Key())
			}
			if err := d.notifier.Notify(res.Diff); err != nil {
				log.Printf("[daemon] notify error: %v", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}
