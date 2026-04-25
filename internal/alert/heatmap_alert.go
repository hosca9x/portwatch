package alert

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/user/portwatch/internal/ports"
)

// HeatmapAlerter emits alert lines for port keys whose heat exceeds a threshold.
type HeatmapAlerter struct {
	out       io.Writer
	minHeat   float64 // observations per minute to trigger an alert
}

// NewHeatmapAlerter creates a HeatmapAlerter writing to w.
// If w is nil, os.Stdout is used.
// minHeat is the minimum heat (obs/min) required to emit an alert line.
func NewHeatmapAlerter(w io.Writer, minHeat float64) *HeatmapAlerter {
	if w == nil {
		w = os.Stdout
	}
	if minHeat <= 0 {
		minHeat = 1.0
	}
	return &HeatmapAlerter{out: w, minHeat: minHeat}
}

// Notify writes an alert line for each report whose Heat meets the threshold.
// Reports are emitted in descending heat order.
func (a *HeatmapAlerter) Notify(reports []ports.HeatmapReport) {
	if len(reports) == 0 {
		return
	}

	var hot []ports.HeatmapReport
	for _, r := range reports {
		if r.Heat >= a.minHeat {
			hot = append(hot, r)
		}
	}
	if len(hot) == 0 {
		return
	}

	sort.Slice(hot, func(i, j int) bool {
		return hot[i].Heat > hot[j].Heat
	})

	for _, r := range hot {
		fmt.Fprintf(a.out, "[heatmap] key=%s count=%d heat=%.2f obs/min\n",
			r.Key, r.Count, r.Heat)
	}
}
