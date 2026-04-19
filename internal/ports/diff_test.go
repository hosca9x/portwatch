package ports

import (
	"testing"
	"time"
)

func makeSnap(ports []int) *Snapshot {
	s := &Snapshot{Timestamp: time.Now()}
	for _, p := range ports {
		s.Ports = append(s.Ports, PortState{Protocol: "tcp", Port: p, Open: true})
	}
	return s
}

func TestDiff_NewPorts(t *testing.T) {
	prev := makeSnap([]int{80, 443})
	curr := makeSnap([]int{80, 443, 8080})

	opened, closed := Diff(prev, curr)

	if len(opened) != 1 || opened[0].Port != 8080 {
		t.Errorf("expected port 8080 to be opened, got %+v", opened)
	}
	if len(closed) != 0 {
		t.Errorf("expected no closed ports, got %+v", closed)
	}
}

func TestDiff_ClosedPorts(t *testing.T) {
	prev := makeSnap([]int{80, 443, 8080})
	curr := makeSnap([]int{80, 443})

	opened, closed := Diff(prev, curr)

	if len(closed) != 1 || closed[0].Port != 8080 {
		t.Errorf("expected port 8080 to be closed, got %+v", closed)
	}
	if len(opened) != 0 {
		t.Errorf("expected no opened ports, got %+v", opened)
	}
}

func TestDiff_NoPrevious(t *testing.T) {
	curr := makeSnap([]int{22, 80})
	opened, closed := Diff(nil, curr)

	if len(opened) != 2 {
		t.Errorf("expected 2 opened ports, got %d", len(opened))
	}
	if len(closed) != 0 {
		t.Errorf("expected no closed ports, got %+v", closed)
	}
}
