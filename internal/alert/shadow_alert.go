package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// ShadowAlerter emits alerts for shadow ports — ports that appeared only
// transiently and are likely ephemeral or suspicious.
type ShadowAlerter struct {
	out io.Writer
}

// NewShadowAlerter creates a ShadowAlerter writing to w.
// If w is nil, os.Stdout is used.
func NewShadowAlerter(w io.Writer) *ShadowAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &ShadowAlerter{out: w}
}

// Notify writes one alert line per shadow entry.
func (a *ShadowAlerter) Notify(shadows []ports.ShadowEntry) {
	if len(shadows) == 0 {
		return
	}
	for _, s := range shadows {
		fmt.Fprintf(
			a.out,
			"[SHADOW] port=%d proto=%s first_seen=%s last_seen=%s appearances=%d\n",
			s.Port,
			s.Proto,
			s.FirstSeen.Format(time.RFC3339),
			s.LastSeen.Format(time.RFC3339),
			s.SeenCount,
		)
	}
}
