package ports

import (
	"testing"
	"time"
)

var fixedBudgetClock = func() func() time.Time {
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	return func() time.Time { return now }
}()

func budgetPolicy() BudgetPolicy {
	return BudgetPolicy{
		MaxScansPerWindow: 3,
		Window:            time.Minute,
	}
}

func TestBudget_AllowsUpToMax(t *testing.T) {
	b := newBudgetTrackerWithClock(budgetPolicy(), fixedBudgetClock)
	for i := 0; i < 3; i++ {
		if !b.Record("host") {
			t.Fatalf("expected record %d to be allowed", i)
		}
	}
}

func TestBudget_BlocksAfterMax(t *testing.T) {
	b := newBudgetTrackerWithClock(budgetPolicy(), fixedBudgetClock)
	for i := 0; i < 3; i++ {
		b.Record("host")
	}
	if b.Record("host") {
		t.Fatal("expected record to be denied after max")
	}
}

func TestBudget_RemainingDecrements(t *testing.T) {
	b := newBudgetTrackerWithClock(budgetPolicy(), fixedBudgetClock)
	if b.Remaining("host") != 3 {
		t.Fatalf("expected 3 remaining initially")
	}
	b.Record("host")
	if b.Remaining("host") != 2 {
		t.Fatalf("expected 2 remaining after one record")
	}
}

func TestBudget_OldEventsExpire(t *testing.T) {
	var now time.Time
	clock := func() time.Time { return now }

	now = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	b := newBudgetTrackerWithClock(budgetPolicy(), clock)
	for i := 0; i < 3; i++ {
		b.Record("host")
	}

	// Advance past the window
	now = now.Add(2 * time.Minute)
	if !b.Record("host") {
		t.Fatal("expected record to be allowed after window expiry")
	}
}

func TestBudget_IndependentKeys(t *testing.T) {
	b := newBudgetTrackerWithClock(budgetPolicy(), fixedBudgetClock)
	for i := 0; i < 3; i++ {
		b.Record("a")
	}
	if !b.Record("b") {
		t.Fatal("key b should have its own budget")
	}
}

func TestBudget_ResetClearsState(t *testing.T) {
	b := newBudgetTrackerWithClock(budgetPolicy(), fixedBudgetClock)
	for i := 0; i < 3; i++ {
		b.Record("host")
	}
	b.Reset("host")
	if !b.Record("host") {
		t.Fatal("expected record to be allowed after reset")
	}
}
