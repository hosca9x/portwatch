package ports

import (
	"strings"
)

// NormalizeRule describes a single normalization transformation applied to a PortEntry.
// Rules are matched by protocol (case-insensitive) and optional port range.
type NormalizeRule struct {
	// Proto restricts the rule to a specific protocol (e.g. "tcp", "udp").
	// An empty string matches any protocol.
	Proto string

	// MinPort and MaxPort define an inclusive port range.
	// Zero values are treated as "no bound" (0 == any low, 0 == any high when both zero).
	MinPort uint16
	MaxPort uint16

	// CanonicalProto, if set, rewrites the protocol field to this value.
	CanonicalProto string

	// TagAdd is a list of tags to attach to matching entries.
	TagAdd []string
}

// Normalizer applies a set of NormalizeRules to PortEntry values, producing
// a consistently labelled stream suitable for downstream alerting and storage.
type Normalizer struct {
	rules []NormalizeRule
}

// NewNormalizer creates a Normalizer with the provided rules.
// Rules are evaluated in order; all matching rules are applied (not first-match).
func NewNormalizer(rules []NormalizeRule) *Normalizer {
	return &Normalizer{rules: rules}
}

// Apply runs all rules against entry and returns the (possibly mutated) copy.
func (n *Normalizer) Apply(e PortEntry) PortEntry {
	for _, r := range n.rules {
		if !r.matches(e) {
			continue
		}
		if r.CanonicalProto != "" {
			e.Proto = r.CanonicalProto
		}
		for _, tag := range r.TagAdd {
			if !containsTag(e.Tags, tag) {
				e.Tags = append(e.Tags, tag)
			}
		}
	}
	return e
}

// ApplyAll normalises a slice of entries in place, returning the updated slice.
func (n *Normalizer) ApplyAll(entries []PortEntry) []PortEntry {
	out := make([]PortEntry, len(entries))
	for i, e := range entries {
		out[i] = n.Apply(e)
	}
	return out
}

// matches reports whether the rule applies to the given entry.
func (r NormalizeRule) matches(e PortEntry) bool {
	if r.Proto != "" && !strings.EqualFold(r.Proto, e.Proto) {
		return false
	}
	// If both bounds are zero the rule applies to every port.
	if r.MinPort == 0 && r.MaxPort == 0 {
		return true
	}
	if r.MinPort > 0 && e.Port < r.MinPort {
		return false
	}
	if r.MaxPort > 0 && e.Port > r.MaxPort {
		return false
	}
	return true
}

// containsTag returns true if tag already exists in the slice.
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
