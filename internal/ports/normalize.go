package ports

import "strings"

// NormalizeRule describes a single normalization transformation.
type NormalizeRule struct {
	Tag      string
	Proto    string
	PortFrom int
	PortTo   int
}

// Normalizer applies a set of rules to canonicalize PortEntry fields.
type Normalizer struct {
	rules []NormalizeRule
}

// NewNormalizer constructs a Normalizer with the given rules.
func NewNormalizer(rules []NormalizeRule) *Normalizer {
	return &Normalizer{rules: rules}
}

// Apply returns a copy of e with normalization rules applied.
func (n *Normalizer) Apply(e PortEntry) PortEntry {
	out := e
	out.Proto = normalizeProto(e.Proto)
	for _, r := range n.rules {
		if r.PortFrom > 0 && e.Port == r.PortFrom {
			if r.Proto == "" || strings.EqualFold(r.Proto, out.Proto) {
				if r.PortTo > 0 {
					out.Port = r.PortTo
				}
			}
		}
		if r.Tag != "" && !containsTag(out.Tags, r.Tag) {
			if r.Proto == "" || strings.EqualFold(r.Proto, out.Proto) {
				out.Tags = append(out.Tags, r.Tag)
			}
		}
	}
	return out
}

// ApplyAll normalizes a slice of entries.
func (n *Normalizer) ApplyAll(entries []PortEntry) []PortEntry {
	out := make([]PortEntry, len(entries))
	for i, e := range entries {
		out[i] = n.Apply(e)
	}
	return out
}

// normalizeProto lowercases and trims the protocol string.
func normalizeProto(proto string) string {
	return strings.ToLower(strings.TrimSpace(proto))
}

// containsTag returns true if tag is already present in tags.
func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
