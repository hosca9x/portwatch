package ports

import (
	"testing"
	"time"
)

func fixedDecayClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestDecayTracker_NewPortWeightIsOne(t *testing.T) {
	d := NewDecayTracker(30 * time.Minute)
	d.clock = fixedDecayClock(time.Unix(1000, 0))

	d.Observe("tcp:80")
	r := d.Report("tcp:80")

	if r.Weight != 1.0 {
		t.Errorf("expected weight 1.0 for brand-new port, got %f", r.Weight)
	}
	if r.IsStable {
		t.Error("expected IsStable=false for brand-new port")
	}
}

func TestDecayTracker_WeightDecreasesOverTime(t *testing.T) {
	base := time.Unix(0, 0)
	d := NewDecayTracker(30 * time.Minute)
	d.clock = fixedDecayClock(base)

	d.Observe("tcp:443")

	// Advance clock by one half-life
	d.clock = fixedDecayClock(base.Add(30 * time.Minute))
	r := d.Report("tcp:443")

	if r.Weight >= 1.0 {
		t.Errorf("expected weight < 1.0 after half-life, got %f", r.Weight)
	}
	if r.Age != 30*time.Minute {
		t.Errorf("expected age=30m, got %v", r.Age)
	}
}

func TestDecayTracker_StableAfterTwoHalfLives(t *testing.T) {
	base := time.Unix(0, 0)
	d := NewDecayTracker(30 * time.Minute)
	d.clock = fixedDecayClock(base)

	d.Observe("udp:53")

	// Two half-lives => weight = 0.25, which is at the threshold
	d.clock = fixedDecayClock(base.Add(60 * time.Minute))
	r := d.Report("udp:53")

	if !r.IsStable {
		t.Errorf("expected IsStable=true after two half-lives, weight=%f", r.Weight)
	}
}

func TestDecayTracker_UnobservedKeyDefaultsToOne(t *testing.T) {
	d := NewDecayTracker(0) // uses default half-life
	r := d.Report("tcp:9999")

	if r.Weight != 1.0 {
		t.Errorf("expected weight=1.0 for unknown key, got %f", r.Weight)
	}
	if r.Age != 0 {
		t.Errorf("expected age=0 for unknown key, got %v", r.Age)
	}
}

func TestDecayTracker_ForgetResetsKey(t *testing.T) {
	base := time.Unix(0, 0)
	d := NewDecayTracker(30 * time.Minute)
	d.clock = fixedDecayClock(base)

	d.Observe("tcp:22")
	d.clock = fixedDecayClock(base.Add(60 * time.Minute))
	d.Forget("tcp:22")

	r := d.Report("tcp:22")
	if r.Weight != 1.0 {
		t.Errorf("expected weight=1.0 after Forget, got %f", r.Weight)
	}
}

func TestDecayTracker_ObserveDoesNotResetTimestamp(t *testing.T) {
	base := time.Unix(0, 0)
	d := NewDecayTracker(30 * time.Minute)
	d.clock = fixedDecayClock(base)

	d.Observe("tcp:8080")

	d.clock = fixedDecayClock(base.Add(15 * time.Minute))
	d.Observe("tcp:8080") // second observe should not reset

	d.clock = fixedDecayClock(base.Add(30 * time.Minute))
	r := d.Report("tcp:8080")

	// Age should be 30m from original observe, not 15m from second
	if r.Age != 30*time.Minute {
		t.Errorf("expected age=30m, got %v", r.Age)
	}
}

func TestDecayTracker_WeightApproximatelyHalfAfterOneHalfLife(t *testing.T) {
	base := time.Unix(0, 0)
	d := NewDecayTracker(30 * time.Minute)
	d.clock = fixedDecayClock(base)

	d.Observe("tcp:8443")

	d.clock = fixedDecayClock(base.Add(30 * time.Minute))
	r := d.Report("tcp:8443")

	// After exactly one half-life the weight should be 0.5
	const want = 0.5
	const tolerance = 0.001
	if r.Weight < want-tolerance || r.Weight > want+tolerance {
		t.Errorf("expected weight≈0.5 after one half-life, got %f", r.Weight)
	}
}
