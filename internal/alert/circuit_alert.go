package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// CircuitEvent describes a circuit breaker state transition.
type CircuitEvent struct {
	Target    string
	State     ports.CircuitState
	Failures  int
	Timestamp time.Time
}

// CircuitAlerter emits alerts when a circuit breaker changes state.
type CircuitAlerter struct {
	out io.Writer
}

// NewCircuitAlerter creates a CircuitAlerter writing to w.
// If w is nil, output defaults to os.Stdout.
func NewCircuitAlerter(w io.Writer) *CircuitAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &CircuitAlerter{out: w}
}

// Notify emits an alert for the given circuit event.
// No output is produced for closed-state transitions (recovery is silent
// unless you want verbose mode — extend as needed).
func (ca *CircuitAlerter) Notify(ev CircuitEvent) {
	switch ev.State {
	case ports.CircuitOpen:
		fmt.Fprintf(ca.out, "[CIRCUIT OPEN]  target=%s failures=%d at=%s\n",
			ev.Target, ev.Failures, ev.Timestamp.Format(time.RFC3339))
	case ports.CircuitHalfOpen:
		fmt.Fprintf(ca.out, "[CIRCUIT PROBE] target=%s probing recovery at=%s\n",
			ev.Target, ev.Timestamp.Format(time.RFC3339))
	case ports.CircuitClosed:
		fmt.Fprintf(ca.out, "[CIRCUIT OK]    target=%s recovered at=%s\n",
			ev.Target, ev.Timestamp.Format(time.RFC3339))
	}
}
