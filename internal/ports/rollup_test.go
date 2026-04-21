package ports

import (
	"testing"
	"time"
)

func fixedRollupClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func makeRollupEntry(port int) PortEntry {
	return PortEntry{Port: port, Proto: "tcp"}
}

func TestRollup_AddCreatesNewBucket(t *testing.T) {
	now := time.Now()
	r := newRollupWithClock(DefaultRollupPolicy(), fixedRollupClock(now))
	r.Add("tcp:80", makeRollupEntry(80))
	keys := r.Keys()
	if len(keys) != 1 || keys[0] != "tcp:80" {
		t.Fatalf("expected key tcp:80, got %v", keys)
	}
}

func TestRollup_FlushReturnsEntries(t *testing.T) {
	now := time.Now()
	r := newRollupWithClock(DefaultRollupPolicy(), fixedRollupClock(now))
	r.Add("tcp:443", makeRollupEntry(443))
	r.Add("tcp:443", makeRollupEntry(443))
	b := r.Flush("tcp:443")
	if b == nil {
		t.Fatal("expected non-nil bucket")
	}
	if len(b.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(b.Entries))
	}
}

func TestRollup_FlushNilWhenMissing(t *testing.T) {
	r := NewRollup(DefaultRollupPolicy())
	b := r.Flush("tcp:9999")
	if b != nil {
		t.Fatal("expected nil for missing key")
	}
}

func TestRollup_FlushClearsBucket(t *testing.T) {
	now := time.Now()
	r := newRollupWithClock(DefaultRollupPolicy(), fixedRollupClock(now))
	r.Add("tcp:22", makeRollupEntry(22))
	r.Flush("tcp:22")
	if len(r.Keys()) != 0 {
		t.Fatal("expected empty keys after flush")
	}
}

func TestRollup_WindowResetOnExpiry(t *testing.T) {
	base := time.Now()
	advance := base
	r := newRollupWithClock(RollupPolicy{Window: 5 * time.Second, MaxItems: 10}, func() time.Time { return advance })
	r.Add("tcp:8080", makeRollupEntry(8080))
	advance = base.Add(10 * time.Second)
	r.Add("tcp:8080", makeRollupEntry(8080))
	b := r.Flush("tcp:8080")
	if b == nil || len(b.Entries) != 1 {
		t.Fatalf("expected 1 entry after window reset, got %v", b)
	}
}

func TestRollup_RespectsMaxItems(t *testing.T) {
	now := time.Now()
	r := newRollupWithClock(RollupPolicy{Window: time.Minute, MaxItems: 3}, fixedRollupClock(now))
	for i := 0; i < 10; i++ {
		r.Add("tcp:80", makeRollupEntry(80))
	}
	b := r.Flush("tcp:80")
	if len(b.Entries) != 3 {
		t.Fatalf("expected 3 entries (max), got %d", len(b.Entries))
	}
}
