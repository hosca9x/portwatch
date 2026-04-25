package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// AdaptiveAlerter emits alerts when the adaptive rate limiter's rate changes
// significantly, indicating the system is under pressure or recovering.
type AdaptiveAlerter struct {
	out          io.Writer
	limiter      *ports.AdaptiveRateLimiter
	lastRate     float64
	changeThresh float64 // fractional change required to emit alert
}

// NewAdaptiveAlerter creates an AdaptiveAlerter wrapping the given limiter.
// changeThreshold is the minimum fractional change (e.g. 0.2 = 20%) to alert.
func NewAdaptiveAlerter(limiter *ports.AdaptiveRateLimiter, changeThreshold float64, out io.Writer) *AdaptiveAlerter {
	if out == nil {
		out = os.Stdout
	}
	if changeThreshold <= 0 {
		changeThreshold = 0.2
	}
	return &AdaptiveAlerter{
		out:          out,
		limiter:      limiter,
		lastRate:     limiter.CurrentRate(),
		changeThresh: changeThreshold,
	}
}

// Check evaluates the current rate and emits an alert if it has changed
// beyond the configured threshold since the last check.
func (a *AdaptiveAlerter) Check() {
	current := a.limiter.CurrentRate()
	if a.lastRate == 0 {
		a.lastRate = current
		return
	}
	delta := (current - a.lastRate) / a.lastRate
	if delta < 0 {
		delta = -delta
	}
	if delta < a.changeThresh {
		return
	}
	direction := "increased"
	if current < a.lastRate {
		direction = "decreased"
	}
	fmt.Fprintf(a.out, "[%s] adaptive-rate-limiter: rate %s from %.2f to %.2f (%.0f%% change)\n",
		time.Now().UTC().Format(time.RFC3339),
		direction,
		a.lastRate,
		current,
		delta*100,
	)
	a.lastRate = current
}
