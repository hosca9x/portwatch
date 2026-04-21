package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// TrendAlerter emits a human-readable alert when a TrendReport indicates
// non-stable port-count movement.
type TrendAlerter struct {
	out   io.Writer
	clock func() time.Time
}

// NewTrendAlerter creates a TrendAlerter writing to w.
// Pass nil to default to os.Stdout.
func NewTrendAlerter(w io.Writer) *TrendAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &TrendAlerter{out: w, clock: time.Now}
}

// Notify writes a trend alert to the configured writer if the report
// direction is not stable. It returns true when an alert was emitted.
func (ta *TrendAlerter) Notify(r ports.TrendReport) bool {
	if r.Direction == ports.TrendStable {
		return false
	}

	sign := "+"
	if r.Delta < 0 {
		sign = ""
	}

	fmt.Fprintf(
		ta.out,
		"[%s] TREND %s: open-port count changed by %s%d (window=%d samples)\n",
		ta.clock().Format(time.RFC3339),
		r.Direction,
		sign,
		r.Delta,
		len(r.Samples),
	)
	return true
}
