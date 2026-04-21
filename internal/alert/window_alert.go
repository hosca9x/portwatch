package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// WindowAlert emits an alert when a port's event count exceeds the window
// policy threshold within the configured sliding window.
type WindowAlert struct {
	writer  io.Writer
	counter *ports.WindowCounter
}

// NewWindowAlerter creates a WindowAlert writing to w.
// If w is nil, os.Stdout is used.
func NewWindowAlerter(w io.Writer, policy ports.WindowPolicy) *WindowAlert {
	if w == nil {
		w = os.Stdout
	}
	return &WindowAlert{
		writer:  w,
		counter: ports.NewWindowCounter(policy),
	}
}

// WindowEvent carries the port entry and the current count within the window.
type WindowEvent struct {
	Entry ports.PortEntry
	Count int
	Limit int
	At    time.Time
}

// Observe records an event for the entry and emits an alert if the threshold
// is exceeded, returning true when an alert was written.
func (wa *WindowAlert) Observe(entry ports.PortEntry) bool {
	key := entry.Key()
	count := wa.counter.Add(key)
	if !wa.counter.Exceeded(key) {
		return false
	}
	wa.emit(WindowEvent{
		Entry: entry,
		Count: count,
		At:    time.Now(),
	})
	return true
}

// Reset clears the window state for the given entry.
func (wa *WindowAlert) Reset(entry ports.PortEntry) {
	wa.counter.Reset(entry.Key())
}

func (wa *WindowAlert) emit(ev WindowEvent) {
	fmt.Fprintf(
		wa.writer,
		"[WINDOW-ALERT] %s port %s/%d exceeded rate limit (%d events) at %s\n",
		ev.Entry.Proto,
		ev.Entry.Proto,
		ev.Entry.Port,
		ev.Count,
		ev.At.Format(time.RFC3339),
	)
}
