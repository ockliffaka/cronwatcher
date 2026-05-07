package retry

import (
	"context"
	"time"
)

// Policy defines how retries are attempted.
type Policy struct {
	MaxAttempts int
	Delay       time.Duration
	Multiplier  float64
}

// DefaultPolicy returns a sensible default retry policy.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 3,
		Delay:       2 * time.Second,
		Multiplier:  2.0,
	}
}

// Result holds the outcome of a retried operation.
type Result struct {
	Attempts int
	Err      error
}

// Do executes fn according to the policy, retrying on non-nil errors.
// It respects context cancellation between attempts.
func Do(ctx context.Context, p Policy, fn func() error) Result {
	if p.MaxAttempts < 1 {
		p.MaxAttempts = 1
	}
	if p.Multiplier <= 0 {
		p.Multiplier = 1.0
	}

	delay := p.Delay
	var lastErr error

	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		lastErr = fn()
		if lastErr == nil {
			return Result{Attempts: attempt, Err: nil}
		}

		if attempt == p.MaxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return Result{Attempts: attempt, Err: ctx.Err()}
		case <-time.After(delay):
			delay = time.Duration(float64(delay) * p.Multiplier)
		}
	}

	return Result{Attempts: p.MaxAttempts, Err: lastErr}
}
