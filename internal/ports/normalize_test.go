package ports

import (
	"testing"
)

func baseNormEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto}
}

func TestNormalizer_ProtoLowercased(t *testing.T) {
	n := NewNormalizer(nil)
	e := n.Apply(baseNormEntry(80, "TCP"))
	if e.Proto != "tcp" {
		t.Fatalf("expected proto 'tcp', got %q", e.Proto)
	}
}

func TestNormalizer_ProtoTrimmed(t *testing.T) {
	n := NewNormalizer(nil)
	e := n.Apply(baseNormEntry(443, "  UDP  "))
	if e.Proto != "udp" {
		t.Fatalf("expected proto 'udp', got %q", e.Proto)
	}
}

func TestNormalizer_PortRemappedByRule(t *testing.T) {
	rules := []NormalizeRule{
		{PortFrom: 8080, PortTo: 80, Proto: "tcp"},
	}
	n := NewNormalizer(rules)
	e := n.Apply(baseNormEntry(8080, "tcp"))
	if e.Port != 80 {
		t.Fatalf("expected port 80, got %d", e.Port)
	}
}

func TestNormalizer_PortNotRemappedWrongProto(t *testing.T) {
	rules := []NormalizeRule{
		{PortFrom: 8080, PortTo: 80, Proto: "tcp"},
	}
	n := NewNormalizer(rules)
	e := n.Apply(baseNormEntry(8080, "udp"))
	if e.Port != 8080 {
		t.Fatalf("expected port 8080 unchanged, got %d", e.Port)
	}
}

func TestNormalizer_TagAddedByRule(t *testing.T) {
	rules := []NormalizeRule{
		{Tag: "web", Proto: "tcp"},
	}
	n := NewNormalizer(rules)
	e := n.Apply(baseNormEntry(80, "tcp"))
	if !containsTag(e.Tags, "web") {
		t.Fatalf("expected tag 'web' to be added, got %v", e.Tags)
	}
}

func TestNormalizer_TagNotDuplicated(t *testing.T) {
	rules := []NormalizeRule{
		{Tag: "web", Proto: "tcp"},
	}
	n := NewNormalizer(rules)
	entry := baseNormEntry(80, "tcp")
	entry.Tags = []string{"web"}
	e := n.Apply(entry)
	count := 0
	for _, tag := range e.Tags {
		if tag == "web" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected tag 'web' exactly once, got %d", count)
	}
}

func TestNormalizer_ApplyAll(t *testing.T) {
	n := NewNormalizer(nil)
	entries := []PortEntry{
		baseNormEntry(22, "TCP"),
		baseNormEntry(53, "UDP"),
	}
	out := n.ApplyAll(entries)
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	if out[0].Proto != "tcp" || out[1].Proto != "udp" {
		t.Fatalf("unexpected protos: %q %q", out[0].Proto, out[1].Proto)
	}
}
