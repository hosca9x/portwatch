package ports

import "github.com/user/portwatch/internal/config"

// Filter holds compiled exclusion rules from config.
type Filter struct {
	excludePorts map[int]struct{}
	excludeProtos map[string]struct{}
}

// NewFilter builds a Filter from the loaded configuration.
func NewFilter(cfg *config.Config) *Filter {
	ep := make(map[int]struct{}, len(cfg.ExcludePorts))
	for _, p := range cfg.ExcludePorts {
		ep[p] = struct{}{}
	}
	proto := make(map[string]struct{}, len(cfg.Protocols))
	for _, p := range cfg.Protocols {
		proto[p] = struct{}{}
	}
	return &Filter{excludePorts: ep, excludeProtos: proto}
}

// Allow returns true when the entry should be included in monitoring.
func (f *Filter) Allow(proto string, port int) bool {
	if _, excluded := f.excludePorts[port]; excluded {
		return false
	}
	if len(f.excludeProtos) > 0 {
		if _, ok := f.excludeProtos[proto]; !ok {
			return false
		}
	}
	return true
}

// Apply filters a slice of PortEntry values in-place (returns new slice).
func (f *Filter) Apply(entries []PortEntry) []PortEntry {
	out := entries[:0:0]
	for _, e := range entries {
		if f.Allow(e.Proto, e.Port) {
			out = append(out, e)
		}
	}
	return out
}
