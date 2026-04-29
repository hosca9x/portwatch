package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// EnvelopeAlerter emits a warning when a scan envelope is determined to be
// stale, indicating that the most recent snapshot may not reflect the current
// state of the host.
type EnvelopeAlerter struct {
	policy ports.EnvelopePolicy
	w      io.Writer
	clock  func() time.Time
}

// NewEnvelopeAlerter returns an EnvelopeAlerter writing to w. If w is nil,
// os.Stdout is used.
func NewEnvelopeAlerter(policy ports.EnvelopePolicy, w io.Writer) *EnvelopeAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &EnvelopeAlerter{policy: policy, w: w, clock: time.Now}
}

// Check inspects env and writes an alert line if the envelope is stale.
// It returns true when an alert was emitted.
func (a *EnvelopeAlerter) Check(env *ports.Envelope) bool {
	if env == nil {
		return false
	}
	age := a.clock().Sub(env.ScannedAt)
	if age <= a.policy.MaxAge {
		return false
	}
	fmt.Fprintf(
		a.w,
		"[envelope] STALE snapshot from %q scanned at %s (age %s > max %s)\n",
		env.Source,
		env.ScannedAt.Format(time.RFC3339),
		age.Round(time.Second),
		a.policy.MaxAge,
	)
	return true
}

// CheckAll calls Check for each envelope and returns the number of stale ones.
func (a *EnvelopeAlerter) CheckAll(envs []*ports.Envelope) int {
	count := 0
	for _, e := range envs {
		if a.Check(e) {
			count++
		}
	}
	return count
}
