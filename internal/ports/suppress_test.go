package ports

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSuppressor_FirstCallNotSuppressed(t *testing.T) {
	s := NewSuppressor(5 * time.Second)
	if s.IsSuppressed("tcp:8080") {
		t.Fatal("first call should not be suppressed")
	}
}

func TestSuppressor_SecondCallWithinWindowSuppressed(t *testing.T) {
	now := time.Now()
	s := NewSuppressor(5 * time.Second)
	s.now = fixedClock(now)

	s.IsSuppressed("tcp:8080") // record it
	if !s.IsSuppressed("tcp:8080") {
		t.Fatal("second call within window should be suppressed")
	}
}

func TestSuppressor_AllowedAfterWindowExpires(t *testing.T) {
	now := time.Now()
	s := NewSuppressor(5 * time.Second)
	s.now = fixedClock(now)
	s.IsSuppressed("tcp:9090")

	// advance clock beyond window
	s.now = fixedClock(now.Add(6 * time.Second))
	if s.IsSuppressed("tcp:9090") {
		t.Fatal("call after window expiry should not be suppressed")
	}
}

func TestSuppressor_IndependentKeys(t *testing.T) {
	s := NewSuppressor(10 * time.Second)
	s.IsSuppressed("tcp:80")

	if s.IsSuppressed("udp:53") {
		t.Fatal("different key should not be suppressed")
	}
}

func TestSuppressor_Reset(t *testing.T) {
	now := time.Now()
	s := NewSuppressor(10 * time.Second)
	s.now = fixedClock(now)
	s.IsSuppressed("tcp:443")

	s.Reset("tcp:443")
	if s.IsSuppressed("tcp:443") {
		t.Fatal("reset key should not be suppressed on next call")
	}
}

func TestSuppressor_Flush(t *testing.T) {
	now := time.Now()
	s := NewSuppressor(5 * time.Second)
	s.now = fixedClock(now)
	s.IsSuppressed("tcp:22")
	s.IsSuppressed("tcp:80")

	// advance past window so both records expire
	s.now = fixedClock(now.Add(6 * time.Second))
	s.Flush()

	if s.IsSuppressed("tcp:22") {
		t.Fatal("flushed key should not be suppressed")
	}
}
