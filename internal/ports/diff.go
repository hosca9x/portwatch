package ports

import "strconv"

// DiffResult contains ports that appeared or disappeared between two snapshots.
type DiffResult struct {
	Opened []Entry
	Closed []Entry
}

// HasChanges reports whether there are any opened or closed ports.
func (d *DiffResult) HasChanges() bool {
	return len(d.Opened) > 0 || len(d.Closed) > 0
}

// Diff computes the difference between a previous and current snapshot.
// When prev is nil (first run) all current ports are treated as opened.
func Diff(prev, current []Entry) *DiffResult {
	if prev == nil {
		return &DiffResult{Opened: current}
	}

	prevIdx := indexPorts(prev)
	currIdx := indexPorts(current)

	dr := &DiffResult{}
	for key, e := range currIdx {
		if _, ok := prevIdx[key]; !ok {
			dr.Opened = append(dr.Opened, e)
		}
	}
	for key, e := range prevIdx {
		if _, ok := currIdx[key]; !ok {
			dr.Closed = append(dr.Closed, e)
		}
	}
	return dr
}

func indexPorts(entries []Entry) map[string]Entry {
	m := make(map[string]Entry, len(entries))
	for _, e := range entries {
		m[e.Key()] = e
	}
	return m
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
