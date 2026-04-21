package ports

import (
	"sync"
	"time"
)

// HealthStatus represents the current health state of the scanner.
type HealthStatus struct {
	Healthy      bool
	LastScanTime time.Time
	LastError    error
	ConsecErrors int
	Uptime       time.Duration
	StartTime    time.Time
}

// HealthChecker tracks scanner liveness and reports degraded state.
type HealthChecker struct {
	mu           sync.RWMutex
	clock        func() time.Time
	startTime    time.Time
	lastScanTime time.Time
	lastError    error
	consecErrors int
	maxErrors    int
	staleness    time.Duration
}

// NewHealthChecker returns a HealthChecker with the given thresholds.
// maxErrors is the number of consecutive errors before reporting unhealthy.
// staleness is how long without a scan before reporting unhealthy.
func NewHealthChecker(maxErrors int, staleness time.Duration) *HealthChecker {
	now := time.Now()
	return &HealthChecker{
		clock:     time.Now,
		startTime: now,
		maxErrors: maxErrors,
		staleness: staleness,
	}
}

// RecordScan records the result of a scan tick.
func (h *HealthChecker) RecordScan(err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastScanTime = h.clock()
	if err != nil {
		h.lastError = err
		h.consecErrors++
	} else {
		h.lastError = nil
		h.consecErrors = 0
	}
}

// Status returns a snapshot of the current health state.
func (h *HealthChecker) Status() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	now := h.clock()
	stale := !h.lastScanTime.IsZero() && now.Sub(h.lastScanTime) > h.staleness
	healthy := h.consecErrors < h.maxErrors && !stale
	return HealthStatus{
		Healthy:      healthy,
		LastScanTime: h.lastScanTime,
		LastError:    h.lastError,
		ConsecErrors: h.consecErrors,
		Uptime:       now.Sub(h.startTime),
		StartTime:    h.startTime,
	}
}

// Reset clears error state, e.g. after a manual recovery.
func (h *HealthChecker) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.consecErrors = 0
	h.lastError = nil
}
