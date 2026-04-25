package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// BudgetReport describes a budget exhaustion event for a key.
type BudgetReport struct {
	Key       string
	Used      int
	Max       int
	Window    time.Duration
	Timestamp time.Time
}

// BudgetAlerter emits alerts when a scan budget is exhausted.
type BudgetAlerter struct {
	out io.Writer
}

// NewBudgetAlerter creates a BudgetAlerter writing to w.
// If w is nil, os.Stdout is used.
func NewBudgetAlerter(w io.Writer) *BudgetAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &BudgetAlerter{out: w}
}

// Notify emits a formatted alert for each BudgetReport in reports.
func (a *BudgetAlerter) Notify(reports []BudgetReport) {
	for _, r := range reports {
		fmt.Fprintf(
			a.out,
			"[BUDGET] %s key=%s used=%d max=%d window=%s\n",
			r.Timestamp.Format(time.RFC3339),
			r.Key,
			r.Used,
			r.Max,
			r.Window,
		)
	}
}
