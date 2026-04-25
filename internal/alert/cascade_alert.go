package alert

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// CascadeAlerter emits a human-readable alert when a cascade of newly-opened
// ports is detected within a short time window.
type CascadeAlerter struct {
	w io.Writer
}

// NewCascadeAlerter creates a CascadeAlerter that writes to w.
// If w is nil it defaults to os.Stdout.
func NewCascadeAlerter(w io.Writer) *CascadeAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &CascadeAlerter{w: w}
}

// Notify writes an alert for the given CascadeReport.
// If report is nil no output is produced.
func (a *CascadeAlerter) Notify(report *ports.CascadeReport) {
	if report == nil || len(report.Ports) == 0 {
		return
	}

	keys := make([]string, 0, len(report.Ports))
	for _, p := range report.Ports {
		keys = append(keys, fmt.Sprintf("%s/%d", p.Proto, p.Port))
	}

	fmt.Fprintf(
		a.w,
		"[CASCADE] %s — %d ports opened: %s\n",
		report.DetectedAt.Format("15:04:05"),
		len(report.Ports),
		strings.Join(keys, ", "),
	)
}
