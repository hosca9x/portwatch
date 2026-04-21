package alert

import (
	"fmt"
	"io"
	"os"

	"github.com/user/portwatch/internal/ports"
)

// AnomalyAlerter emits human-readable alerts for detected port anomalies.
type AnomalyAlerter struct {
	out io.Writer
}

// NewAnomalyAlerter creates an AnomalyAlerter writing to w.
// If w is nil, os.Stdout is used.
func NewAnomalyAlerter(w io.Writer) *AnomalyAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &AnomalyAlerter{out: w}
}

// Notify emits an alert for the given AnomalyReport.
// Returns without output if report is nil.
func (a *AnomalyAlerter) Notify(report *ports.AnomalyReport) error {
	if report == nil {
		return nil
	}
	switch report.Kind {
	case ports.AnomalyFlapping:
		_, err := fmt.Fprintf(
			a.out,
			"[ANOMALY] FLAPPING detected on %s: %d transitions in %s\n",
			report.Key, report.Count, report.Window,
		)
		return err
	case ports.AnomalyBurst:
		_, err := fmt.Fprintf(
			a.out,
			"[ANOMALY] BURST detected: %d new ports opened within %s\n",
			report.Count, report.Window,
		)
		return err
	default:
		_, err := fmt.Fprintf(
			a.out,
			"[ANOMALY] %s on %s: count=%d window=%s\n",
			report.Kind, report.Key, report.Count, report.Window,
		)
		return err
	}
}
