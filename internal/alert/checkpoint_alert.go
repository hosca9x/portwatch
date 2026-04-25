package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// CheckpointAlerter emits a summary line whenever a checkpoint is saved,
// reporting the number of entries persisted and the timestamp.
type CheckpointAlerter struct {
	w io.Writer
}

// NewCheckpointAlerter creates a CheckpointAlerter that writes to w.
// If w is nil, os.Stdout is used.
func NewCheckpointAlerter(w io.Writer) *CheckpointAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &CheckpointAlerter{w: w}
}

// Notify writes a checkpoint summary line to the configured writer.
// It is a no-op when entries is empty.
func (a *CheckpointAlerter) Notify(entries []ports.PortEntry, savedAt time.Time) {
	if len(entries) == 0 {
		return
	}
	fmt.Fprintf(
		a.w,
		"[checkpoint] %s — %d port(s) persisted\n",
		savedAt.UTC().Format(time.RFC3339),
		len(entries),
	)
	for _, e := range entries {
		fmt.Fprintf(a.w, "  %s/%d\n", e.Proto, e.Port)
	}
}
