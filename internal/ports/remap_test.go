package ports

import (
	"testing"
)

func makeRemapEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto}
}

func TestRemapper_NoRulesReturnsOriginal(t *testing.T) {
	r := NewRemapper(nil)
	e := makeRemapEntry(8080, "tcp")
	got := r.Apply(e)
	if got.Port != 8080 {
		t.Fatalf("expected 8080, got %d", got.Port)
	}
}

func TestRemapper_MatchingRuleRewritesPort(t *testing.T) {
	rules := []RemapRule{{FromPort: 8080, ToPort: 80, Proto: "tcp"}}
	r := NewRemapper(rules)
	got := r.Apply(makeRemapEntry(8080, "tcp"))
	if got.Port != 80 {
		t.Fatalf("expected 80, got %d", got.Port)
	}
}

func TestRemapper_WrongProtoNoRewrite(t *testing.T) {
	rules := []RemapRule{{FromPort: 8080, ToPort: 80, Proto: "tcp"}}
	r := NewRemapper(rules)
	got := r.Apply(makeRemapEntry(8080, "udp"))
	if got.Port != 8080 {
		t.Fatalf("expected 8080 (no rewrite), got %d", got.Port)
	}
}

func TestRemapper_EmptyProtoMatchesAny(t *testing.T) {
	rules := []RemapRule{{FromPort: 9000, ToPort: 443, Proto: ""}}
	r := NewRemapper(rules)
	for _, proto := range []string{"tcp", "udp"} {
		got := r.Apply(makeRemapEntry(9000, proto))
		if got.Port != 443 {
			t.Fatalf("proto=%s: expected 443, got %d", proto, got.Port)
		}
	}
}

func TestRemapper_AddRuleDynamic(t *testing.T) {
	r := NewRemapper(nil)
	r.AddRule(RemapRule{FromPort: 5432, ToPort: 5432, Proto: "tcp"})
	got := r.Apply(makeRemapEntry(5432, "tcp"))
	if got.Port != 5432 {
		t.Fatalf("expected 5432, got %d", got.Port)
	}
}

func TestRemapper_ApplyAll(t *testing.T) {
	rules := []RemapRule{{FromPort: 8080, ToPort: 80, Proto: "tcp"}}
	r := NewRemapper(rules)
	entries := []PortEntry{
		makeRemapEntry(8080, "tcp"),
		makeRemapEntry(443, "tcp"),
	}
	out := r.ApplyAll(entries)
	if len(out) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(out))
	}
	if out[0].Port != 80 {
		t.Errorf("entry 0: expected 80, got %d", out[0].Port)
	}
	if out[1].Port != 443 {
		t.Errorf("entry 1: expected 443, got %d", out[1].Port)
	}
}
