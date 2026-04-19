package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/ports"
)

// Daemon periodically scans open ports and alerts on changes.
type Daemon struct {
	cfg      *config.Config
	scanner  *ports.Scanner
	notifier *alert.Notifier
	snapshotPath string
}

// New creates a new Daemon with the given config.
func New(cfg *config.Config, snapshotPath string) (*Daemon, error) {
	scanner := ports.NewScanner(cfg)
	notifier, err := alert.NewNotifier(cfg)
	if err != nil {
		return nil, err
	}
	return &Daemon{
		cfg:          cfg,
		scanner:      scanner,
		notifier:     notifier,
		snapshotPath: snapshotPath,
	}, nil
}

// Run starts the daemon loop, blocking until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	interval := time.Duration(d.cfg.IntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("portwatch daemon started (interval: %s)", interval)

	// Run once immediately on start.
	if err := d.tick(); err != nil {
		log.Printf("scan error: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := d.tick(); err != nil {
				log.Printf("scan error: %v", err)
			}
		case <-ctx.Done():
			log.Println("portwatch daemon stopped")
			return ctx.Err()
		}
	}
}

func (d *Daemon) tick() error {
	prev, err := ports.LoadSnapshot(d.snapshotPath)
	if err != nil {
		return err
	}

	current, err := d.scanner.Scan()
	if err != nil {
		return err
	}

	diff := ports.Diff(prev, current)
	d.notifier.Notify(diff)

	return current.Save(d.snapshotPath)
}
