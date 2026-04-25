package ports

import (
	"testing"
	"time"
)

func TestBudget_ExhaustThenExpireAndRecover(t *testing.T) {
	var now time.Time
	clock := func() time.Time { return now }

	policy := BudgetPolicy{
		MaxScansPerWindow: 2,
		Window:            30 * time.Second,
	}
	b := newBudgetTrackerWithClock(policy, clock)

	now = time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

	if !b.Record("svc") {
		t.Fatal("first record should be allowed")
	}
	if !b.Record("svc") {
		t.Fatal("second record should be allowed")
	}
	if b.Record("svc") {
		t.Fatal("third record should be denied")
	}
	if b.Remaining("svc") != 0 {
		t.Fatalf("expected 0 remaining, got %d", b.Remaining("svc"))
	}

	// Advance past the window — old events expire
	now = now.Add(31 * time.Second)

	if b.Remaining("svc") != 2 {
		t.Fatalf("expected full budget after window, got %d", b.Remaining("svc"))
	}
	if !b.Record("svc") {
		t.Fatal("record should be allowed after window expiry")
	}
}

func TestBudget_MultipleKeysIsolated(t *testing.T) {
	var now time.Time
	clock := func() time.Time { return now }
	now = time.Date(2024, 6, 1, 10, 0, 0, 0, time.UTC)

	policy := BudgetPolicy{MaxScansPerWindow: 1, Window: time.Minute}
	b := newBudgetTrackerWithClock(policy, clock)

	b.Record("alpha")

	if !b.Record("beta") {
		t.Fatal("beta should have independent budget from alpha")
	}
	if b.Record("alpha") {
		t.Fatal("alpha should be exhausted")
	}
}
