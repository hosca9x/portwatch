package ports

import (
	"errors"
	"testing"
	"time"
)

var fixedHealthClock = func() time.Time {
	return time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
}

func newTestHealthChecker(maxErrors int, staleness time.Duration) *HealthChecker {
	hc := NewHealthChecker(maxErrors, staleness)
	hc.clock = fixedHealthClock
	hc.startTime = fixedHealthClock()
	return hc
}

func TestHealthChecker_InitiallyHealthy(t *testing.T) {
	hc := newTestHealthChecker(3, 10*time.Second)
	// No scans yet — staleness check requires a non-zero lastScanTime, so healthy.
	st := hc.Status()
	if !st.Healthy {
		t.Errorf("expected healthy before any scan, got unhealthy")
	}
}

func TestHealthChecker_UnhealthyAfterMaxErrors(t *testing.T) {
	hc := newTestHealthChecker(3, time.Minute)
	err := errors.New("scan failed")
	for i := 0; i < 3; i++ {
		hc.RecordScan(err)
	}
	st := hc.Status()
	if st.Healthy {
		t.Errorf("expected unhealthy after %d errors", 3)
	}
	if st.ConsecErrors != 3 {
		t.Errorf("expected 3 consec errors, got %d", st.ConsecErrors)
	}
}

func TestHealthChecker_RecoveryResetsErrors(t *testing.T) {
	hc := newTestHealthChecker(3, time.Minute)
	err := errors.New("oops")
	hc.RecordScan(err)
	hc.RecordScan(err)
	hc.RecordScan(nil) // success clears streak
	st := hc.Status()
	if !st.Healthy {
		t.Errorf("expected healthy after successful scan")
	}
	if st.ConsecErrors != 0 {
		t.Errorf("expected 0 consec errors, got %d", st.ConsecErrors)
	}
}

func TestHealthChecker_UnhealthyWhenStale(t *testing.T) {
	hc := newTestHealthChecker(3, 5*time.Second)
	// Record a scan at fixed time, then advance clock past staleness window.
	hc.RecordScan(nil)
	hc.clock = func() time.Time {
		return fixedHealthClock().Add(10 * time.Second)
	}
	st := hc.Status()
	if st.Healthy {
		t.Errorf("expected unhealthy when scan is stale")
	}
}

func TestHealthChecker_Reset(t *testing.T) {
	hc := newTestHealthChecker(2, time.Minute)
	hc.RecordScan(errors.New("err"))
	hc.RecordScan(errors.New("err"))
	if hc.Status().Healthy {
		t.Fatal("should be unhealthy before reset")
	}
	hc.Reset()
	// After reset consec errors cleared; still need a scan to not be stale.
	st := hc.Status()
	if st.ConsecErrors != 0 {
		t.Errorf("expected 0 after reset, got %d", st.ConsecErrors)
	}
}

func TestHealthChecker_UptimeIncreases(t *testing.T) {
	hc := newTestHealthChecker(3, time.Minute)
	hc.clock = func() time.Time {
		return fixedHealthClock().Add(30 * time.Second)
	}
	st := hc.Status()
	if st.Uptime < 30*time.Second {
		t.Errorf("expected uptime >= 30s, got %v", st.Uptime)
	}
}
