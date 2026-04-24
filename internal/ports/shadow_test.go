package ports

import (
	"testing"
	"time"
)

var fixedShadowEpoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func fixedShadowClock(offset time.Duration) func() time.Time {
	return func() time.Time { return fixedShadowEpoch.Add(offset) }
}

func makeShadowEntry(port int, proto string) PortEntry {
	return PortEntry{Port: port, Proto: proto}
}

func TestShadow_BelowThresholdReportedAsShadow(t *testing.T) {
	cfg := ShadowConfig{MaxLifetime: time.Minute, MinAppearances: 3}
	var tick time.Duration
	st := newShadowTrackerWithClock(cfg, func() time.Time { return fixedShadowEpoch.Add(tick) })

	e := makeShadowEntry(8080, "tcp")
	st.Record(e) // seen once

	tick = 2 * time.Minute // past lifetime
	shadows := st.Shadows()
	if len(shadows) != 1 {
		t.Fatalf("expected 1 shadow, got %d", len(shadows))
	}
	if shadows[0].Port != 8080 {
		t.Errorf("expected port 8080, got %d", shadows[0].Port)
	}
}

func TestShadow_AboveThresholdNotReported(t *testing.T) {
	cfg := ShadowConfig{MaxLifetime: time.Minute, MinAppearances: 3}
	var tick time.Duration
	st := newShadowTrackerWithClock(cfg, func() time.Time { return fixedShadowEpoch.Add(tick) })

	e := makeShadowEntry(443, "tcp")
	for i := 0; i < 3; i++ {
		st.Record(e)
	}

	tick = 2 * time.Minute
	if shadows := st.Shadows(); len(shadows) != 0 {
		t.Errorf("expected no shadows, got %d", len(shadows))
	}
}

func TestShadow_WithinLifetimeNotYetReported(t *testing.T) {
	cfg := ShadowConfig{MaxLifetime: time.Minute, MinAppearances: 3}
	var tick time.Duration
	st := newShadowTrackerWithClock(cfg, func() time.Time { return fixedShadowEpoch.Add(tick) })

	st.Record(makeShadowEntry(9000, "udp"))
	tick = 30 * time.Second // still within window

	if shadows := st.Shadows(); len(shadows) != 0 {
		t.Errorf("expected no shadows yet, got %d", len(shadows))
	}
}

func TestShadow_EvictClearsStaleEntries(t *testing.T) {
	cfg := ShadowConfig{MaxLifetime: time.Minute, MinAppearances: 3}
	var tick time.Duration
	st := newShadowTrackerWithClock(cfg, func() time.Time { return fixedShadowEpoch.Add(tick) })

	st.Record(makeShadowEntry(1234, "tcp"))
	tick = 2 * time.Minute
	st.Evict()

	if shadows := st.Shadows(); len(shadows) != 0 {
		t.Errorf("expected evicted entry to be gone, got %d", len(shadows))
	}
}

func TestShadow_IndependentKeys(t *testing.T) {
	cfg := ShadowConfig{MaxLifetime: time.Minute, MinAppearances: 2}
	var tick time.Duration
	st := newShadowTrackerWithClock(cfg, func() time.Time { return fixedShadowEpoch.Add(tick) })

	st.Record(makeShadowEntry(80, "tcp"))
	st.Record(makeShadowEntry(80, "tcp"))
	st.Record(makeShadowEntry(53, "udp")) // only once

	tick = 2 * time.Minute
	shadows := st.Shadows()
	if len(shadows) != 1 {
		t.Fatalf("expected 1 shadow, got %d", len(shadows))
	}
	if shadows[0].Proto != "udp" {
		t.Errorf("expected udp shadow, got %s", shadows[0].Proto)
	}
}
