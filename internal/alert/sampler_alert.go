package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/user/portwatch/internal/ports"
)

// SamplerReport summarises a snapshot window for alerting.
type SamplerReport struct {
	WindowStart time.Time
	WindowEnd   time.Time
	SampleCount int
	UniquePorts []int
}

// SamplerAlerter emits periodic digest alerts derived from sampler windows.
type SamplerAlerter struct {
	out io.Writer
}

// NewSamplerAlerter returns a SamplerAlerter writing to w.
// If w is nil, os.Stdout is used.
func NewSamplerAlerter(w io.Writer) *SamplerAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &SamplerAlerter{out: w}
}

// Digest builds a SamplerReport from the provided samples and emits it if
// the window contains at least one sample.
func (a *SamplerAlerter) Digest(samples []ports.Sample) *SamplerReport {
	if len(samples) == 0 {
		return nil
	}

	seen := map[int]struct{}{}
	for _, s := range samples {
		for _, e := range s.Entries {
			seen[e.Port] = struct{}{}
		}
	}

	unique := make([]int, 0, len(seen))
	for p := range seen {
		unique = append(unique, p)
	}

	report := &SamplerReport{
		WindowStart: samples[0].At,
		WindowEnd:   samples[len(samples)-1].At,
		SampleCount: len(samples),
		UniquePorts: unique,
	}

	fmt.Fprintf(a.out,
		"[sampler] window %s–%s: %d samples, %d unique ports\n",
		report.WindowStart.Format(time.RFC3339),
		report.WindowEnd.Format(time.RFC3339),
		report.SampleCount,
		len(report.UniquePorts),
	)
	return report
}
