package ports

import (
	"testing"

	"github.com/user/portwatch/internal/config"
)

func baseCfg() *config.Config {
	c := config.Default()
	return c
}

func TestFilter_AllowExcludedPort(t *testing.T) {
	cfg := baseCfg()
	cfg.ExcludePorts = []int{22, 80}
	f := NewFilter(cfg)
	if f.Allow("tcp", 22) {
		t.Error("port 22 should be excluded")
	}
	if !f.Allow("tcp", 443) {
		t.Error("port 443 should be allowed")
	}
}

func TestFilter_AllowProto(t *testing.T) {
	cfg := baseCfg()
	cfg.Protocols = []string{"tcp"}
	f := NewFilter(cfg)
	if f.Allow("udp", 53) {
		t.Error("udp should not be allowed when only tcp configured")
	}
	if !f.Allow("tcp", 8080) {
		t.Error("tcp should be allowed")
	}
}

func TestFilter_Apply(t *testing.T) {
	cfg := baseCfg()
	cfg.ExcludePorts = []int{22}
	cfg.Protocols = []string{"tcp", "udp"}
	f := NewFilter(cfg)
	entries := []PortEntry{
		{Proto: "tcp", Port: 22},
		{Proto: "tcp", Port: 443},
		{Proto: "udp", Port: 53},
	}
	result := f.Apply(entries)
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}

func TestFilter_EmptyExclusions(t *testing.T) {
	cfg := baseCfg()
	f := NewFilter(cfg)
	if !f.Allow("tcp", 9999) {
		t.Error("all ports should be allowed with empty exclusions and matching proto")
	}
}
