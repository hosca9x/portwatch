// Package ports provides functionality for scanning, filtering, and
// comparing open network ports on the local system.
package ports

import (
	"fmt"
	"net"
	"time"

	"github.com/user/portwatch/internal/config"
)

// Scanner scans the local system for open ports.
type Scanner struct {
	cfg    *config.Config
	filter *Filter
}

// NewScanner creates a new Scanner using the provided configuration.
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		cfg:    cfg,
		filter: NewFilter(cfg),
	}
}

// Scan probes each protocol/port combination defined by the config and
// returns a snapshot of all ports that are currently open and pass the
// filter rules.
func (s *Scanner) Scan() ([]Entry, error) {
	var entries []Entry

	for _, proto := range s.cfg.Protocols {
		for port := 1; port <= 65535; port++ {
			if !s.filter.Allow(proto, port) {
				continue
			}

			open, err := probePort(proto, port, s.cfg.Timeout)
			if err != nil {
				// Connection refused or timeout — port is closed, skip silently.
				continue
			}
			if open {
				entries = append(entries, Entry{
					Proto: proto,
					Port:  port,
				})
			}
		}
	}

	return entries, nil
}

// probePort attempts a connection to the given protocol/port on localhost.
// Returns true if the port is open, false if it is closed or filtered.
func probePort(proto string, port int, timeout time.Duration) (bool, error) {
	address := fmt.Sprintf("127.0.0.1:%d", port)

	switch proto {
	case "tcp", "tcp4", "tcp6":
		conn, err := net.DialTimeout(proto, address, timeout)
		if err != nil {
			return false, err
		}
		_ = conn.Close()
		return true, nil

	case "udp", "udp4", "udp6":
		// UDP probing is inherently unreliable; we attempt a send/recv
		// with a short deadline and treat a response as "open".
		conn, err := net.DialTimeout(proto, address, timeout)
		if err != nil {
			return false, err
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(timeout))
		_, _ = conn.Write([]byte{})
		buf := make([]byte, 1)
		_, err = conn.Read(buf)
		if err != nil {
			// No response is ambiguous for UDP; treat as closed.
			return false, err
		}
		return true, nil

	default:
		return false, fmt.Errorf("unsupported protocol: %s", proto)
	}
}
