package ports

import (
	"context"
	"log"
	"time"
)

// WatchResult holds the outcome of a single scan cycle.
type WatchResult struct {
	Snapshot []Entry
	Diff     *DiffResult
	Err      error
}

// Watcher periodically scans ports and emits diffs against the previous snapshot.
type Watcher struct {
	scanner  *Scanner
	filter   *Filter
	interval time.Duration
	snapshotPath string
}

// NewWatcher constructs a Watcher with the given scanner, filter, poll interval,
// and path used to persist the previous snapshot between cycles.
func NewWatcher(scanner *Scanner, filter *Filter, interval time.Duration, snapshotPath string) *Watcher {
	return &Watcher{
		scanner:      scanner,
		filter:       filter,
		interval:     interval,
		snapshotPath: snapshotPath,
	}
}

// Watch runs the scan loop, sending WatchResult values on the returned channel
// until ctx is cancelled. The channel is closed on exit.
func (w *Watcher) Watch(ctx context.Context) <-chan WatchResult {
	ch := make(chan WatchResult, 1)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				result := w.scan()
				select {
				case ch <- result:
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return ch
}

func (w *Watcher) scan()	current, err := w.scanner.Scan()
	if err != nil {
		log.Printf("wdr := Diff(prev, filtered)

	if err := SaveSnapshot(w.snapshotPath, filtered); err != nil {
		log.Printf("watcher: failed to save snapshot: %v", err)
	}

	return WatchResult{Snapshot: filtered, Diff: dr}
}
