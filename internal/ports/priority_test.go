package ports

import (
	"testing"
)

func makePriorityEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto, Addr: "0.0.0.0"}
}

func TestPrioritizer_HighPort(t *testing.T) {
	p := NewPrioritizer([]int{22, 443}, []int{80, 8080})
	events := p.Prioritize([]PortEntry{makePriorityEntry(22, "tcp")})
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].Severity != SeverityHigh {
		t.Errorf("expected SeverityHigh, got %v", events[0].Severity)
	}
}

func TestPrioritizer_MediumPort(t *testing.T) {
	p := NewPrioritizer([]int{22}, []int{80})
	events := p.Prioritize([]PortEntry{makePriorityEntry(80, "tcp")})
	if events[0].Severity != SeverityMedium {
		t.Errorf("expected SeverityMedium, got %v", events[0].Severity)
	}
}

func TestPrioritizer_PrivilegedFallback(t *testing.T) {
	p := NewPrioritizer(nil, nil)
	events := p.Prioritize([]PortEntry{makePriorityEntry(512, "tcp")})
	if events[0].Severity != SeverityMedium {
		t.Errorf("expected SeverityMedium for privileged port, got %v", events[0].Severity)
	}
}

func TestPrioritizer_LowPort(t *testing.T) {
	p := NewPrioritizer(nil, nil)
	events := p.Prioritize([]PortEntry{makePriorityEntry(9000, "tcp")})
	if events[0].Severity != SeverityLow {
		t.Errorf("expected SeverityLow, got %v", events[0].Severity)
	}
}

func TestPrioritizer_SortedHighFirst(t *testing.T) {
	p := NewPrioritizer([]int{443}, []int{80})
	entries := []PortEntry{
		makePriorityEntry(9000, "tcp"),
		makePriorityEntry(80, "tcp"),
		makePriorityEntry(443, "tcp"),
	}
	events := p.Prioritize(entries)
	if events[0].Severity != SeverityHigh {
		t.Errorf("first event should be high priority, got %v", events[0].Severity)
	}
	if events[1].Severity != SeverityMedium {
		t.Errorf("second event should be medium priority, got %v", events[1].Severity)
	}
	if events[2].Severity != SeverityLow {
		t.Errorf("third event should be low priority, got %v", events[2].Severity)
	}
}
