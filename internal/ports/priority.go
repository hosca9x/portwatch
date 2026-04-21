package ports

import "sort"

// Severity represents the urgency level of a port event.
type Severity int

const (
	SeverityLow    Severity = 1
	SeverityMedium Severity = 2
	SeverityHigh   Severity = 3
)

// PriorityEvent wraps a PortEntry with an assigned severity.
type PriorityEvent struct {
	Entry    PortEntry
	Severity Severity
	Reason   string
}

// Prioritizer assigns severity levels to port events based on configuration.
type Prioritizer struct {
	highPorts   map[int]struct{}
	mediumPorts map[int]struct{}
}

// NewPrioritizer creates a Prioritizer with known high- and medium-priority ports.
func NewPrioritizer(highPorts, mediumPorts []int) *Prioritizer {
	toSet := func(ports []int) map[int]struct{} {
		m := make(map[int]struct{}, len(ports))
		for _, p := range ports {
			m[p] = struct{}{}
		}
		return m
	}
	return &Prioritizer{
		highPorts:   toSet(highPorts),
		mediumPorts: toSet(mediumPorts),
	}
}

// Prioritize assigns a severity to each entry and returns sorted events (high first).
func (p *Prioritizer) Prioritize(entries []PortEntry) []PriorityEvent {
	events := make([]PriorityEvent, 0, len(entries))
	for _, e := range entries {
		sev, reason := p.classify(e)
		events = append(events, PriorityEvent{Entry: e, Severity: sev, Reason: reason})
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].Severity > events[j].Severity
	})
	return events
}

func (p *Prioritizer) classify(e PortEntry) (Severity, string) {
	if _, ok := p.highPorts[e.Port]; ok {
		return SeverityHigh, "known high-priority port"
	}
	if _, ok := p.mediumPorts[e.Port]; ok {
		return SeverityMedium, "known medium-priority port"
	}
	if e.Port < 1024 {
		return SeverityMedium, "privileged port (<1024)"
	}
	return SeverityLow, "unprivileged port"
}
