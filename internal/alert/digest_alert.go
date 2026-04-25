package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// DigestAlert emits a line whenever the digest of the observed port set changes.
type DigestAlert struct {
	w       io.Writer
	tracker *ports.DigestTracker
	lastDigest string
}

// NewDigestAlerter creates a DigestAlert writing to w.
// If w is nil, os.Stdout is used.
func NewDigestAlerter(w io.Writer, policy ports.DigestPolicy) *DigestAlert {
	if w == nil {
		w = os.Stdout
	}
	return &DigestAlert{
		w:       w,
		tracker: ports.NewDigestTracker(policy),
	}
}

// Notify checks whether the digest of entries has changed since the last call.
// If it has, an alert line is written to the configured writer.
func (d *DigestAlert) Notify(label string, entries []ports.PortEntry) {
	current := d.tracker.Compute(label, entries)
	if current == d.lastDigest {
		return
	}
	prev := d.lastDigest
	if prev == "" {
		prev = "(none)"
	}
	fmt.Fprintf(d.w, "[%s] digest changed label=%s prev=%s current=%s\n",
		time.Now().UTC().Format(time.RFC3339), label, prev, current)
	d.lastDigest = current
}

// Reset clears the last-seen digest so the next Notify always emits.
func (d *DigestAlert) Reset() {
	d.lastDigest = ""
	d.tracker.Invalidate("")
}
