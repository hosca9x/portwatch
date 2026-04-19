package ports

// Diff compares two snapshots and returns newly opened and newly closed ports.
func Diff(prev, curr *Snapshot) (opened, closed []PortState) {
	prevMap := indexPorts(prev)
	currMap := indexPorts(curr)

	for key, state := range currMap {
		if _, existed := prevMap[key]; !existed {
			opened = append(opened, state)
		}
	}

	for key, state := range prevMap {
		if _, exists := currMap[key]; !exists {
			closed = append(closed, state)
		}
	}

	return opened, closed
}

func indexPorts(snap *Snapshot) map[string]PortState {
	m := make(map[string]PortState)
	if snap == nil {
		return m
	}
	for _, p := range snap.Ports {
		key := p.Protocol + ":" + itoa(p.Port)
		m[key] = p
	}
	return m
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}
