package alert

import (
	"fmt"
	"io"
	"os"

	"github.com/user/portwatch/internal/ports"
)

// ClusterAlerter emits alerts when contiguous port clusters are detected.
type ClusterAlerter struct {
	w io.Writer
}

// NewClusterAlerter creates a ClusterAlerter writing to w.
// If w is nil, output goes to stdout.
func NewClusterAlerter(w io.Writer) *ClusterAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &ClusterAlerter{w: w}
}

// Notify emits one alert line per cluster report.
func (a *ClusterAlerter) Notify(reports []ports.ClusterReport) {
	for _, r := range reports {
		fmt.Fprintf(a.w, "[cluster] proto=%s ports=%d-%d size=%d seen=%s\n",
			r.Proto, r.Start, r.End, r.Size, r.SeenAt.Format("15:04:05"))
	}
}

// NotifyIfChanged emits alerts only when the cluster set has changed since last call.
func (a *ClusterAlerter) NotifyIfChanged(prev, curr []ports.ClusterReport) {
	if clustersEqual(prev, curr) {
		return
	}
	a.Notify(curr)
}

func clustersEqual(a, b []ports.ClusterReport) bool {
	if len(a) != len(b) {
		return false
	}
	index := make(map[string]ports.ClusterReport, len(a))
	for _, r := range a {
		index[clusterKey(r)] = r
	}
	for _, r := range b {
		if _, ok := index[clusterKey(r)]; !ok {
			return false
		}
	}
	return true
}

func clusterKey(r ports.ClusterReport) string {
	return fmt.Sprintf("%s:%d-%d", r.Proto, r.Start, r.End)
}
