package ports

import (
	"testing"
	"time"
)

func fixedDedupClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestDeduplicator_FirstCallNotDuplicate(t *testing.T) {
	d := NewDeduplicator(5 * time.Second)
	if d.IsDuplicate("tcp:8080") {
		t.Fatal("first call should not be a duplicate")
	}
}

func TestDeduplicator_SecondCallWithinWindowIsDuplicate(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(10 * time.Second)
	d.clock = fixedDedupClock(now)

	d.IsDuplicate("tcp:443")
	if !d.IsDuplicate("tcp:443") {
		t.Fatal("second call within window should be a duplicate")
	}
}

func TestDeduplicator_AllowedAfterWindowExpires(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(5 * time.Second)
	d.clock = fixedDedupClock(now)
	d.IsDuplicate("tcp:22")

	// advance clock beyond window
	d.clock = fixedDedupClock(now.Add(6 * time.Second))
	if d.IsDuplicate("tcp:22") {
		t.Fatal("call after window expiry should not be a duplicate")
	}
}

func TestDeduplicator_IndependentKeys(t *testing.T) {
	d := NewDeduplicator(10 * time.Second)
	d.IsDuplicate("tcp:80")

	if d.IsDuplicate("tcp:443") {
		t.Fatal("different key should not be considered duplicate")
	}
}

func TestDeduplicator_Reset(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(10 * time.Second)
	d.clock = fixedDedupClock(now)
	d.IsDuplicate("udp:53")

	d.Reset("udp:53")
	if d.IsDuplicate("udp:53") {
		t.Fatal("after Reset, key should not be a duplicate")
	}
}

func TestDeduplicator_Flush_RemovesExpired(t *testing.T) {
	now := time.Now()
	d := NewDeduplicator(5 * time.Second)
	d.clock = fixedDedupClock(now)

	d.IsDuplicate("tcp:8080")
	d.IsDuplicate("tcp:9090")

	if d.Len() != 2 {
		t.Fatalf("expected 2 tracked keys, got %d", d.Len())
	}

	d.clock = fixedDedupClock(now.Add(10 * time.Second))
	d.Flush()

	if d.Len() != 0 {
		t.Fatalf("expected 0 keys after flush, got %d", d.Len())
	}
}
