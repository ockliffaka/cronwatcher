package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/cronwatcher/internal/circuitbreaker"
)

func TestAllow_PermitsWhenClosed(t *testing.T) {
	b := circuitbreaker.New(3, time.Second)
	if err := b.Allow("backup"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_EmptyJob_ReturnsError(t *testing.T) {
	b := circuitbreaker.New(3, time.Second)
	if err := b.Allow(""); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestRecordFailure_OpensCircuitAtThreshold(t *testing.T) {
	b := circuitbreaker.New(3, time.Minute)
	for i := 0; i < 3; i++ {
		b.RecordFailure("sync")
	}
	if err := b.Allow("sync"); err != circuitbreaker.ErrCircuitOpen {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestRecordFailure_BelowThreshold_StillPermits(t *testing.T) {
	b := circuitbreaker.New(3, time.Minute)
	b.RecordFailure("sync")
	b.RecordFailure("sync")
	if err := b.Allow("sync"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecordSuccess_ResetsFailures(t *testing.T) {
	b := circuitbreaker.New(3, time.Minute)
	for i := 0; i < 3; i++ {
		b.RecordFailure("sync")
	}
	b.RecordSuccess("sync")
	if err := b.Allow("sync"); err != nil {
		t.Fatalf("expected nil after success reset, got %v", err)
	}
}

func TestAllow_PermitsAfterCooldown(t *testing.T) {
	b := circuitbreaker.New(1, 10*time.Millisecond)
	b.RecordFailure("nightly")
	if err := b.Allow("nightly"); err != circuitbreaker.ErrCircuitOpen {
		t.Fatalf("expected open circuit, got %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	if err := b.Allow("nightly"); err != nil {
		t.Fatalf("expected nil after cooldown, got %v", err)
	}
}

func TestState_ReflectsCircuitState(t *testing.T) {
	b := circuitbreaker.New(2, 50*time.Millisecond)
	if s := b.State("job"); s != circuitbreaker.StateClosed {
		t.Fatalf("expected Closed, got %v", s)
	}
	b.RecordFailure("job")
	b.RecordFailure("job")
	if s := b.State("job"); s != circuitbreaker.StateOpen {
		t.Fatalf("expected Open, got %v", s)
	}
	time.Sleep(60 * time.Millisecond)
	if s := b.State("job"); s != circuitbreaker.StateHalfOpen {
		t.Fatalf("expected HalfOpen, got %v", s)
	}
}

func TestReset_ClearsAllState(t *testing.T) {
	b := circuitbreaker.New(1, time.Minute)
	b.RecordFailure("clean")
	b.Reset("clean")
	if err := b.Allow("clean"); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
	if s := b.State("clean"); s != circuitbreaker.StateClosed {
		t.Fatalf("expected Closed after reset, got %v", s)
	}
}
