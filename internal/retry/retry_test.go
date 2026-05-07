package retry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cronwatcher/internal/retry"
)

var errFail = errors.New("operation failed")

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, Delay: 0, Multiplier: 1.0}
	res := retry.Do(context.Background(), p, func() error {
		return nil
	})
	if res.Err != nil {
		t.Fatalf("expected no error, got %v", res.Err)
	}
	if res.Attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", res.Attempts)
	}
}

func TestDo_RetriesOnFailure(t *testing.T) {
	calls := 0
	p := retry.Policy{MaxAttempts: 3, Delay: 0, Multiplier: 1.0}
	res := retry.Do(context.Background(), p, func() error {
		calls++
		if calls < 3 {
			return errFail
		}
		return nil
	})
	if res.Err != nil {
		t.Fatalf("expected success after retries, got %v", res.Err)
	}
	if res.Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", res.Attempts)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	p := retry.Policy{MaxAttempts: 3, Delay: 0, Multiplier: 1.0}
	res := retry.Do(context.Background(), p, func() error {
		return errFail
	})
	if !errors.Is(res.Err, errFail) {
		t.Fatalf("expected errFail, got %v", res.Err)
	}
	if res.Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", res.Attempts)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p := retry.Policy{MaxAttempts: 5, Delay: 10 * time.Millisecond, Multiplier: 1.0}
	calls := 0
	res := retry.Do(ctx, p, func() error {
		calls++
		return errFail
	})
	if !errors.Is(res.Err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", res.Err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call before cancel, got %d", calls)
	}
}

func TestDefaultPolicy(t *testing.T) {
	p := retry.DefaultPolicy()
	if p.MaxAttempts != 3 {
		t.Errorf("expected MaxAttempts 3, got %d", p.MaxAttempts)
	}
	if p.Multiplier != 2.0 {
		t.Errorf("expected Multiplier 2.0, got %f", p.Multiplier)
	}
}
