package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// PipelineAlerter consumes PortEntry values from a channel produced by a
// ports.Pipeline and writes a human-readable alert line for each entry.
type PipelineAlerter struct {
	out io.Writer
}

// NewPipelineAlerter returns a PipelineAlerter that writes to w.
// If w is nil, os.Stdout is used.
func NewPipelineAlerter(w io.Writer) *PipelineAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &PipelineAlerter{out: w}
}

// Consume reads all entries from ch until it is closed and emits an alert
// line for each one. It blocks until ch is closed.
func (a *PipelineAlerter) Consume(ch <-chan ports.PortEntry) {
	for e := range ch {
		a.emit(e)
	}
}

func (a *PipelineAlerter) emit(e ports.PortEntry) {
	fmt.Fprintf(
		a.out,
		"[pipeline-alert] %s port=%d proto=%s\n",
		time.Now().UTC().Format(time.RFC3339),
		e.Port,
		e.Proto,
	)
}
