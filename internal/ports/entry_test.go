package ports

import "testing"

func TestPortEntry_Key(t *testing.T) {
	tests := []struct {
		entry PortEntry
		want  string
	}{
		{PortEntry{Proto: "tcp", Port: 80}, "tcp:80"},
		{PortEntry{Proto: "udp", Port: 53}, "udp:53"},
		{PortEntry{Proto: "tcp", Port: 443, PID: 1234, Process: "nginx"}, "tcp:443"},
	}
	for _, tc := range tests {
		got := tc.entry.Key()
		if got != tc.want {
			t.Errorf("Key() = %q, want %q", got, tc.want)
		}
	}
}

func TestPortEntry_KeyUniqueness(t *testing.T) {
	a := PortEntry{Proto: "tcp", Port: 80}
	b := PortEntry{Proto: "udp", Port: 80}
	if a.Key() == b.Key() {
		t.Error("tcp:80 and udp:80 should have different keys")
	}
}
