package alert

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/portwatch/internal/ports"
)

func makeAdaptiveAlerter(threshold float64) (*ports.AdaptiveRateLimiter, *AdaptiveAlerter, *bytes.Buffer) {
	cfg := ports.DefaultAdaptiveRateLimiterConfig()
	cfg.BaseRate = 10.0
	limiter := ports.NewAdaptiveRateLimiter(cfg)
	var buf bytes.Buffer
	alerter := NewAdaptiveAlerter(limiter, threshold, &buf)
	return limiter, alerter, &buf
}

func TestAdaptiveAlert_NoOutputWhenRateStable(t *testing.T) {
	_, alerter, buf := makeAdaptiveAlerter(0.2)
	alerter.Check()
	alerter.Check()
	if buf.Len() != 0 {
		t.Fatalf("expected no output for stable rate, got: %s", buf.String())
	}
}

func TestAdaptiveAlert_DefaultsToStdout(t *testing.T) {
	cfg := ports.DefaultAdaptiveRateLimiterConfig()
	limiter := ports.NewAdaptiveRateLimiter(cfg)
	alerter := NewAdaptiveAlerter(limiter, 0.2, nil)
	if alerter.out == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestAdaptiveAlert_DefaultThresholdApplied(t *testing.T) {
	cfg := ports.DefaultAdaptiveRateLimiterConfig()
	limiter := ports.NewAdaptiveRateLimiter(cfg)
	var buf bytes.Buffer
	alerter := NewAdaptiveAlerter(limiter, 0, &buf)
	if alerter.changeThresh <= 0 {
		t.Fatal("expected positive default threshold")
	}
}

func TestAdaptiveAlert_EmitsOnRateIncrease(t *testing.T) {
	cfg := ports.DefaultAdaptiveRateLimiterConfig()
	cfg.BaseRate = 10.0
	limiter := ports.NewAdaptiveRateLimiter(cfg)
	var buf bytes.Buffer
	alerter := NewAdaptiveAlerter(limiter, 0.05, &buf)

	// Manually prime lastRate low so next Check sees a big jump.
	alerter.lastRate = 5.0
	// Simulate that the limiter now reports a higher rate by using a fresh
	// limiter whose base is above lastRate by >5%.
	alerter.Check()

	if buf.Len() == 0 {
		t.Fatal("expected alert output on rate increase")
	}
	if !strings.Contains(buf.String(), "increased") {
		t.Errorf("expected 'increased' in output, got: %s", buf.String())
	}
}

func TestAdaptiveAlert_EmitsOnRateDecrease(t *testing.T) {
	cfg := ports.DefaultAdaptiveRateLimiterConfig()
	cfg.BaseRate = 10.0
	limiter := ports.NewAdaptiveRateLimiter(cfg)
	var buf bytes.Buffer
	alerter := NewAdaptiveAlerter(limiter, 0.05, &buf)

	// Prime lastRate high so current rate looks like a decrease.
	alerter.lastRate = 50.0
	alerter.Check()

	if buf.Len() == 0 {
		t.Fatal("expected alert output on rate decrease")
	}
	if !strings.Contains(buf.String(), "decreased") {
		t.Errorf("expected 'decreased' in output, got: %s", buf.String())
	}
}

func TestAdaptiveAlert_LineContainsRateValues(t *testing.T) {
	cfg := ports.DefaultAdaptiveRateLimiterConfig()
	cfg.BaseRate = 10.0
	limiter := ports.NewAdaptiveRateLimiter(cfg)
	var buf bytes.Buffer
	alerter := NewAdaptiveAlerter(limiter, 0.05, &buf)
	alerter.lastRate = 5.0
	alerter.Check()

	line := buf.String()
	if !strings.Contains(line, "adaptive-rate-limiter") {
		t.Errorf("expected label in output, got: %s", line)
	}
	if !strings.Contains(line, "10.00") {
		t.Errorf("expected current rate in output, got: %s", line)
	}
}
