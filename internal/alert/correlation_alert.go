package alert

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// CorrelationAlerter emits alerts when a CorrelationReport contains
// multiple related port events within a short window.
type CorrelationAlerter struct {
	out      io.Writer
	minGroup int
}

// NewCorrelationAlerter returns a CorrelationAlerter that writes to out.
// minGroup sets the minimum number of events required to emit an alert.
// If out is nil, os.Stdout is used.
func NewCorrelationAlerter(out io.Writer, minGroup int) *CorrelationAlerter {
	if out == nil {
		out = os.Stdout
	}
	if minGroup < 1 {
		minGroup = 1
	}
	return &CorrelationAlerter{out: out, minGroup: minGroup}
}

// Notify emits a formatted alert if the report meets the minimum group size.
// It is a no-op when report is nil or below threshold.
func (a *CorrelationAlerter) Notify(report *ports.CorrelationReport) {
	if report == nil || len(report.Events) < a.minGroup {
		return
	}

	var parts []string
	for _, e := range report.Events {
		dir := "opened"
		if !e.Opened {
			dir = "closed"
		}
		parts = append(parts, fmt.Sprintf("%s/%d(%s)", e.Proto, e.Port, dir))
	}

	fmt.Fprintf(
		a.out,
		"[CORRELATION] %d related events between %s and %s: %s\n",
		len(report.Events),
		report.StartedAt.Format("15:04:05"),
		report.EndedAt.Format("15:04:05"),
		strings.Join(parts, ", "),
	)
}
