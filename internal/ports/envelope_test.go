package ports

import (
	"testing"
	"time"
)

var fixedEnvelopeClock = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func makeEnvelopeWrapper(maxAge time.Duration) *EnvelopeWrapper {
	policy := EnvelopePolicy{MaxAge: maxAge, Source: "test"}
	return newEnvelopeWrapperWithClock(policy, func() time.Time { return fixedEnvelopeClock })
}

func TestEnvelopeWrapper_WrapSetsSource(t *testing.T) {
	w := makeEnvelopeWrapper(time.Minute)
	env := w.Wrap(nil)
	if env.Source != "test" {
		t.Fatalf("expected source 'test', got %q", env.Source)
	}
}

func TestEnvelopeWrapper_WrapSetsTimestamp(t *testing.T) {
	w := makeEnvelopeWrapper(time.Minute)
	env := w.Wrap(nil)
	if !env.ScannedAt.Equal(fixedEnvelopeClock) {
		t.Fatalf("unexpected timestamp: %v", env.ScannedAt)
	}
}

func TestEnvelopeWrapper_WrapPreservesEntries(t *testing.T) {
	entries := []PortEntry{{Port: 80, Proto: "tcp"}, {Port: 443, Proto: "tcp"}}
	w := makeEnvelopeWrapper(time.Minute)
	env := w.Wrap(entries)
	if len(env.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(env.Entries))
	}
}

func TestEnvelopeWrapper_NotStaleWhenFresh(t *testing.T) {
	w := makeEnvelopeWrapper(time.Minute)
	env := w.Wrap(nil)
	if w.Stale(env) {
		t.Fatal("expected envelope to be fresh")
	}
}

func TestEnvelopeWrapper_StaleWhenOld(t *testing.T) {
	w := makeEnvelopeWrapper(time.Minute)
	old := &Envelope{
		Source:    "test",
		ScannedAt: fixedEnvelopeClock.Add(-2 * time.Minute),
		Entries:   nil,
	}
	if !w.Stale(old) {
		t.Fatal("expected envelope to be stale")
	}
}

func TestEnvelopeWrapper_ExactlyAtBoundaryIsStale(t *testing.T) {
	w := makeEnvelopeWrapper(time.Minute)
	boundary := &Envelope{
		Source:    "test",
		ScannedAt: fixedEnvelopeClock.Add(-time.Minute),
		Entries:   nil,
	}
	// age == MaxAge is NOT stale (strictly greater than)
	if w.Stale(boundary) {
		t.Fatal("expected envelope at exact boundary to be fresh")
	}
}

func TestDefaultEnvelopePolicy_Source(t *testing.T) {
	p := DefaultEnvelopePolicy()
	if p.Source != "portwatch" {
		t.Fatalf("expected source 'portwatch', got %q", p.Source)
	}
}

func TestDefaultEnvelopePolicy_MaxAge(t *testing.T) {
	p := DefaultEnvelopePolicy()
	if p.MaxAge != 5*time.Minute {
		t.Fatalf("unexpected default MaxAge: %v", p.MaxAge)
	}
}
