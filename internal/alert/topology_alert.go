package alert

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/user/portwatch/internal/ports"
)

// TopologyAlerter emits human-readable lines describing the current
// port co-occurrence graph produced by a TopologyTracker.
type TopologyAlerter struct {
	out io.Writer
}

// NewTopologyAlerter returns a TopologyAlerter that writes to w.
// If w is nil it defaults to os.Stdout.
func NewTopologyAlerter(w io.Writer) *TopologyAlerter {
	if w == nil {
		w = os.Stdout
	}
	return &TopologyAlerter{out: w}
}

// Notify writes one line per edge pair in the snapshot.
// If the snapshot has no edges nothing is written.
func (a *TopologyAlerter) Notify(snap ports.TopologySnapshot) {
	if len(snap.Edges) == 0 {
		return
	}

	// Collect unique canonical pairs to avoid printing A-B and B-A.
	emitted := make(map[string]struct{})
	var lines []string
	for src, neighbours := range snap.Edges {
		for _, dst := range neighbours {
			pair := canonicalPair(src, dst)
			if _, ok := emitted[pair]; ok {
				continue
			}
			emitted[pair] = struct{}{}
			lines = append(lines, fmt.Sprintf(
				"[topology] co-observed: %s <-> %s (at %s)",
				src, dst, snap.Timestamp.Format("15:04:05"),
			))
		}
	}
	sort.Strings(lines)
	for _, l := range lines {
		fmt.Fprintln(a.out, l)
	}
}

func canonicalPair(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return strings.Join([]string{a, b}, "|")
}
