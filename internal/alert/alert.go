package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelAlert Level = "ALERT"
)

// Alert holds a single alert event.
type Alert struct {
	Timestamp time.Time
	Level     Level
	Message   string
}

// Notifier writes alerts to an output destination.
type Notifier struct {
	out io.Writer
}

// NewNotifier creates a Notifier that writes to w.
// If w is nil, os.Stdout is used.
func NewNotifier(w io.Writer) *Notifier {
	if w == nil {
		w = os.Stdout
	}
	return &Notifier{out: w}
}

// Notify formats and emits alerts derived from a Diff result.
func (n *Notifier) Notify(d ports.DiffResult) {
	for _, p := range d.Opened {
		n.emit(Alert{
			Timestamp: time.Now(),
			Level:     LevelAlert,
			Message:   fmt.Sprintf("new port opened: %d/%s (pid %d)", p.Port, p.Proto, p.PID),
		})
	}
	for _, p := range d.Closed {
		n.emit(Alert{
			Timestamp: time.Now(),
			Level:     LevelWarn,
			Message:   fmt.Sprintf("port closed: %d/%s (pid %d)", p.Port, p.Proto, p.PID),
		})
	}
}

func (n *Notifier) emit(a Alert) {
	fmt.Fprintf(n.out, "[%s] %s %s\n", a.Timestamp.Format(time.RFC3339), a.Level, a.Message)
}
