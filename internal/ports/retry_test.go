package ports

import (
	"context"
	"errors"
	"testing"
	"time"
)

func defaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		Multiplier:  2.0,
	}
}

func TestRetrier_SucceedsOnFirstAttempt(t *testing.T) {
	r := NewRetrier(defaultRetryPolicy())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestRetrier_RetriesOnError(t *testing.T) {
	r := NewRetrier(defaultRetryPolicy())
	calls := 0
	sentinel := errors.New("transient")
	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return sentinel
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil after retry, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetrier_ExhaustsAttempts(t *testing.T) {
	r := NewRetrier(defaultRetryPolicy())
	sentinel := errors.New("permanent")
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestRetrier_RespectsContextCancel(t *testing.T) {
	policy := defaultRetryPolicy()
	policy.BaseDelay = 500 * time.Millisecond
	r := NewRetrier(policy)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.Do(ctx, func() error {
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestRetrier_Attempts(t *testing.T) {
	r := NewRetrier(defaultRetryPolicy())
	if r.Attempts() != 3 {
		t.Fatalf("expected 3, got %d", r.Attempts())
	}
}

func TestDefaultRetryPolicy(t *testing.T) {
	p := DefaultRetryPolicy()
	if p.MaxAttempts <= 0 {
		t.Fatal("MaxAttempts must be positive")
	}
	if p.BaseDelay <= 0 {
		t.Fatal("BaseDelay must be positive")
	}
	if p.Multiplier < 1.0 {
		t.Fatal("Multiplier must be >= 1.0")
	}
}
