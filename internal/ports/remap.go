package ports

import "sync"

// RemapRule defines a port aliasing rule: traffic seen on From is reported as To.
type RemapRule struct {
	FromPort  int
	ToPort    int
	Proto     string // "tcp", "udp", or "" for any
}

// Remapper rewrites PortEntry fields according to a set of RemapRules.
// This is useful when a service binds on an ephemeral port but should be
// tracked under its canonical port number.
type Remapper struct {
	mu    sync.RWMutex
	rules []RemapRule
}

// NewRemapper creates a Remapper pre-loaded with the supplied rules.
func NewRemapper(rules []RemapRule) *Remapper {
	r := &Remapper{}
	r.rules = append(r.rules, rules...)
	return r
}

// AddRule appends a rule at runtime.
func (r *Remapper) AddRule(rule RemapRule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, rule)
}

// Apply returns a copy of entry with the port rewritten if a matching rule
// exists. If no rule matches, the original entry is returned unchanged.
func (r *Remapper) Apply(entry PortEntry) PortEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, rule := range r.rules {
		if rule.FromPort != entry.Port {
			continue
		}
		if rule.Proto != "" && rule.Proto != entry.Proto {
			continue
		}
		entry.Port = rule.ToPort
		return entry
	}
	return entry
}

// ApplyAll rewrites every entry in the slice, returning a new slice.
func (r *Remapper) ApplyAll(entries []PortEntry) []PortEntry {
	out := make([]PortEntry, len(entries))
	for i, e := range entries {
		out[i] = r.Apply(e)
	}
	return out
}
