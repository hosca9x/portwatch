package alert

import (
	"fmt"
	"io"
	"os"

	"github.com/user/portwatch/internal/ports"
)

// PriorityAlerter emits alerts filtered by minimum severity.
type PriorityAlerter struct {
	w          io.Writer
	prioritizer *ports.Prioritizer
	minSeverity ports.Severity
}

// NewPriorityAlerter creates a PriorityAlerter. If w is nil, os.Stdout is used.
func NewPriorityAlerter(w io.Writer, p *ports.Prioritizer, minSev ports.Severity) *PriorityAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &PriorityAlerter{w: w, prioritizer: p, minSeverity: minSev}
}

// Notify emits prioritized alerts for opened and closed port entries.
func (a *PriorityAlerter) Notify(opened, closed []ports.PortEntry) {
	if len(opened) > 0 {
		a.emit("OPENED", opened)
	}
	if len(closed) > 0 {
		a.emit("CLOSED", closed)
	}
}

func (a *PriorityAlerter) emit(action string, entries []ports.PortEntry) {
	events := a.prioritizer.Prioritize(entries)
	for _, ev := range events {
		if ev.Severity < a.minSeverity {
			continue
		}
		sevLabel := severityLabel(ev.Severity)
		fmt.Fprintf(a.w, "[%s] %s %s:%d/%s — %s\n",
			sevLabel, action,
			ev.Entry.Addr, ev.Entry.Port, ev.Entry.Proto,
			ev.Reason,
		)
	}
}

func severityLabel(s ports.Severity) string {
	switch s {
	case ports.SeverityHigh:
		return "HIGH"
	case ports.SeverityMedium:
		return "MEDIUM"
	default:
		return "LOW"
	}
}
