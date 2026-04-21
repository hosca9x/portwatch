package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// QuotaAlerter emits an alert when a port key exceeds its event quota
// within the configured rolling window.
type QuotaAlerter struct {
	tracker *ports.QuotaTracker
	out     io.Writer
}

// NewQuotaAlerter creates a QuotaAlerter using the provided policy.
// If out is nil it defaults to os.Stdout.
func NewQuotaAlerter(policy ports.QuotaPolicy, out io.Writer) *QuotaAlerter {
	if out == nil {
		out = os.Stdout
	}
	return &QuotaAlerter{
		tracker: ports.NewQuotaTracker(policy),
		out:     out,
	}
}

// Notify records the event for the given entry and writes an alert if the
// quota has been exceeded. It returns true when an alert was emitted.
func (a *QuotaAlerter) Notify(entry ports.PortEntry) bool {
	key := entry.Key()
	if a.tracker.Record(key) {
		count := a.tracker.Count(key)
		fmt.Fprintf(
			a.out,
			"[QUOTA] %s port %d/%s exceeded event quota (count=%d) at %s\n",
			"ALERT",
			entry.Port,
			entry.Proto,
			count,
			time.Now().Format(time.RFC3339),
		)
		return true
	}
	return false
}

// Reset clears the quota state for a specific key.
func (a *QuotaAlerter) Reset(key string) {
	a.tracker.Reset(key)
}
