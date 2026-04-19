package ports

import (
	"fmt"
	"net"
	"time"
)

// PortState represents the state of a single port.
type PortState struct {
	Protocol string
	Port     int
	Open     bool
}

// Snapshot holds all observed open ports at a point in time.
type Snapshot struct {
	Timestamp time.Time
	Ports     []PortState
}

// Scanner scans local ports within a given range.
type Scanner struct {
	StartPort int
	EndPort   int
	Timeout   time.Duration
}

// NewScanner creates a Scanner with sensible defaults.
func NewScanner(start, end int) *Scanner {
	return &Scanner{
		StartPort: start,
		EndPort:   end,
		Timeout:   500 * time.Millisecond,
	}
}

// Scan probes each TCP port in the configured range and returns a Snapshot.
func (s *Scanner) Scan() (*Snapshot, error) {
	if s.StartPort < 1 || s.EndPort > 65535 || s.StartPort > s.EndPort {
		return nil, fmt.Errorf("invalid port range %d-%d", s.StartPort, s.EndPort)
	}

	snap := &Snapshot{Timestamp: time.Now()}

	for port := s.StartPort; port <= s.EndPort; port++ {
		addr := fmt.Sprintf("127.0.0.1:%d", port)
		conn, err := net.DialTimeout("tcp", addr, s.Timeout)
		if err == nil {
			conn.Close()
			snap.Ports = append(snap.Ports, PortState{Protocol: "tcp", Port: port, Open: true})
		}
	}

	return snap, nil
}
