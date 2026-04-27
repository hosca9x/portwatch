package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// BurstAlerter writes a human-readable alert line for each BurstReport.
type BurstAlerter struct {
	out io.Writer
}

// NewBurstAlerter creates a BurstAlerter. If out is nil, os.Stdout is used.
func NewBurstAlerter(out io.Writer) *BurstAlerter {
	if out == nil {
		out = os.Stdout
	}
	return &BurstAlerter{out: out}
}

// Notify writes an alert line if report is non-nil.
func (a *BurstAlerter) Notify(report *ports.BurstReport) {
	if report == nil {
		return
	}
	fmt.Fprintf(
		a.out,
		"[BURST] key=%s count=%d window=%s detected_at=%s\n",
		report.Key,
		report.Count,
		report.Window.String(),
		report.DetectedAt.Format(time.RFC3339),
	)
}

// NotifyAll writes alert lines for each non-nil report in the slice.
func (a *BurstAlerter) NotifyAll(reports []*ports.BurstReport) {
	for _, r := range reports {
		a.Notify(r)
	}
}
