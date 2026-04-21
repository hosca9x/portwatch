package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// RetryEvent describes a retry lifecycle event for a scan operation.
type RetryEvent struct {
	Target    string
	Attempt   int
	MaxAttempt int
	Err       error
	Final     bool
	Timestamp time.Time
}

// RetryAlerter emits alerts when scan retries occur or are exhausted.
type RetryAlerter struct {
	out io.Writer
}

// NewRetryAlerter creates a RetryAlerter writing to w.
// If w is nil, os.Stdout is used.
func NewRetryAlerter(w io.Writer) *RetryAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &RetryAlerter{out: w}
}

// Notify emits a message for the given RetryEvent.
// Final events (all attempts exhausted) are marked as CRITICAL.
func (a *RetryAlerter) Notify(ev RetryEvent) {
	if ev.Err == nil {
		return
	}
	level := "WARN"
	if ev.Final {
		level = "CRITICAL"
	}
	ts := ev.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}
	fmt.Fprintf(
		a.out,
		"[%s] [%s] retry target=%s attempt=%d/%d err=%v\n",
		ts.Format(time.RFC3339),
		level,
		ev.Target,
		ev.Attempt,
		ev.MaxAttempt,
		ev.Err,
	)
}
