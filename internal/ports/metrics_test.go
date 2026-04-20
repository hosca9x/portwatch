package ports

import (
	"errors"
	"testing"
	"time"
)

func fixedMetricsClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestCollector_InitialState(t *testing.T) {
	c := NewCollector()
	snap := c.Snapshot()

	if snap.TotalScans != 0 || snap.TotalNew != 0 || snap.TotalClosed != 0 || snap.Errors != 0 {
		t.Errorf("expected zero initial metrics, got %+v", snap)
	}
}

func TestCollector_RecordScan_NoError(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	c := &Collector{clock: fixedMetricsClock(now)}

	c.RecordScan(42, 3, 1, nil)
	snap := c.Snapshot()

	if snap.TotalScans != 1 {
		t.Errorf("TotalScans: want 1, got %d", snap.TotalScans)
	}
	if snap.TotalNew != 3 {
		t.Errorf("TotalNew: want 3, got %d", snap.TotalNew)
	}
	if snap.TotalClosed != 1 {
		t.Errorf("TotalClosed: want 1, got %d", snap.TotalClosed)
	}
	if snap.LastScanPorts != 42 {
		t.Errorf("LastScanPorts: want 42, got %d", snap.LastScanPorts)
	}
	if !snap.LastScanAt.Equal(now) {
		t.Errorf("LastScanAt: want %v, got %v", now, snap.LastScanAt)
	}
	if snap.Errors != 0 {
		t.Errorf("Errors: want 0, got %d", snap.Errors)
	}
}

func TestCollector_RecordScan_WithError(t *testing.T) {
	c := NewCollector()
	c.RecordScan(0, 0, 0, errors.New("scan failed"))

	snap := c.Snapshot()
	if snap.Errors != 1 {
		t.Errorf("Errors: want 1, got %d", snap.Errors)
	}
}

func TestCollector_Accumulates(t *testing.T) {
	c := NewCollector()
	c.RecordScan(10, 2, 0, nil)
	c.RecordScan(12, 1, 3, nil)
	c.RecordScan(9, 0, 0, errors.New("oops"))

	snap := c.Snapshot()
	if snap.TotalScans != 3 {
		t.Errorf("TotalScans: want 3, got %d", snap.TotalScans)
	}
	if snap.TotalNew != 3 {
		t.Errorf("TotalNew: want 3, got %d", snap.TotalNew)
	}
	if snap.TotalClosed != 3 {
		t.Errorf("TotalClosed: want 3, got %d", snap.TotalClosed)
	}
	if snap.Errors != 1 {
		t.Errorf("Errors: want 1, got %d", snap.Errors)
	}
}

func TestCollector_Reset(t *testing.T) {
	c := NewCollector()
	c.RecordScan(5, 1, 1, nil)
	c.Reset()

	snap := c.Snapshot()
	if snap.TotalScans != 0 || snap.TotalNew != 0 || snap.Errors != 0 {
		t.Errorf("expected zeroed metrics after Reset, got %+v", snap)
	}
}
