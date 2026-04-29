package alert

import (
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/user/portwatch/internal/ports"
)

// CensusAlerter emits a summary of port census data to a writer.
type CensusAlerter struct {
	out       io.Writer
	minCount  int
}

// NewCensusAlerter creates a CensusAlerter that only reports entries
// observed at least minCount times. Defaults to stdout when w is nil.
func NewCensusAlerter(w io.Writer, minCount int) *CensusAlerter {
	if w == nil {
		w = os.Stdout
	}
	if minCount < 1 {
		minCount = 1
	}
	return &CensusAlerter{out: w, minCount: minCount}
}

// Notify writes a census report for all entries meeting the minimum count
// threshold. Entries are sorted by key for deterministic output.
func (a *CensusAlerter) Notify(snap ports.CensusSnapshot) {
	if len(snap.Entries) == 0 {
		return
	}
	filtered := make([]ports.CensusEntry, 0, len(snap.Entries))
	for _, e := range snap.Entries {
		if e.Count >= a.minCount {
			filtered = append(filtered, e)
		}
	}
	if len(filtered) == 0 {
		return
	}
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Key < filtered[j].Key
	})
	fmt.Fprintf(a.out, "[census] snapshot at %s — %d port(s) active\n",
		snap.TakenAt.Format("2006-01-02T15:04:05Z07:00"), len(filtered))
	for _, e := range filtered {
		fmt.Fprintf(a.out, "  %s  seen=%d  first=%s  last=%s\n",
			e.Key,
			e.Count,
			e.FirstSeen.Format("15:04:05"),
			e.LastSeen.Format("15:04:05"),
		)
	}
}
