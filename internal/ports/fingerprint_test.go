package ports

import (
	"testing"
)

func makeEntries(specs [][2]interface{}) []PortEntry {
	out := make([]PortEntry, 0, len(specs))
	for _, s := range specs {
		out = append(out, PortEntry{
			Proto: s[0].(string),
			Port:  s[1].(int),
		})
	}
	return out
}

func TestCompute_DeterministicRegardlessOfOrder(t *testing.T) {
	a := makeEntries([][2]interface{}{{"tcp", 80}, {"tcp", 443}})
	b := makeEntries([][2]interface{}{{"tcp", 443}, {"tcp", 80}})

	if Compute(a) != Compute(b) {
		t.Error("expected same fingerprint for same entries in different order")
	}
}

func TestCompute_DiffersOnChange(t *testing.T) {
	a := makeEntries([][2]interface{}{{"tcp", 80}})
	b := makeEntries([][2]interface{}{{"tcp", 8080}})

	if Compute(a) == Compute(b) {
		t.Error("expected different fingerprints for different entries")
	}
}

func TestCompute_EmptyIsStable(t *testing.T) {
	f1 := Compute(nil)
	f2 := Compute([]PortEntry{})
	if f1 != f2 {
		t.Error("expected nil and empty slice to produce same fingerprint")
	}
}

func TestFingerprintTracker_ChangedOnFirst(t *testing.T) {
	ft := NewFingerprintTracker()
	entries := makeEntries([][2]interface{}{{"tcp", 22}})

	if !ft.Changed(entries) {
		t.Error("expected Changed to return true on first call")
	}
}

func TestFingerprintTracker_NoChangeOnSameEntries(t *testing.T) {
	ft := NewFingerprintTracker()
	entries := makeEntries([][2]interface{}{{"tcp", 22}})

	ft.Changed(entries)
	if ft.Changed(entries) {
		t.Error("expected Changed to return false when entries are identical")
	}
}

func TestFingerprintTracker_ChangedAfterReset(t *testing.T) {
	ft := NewFingerprintTracker()
	entries := makeEntries([][2]interface{}{{"tcp", 22}})

	ft.Changed(entries)
	ft.Reset()

	if !ft.Changed(entries) {
		t.Error("expected Changed to return true after Reset")
	}
}

func TestFingerprintTracker_CurrentReflectsLastSeen(t *testing.T) {
	ft := NewFingerprintTracker()
	entries := makeEntries([][2]interface{}{{"udp", 53}})

	ft.Changed(entries)
	expected := Compute(entries)

	if ft.Current() != expected {
		t.Errorf("Current() = %q, want %q", ft.Current(), expected)
	}
}
