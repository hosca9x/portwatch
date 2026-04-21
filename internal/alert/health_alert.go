package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// HealthAlerter emits alerts when the scanner reports an unhealthy status.
type HealthAlerter struct {
	out         io.Writer
	lastHealthy bool
	lastAlerted time.Time
	cooldown    time.Duration
	clock       func() time.Time
}

// NewHealthAlerter returns a HealthAlerter writing to w.
// If w is nil, os.Stdout is used.
// cooldown limits how often repeated unhealthy alerts are emitted.
func NewHealthAlerter(w io.Writer, cooldown time.Duration) *HealthAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &HealthAlerter{
		out:         w,
		lastHealthy: true,
		cooldown:    cooldown,
		clock:       time.Now,
	}
}

// Evaluate checks the provided HealthStatus and emits an alert if the scanner
// has become unhealthy or remains unhealthy past the cooldown window.
func (ha *HealthAlerter) Evaluate(st ports.HealthStatus) {
	now := ha.clock()
	if st.Healthy {
		if !ha.lastHealthy {
			fmt.Fprintf(ha.out, "[HEALTH] scanner recovered after %d consecutive errors\n",
				st.ConsecErrors)
		}
		ha.lastHealthy = true
		return
	}

	// Unhealthy: emit if first transition or cooldown has elapsed.
	if ha.lastHealthy || now.Sub(ha.lastAlerted) >= ha.cooldown {
		reason := "unknown"
		if st.LastError != nil {
			reason = st.LastError.Error()
		} else if !st.LastScanTime.IsZero() {
			reason = fmt.Sprintf("stale scan (last: %s)", st.LastScanTime.Format(time.RFC3339))
		}
		fmt.Fprintf(ha.out, "[HEALTH] scanner unhealthy: %s (consec_errors=%d uptime=%s)\n",
			reason, st.ConsecErrors, st.Uptime.Round(time.Second))
		ha.lastAlerted = now
	}
	ha.lastHealthy = false
}
