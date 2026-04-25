package ports

import (
	"testing"
	"time"
)

func fixedAdaptiveClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestAdaptiveRateLimiter_DefaultRate(t *testing.T) {
	cfg := DefaultAdaptiveRateLimiterConfig()
	al := newAdaptiveRateLimiterWithClock(cfg, fixedAdaptiveClock(time.Now()))
	if al.CurrentRate() != cfg.BaseRate {
		t.Fatalf("expected base rate %.1f, got %.1f", cfg.BaseRate, al.CurrentRate())
	}
}

func TestAdaptiveRateLimiter_AllowFirstCall(t *testing.T) {
	cfg := DefaultAdaptiveRateLimiterConfig()
	al := newAdaptiveRateLimiterWithClock(cfg, fixedAdaptiveClock(time.Now()))
	if !al.Allow() {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAdaptiveRateLimiter_RateDecreasesOnError(t *testing.T) {
	base := time.Now()
	now := base
	clock := func() time.Time { return now }

	cfg := DefaultAdaptiveRateLimiterConfig()
	cfg.AdjustPeriod = 1 * time.Second
	al := newAdaptiveRateLimiterWithClock(cfg, clock)

	for i := 0; i < 5; i++ {
		al.RecordError()
	}

	now = base.Add(2 * time.Second)
	al.Allow() // triggers adjustment

	if al.CurrentRate() >= cfg.BaseRate {
		t.Fatalf("expected rate to decrease below base %.1f, got %.1f", cfg.BaseRate, al.CurrentRate())
	}
}

func TestAdaptiveRateLimiter_RateIncreasesWithoutErrors(t *testing.T) {
	base := time.Now()
	now := base
	clock := func() time.Time { return now }

	cfg := DefaultAdaptiveRateLimiterConfig()
	cfg.BaseRate = 10.0
	cfg.MaxRate = 50.0
	cfg.AdjustPeriod = 1 * time.Second
	al := newAdaptiveRateLimiterWithClock(cfg, clock)

	now = base.Add(2 * time.Second)
	al.Allow()

	if al.CurrentRate() <= cfg.BaseRate {
		t.Fatalf("expected rate to increase above base %.1f, got %.1f", cfg.BaseRate, al.CurrentRate())
	}
}

func TestAdaptiveRateLimiter_RateClampedToMin(t *testing.T) {
	base := time.Now()
	now := base
	clock := func() time.Time { return now }

	cfg := DefaultAdaptiveRateLimiterConfig()
	cfg.BaseRate = 2.0
	cfg.MinRate = 1.0
	cfg.AdjustPeriod = 1 * time.Second
	al := newAdaptiveRateLimiterWithClock(cfg, clock)

	for cycle := 0; cycle < 10; cycle++ {
		for i := 0; i < 20; i++ {
			al.RecordError()
		}
		now = now.Add(2 * time.Second)
		al.Allow()
	}

	if al.CurrentRate() < cfg.MinRate {
		t.Fatalf("rate %.2f dropped below min %.2f", al.CurrentRate(), cfg.MinRate)
	}
}

func TestAdaptiveRateLimiter_RateClampedToMax(t *testing.T) {
	base := time.Now()
	now := base
	clock := func() time.Time { return now }

	cfg := DefaultAdaptiveRateLimiterConfig()
	cfg.BaseRate = 40.0
	cfg.MaxRate = 50.0
	cfg.AdjustPeriod = 1 * time.Second
	al := newAdaptiveRateLimiterWithClock(cfg, clock)

	for cycle := 0; cycle < 20; cycle++ {
		now = now.Add(2 * time.Second)
		al.Allow()
	}

	if al.CurrentRate() > cfg.MaxRate {
		t.Fatalf("rate %.2f exceeded max %.2f", al.CurrentRate(), cfg.MaxRate)
	}
}
