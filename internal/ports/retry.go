package ports

import (
	"context"
	"time"
)

// RetryPolicy defines how retries behave.
type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// DefaultRetryPolicy returns a sensible default retry policy.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   200 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		Multiplier:  2.0,
	}
}

// Retrier executes a function with retry logic using exponential backoff.
type Retrier struct {
	policy RetryPolicy
	sleep  func(time.Duration)
}

// NewRetrier creates a Retrier with the given policy.
func NewRetrier(policy RetryPolicy) *Retrier {
	return &Retrier{
		policy: policy,
		sleep:  time.Sleep,
	}
}

// Do executes fn, retrying on non-nil error up to MaxAttempts times.
// Returns the last error if all attempts fail.
func (r *Retrier) Do(ctx context.Context, fn func() error) error {
	delay := r.policy.BaseDelay
	var err error
	for attempt := 0; attempt < r.policy.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = fn()
		if err == nil {
			return nil
		}
		if attempt < r.policy.MaxAttempts-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay = time.Duration(float64(delay) * r.policy.Multiplier)
			if delay > r.policy.MaxDelay {
				delay = r.policy.MaxDelay
			}
		}
	}
	return err
}

// Attempts returns the configured maximum number of attempts.
func (r *Retrier) Attempts() int {
	return r.policy.MaxAttempts
}
