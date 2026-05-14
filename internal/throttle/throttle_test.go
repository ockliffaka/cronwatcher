package throttle

import (
	"testing"
	"time"
)

func TestAllow_PermitsUpToMax(t *testing.T) {
	s, _ := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !s.Allow("job-a") {
			t.Fatalf("expected allow on attempt %d", i+1)
		}
	}
	if s.Allow("job-a") {
		t.Fatal("expected deny after max executions")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	s, _ := New(1, 50*time.Millisecond)
	if !s.Allow("job-b") {
		t.Fatal("expected first allow")
	}
	if s.Allow("job-b") {
		t.Fatal("expected deny within window")
	}
	time.Sleep(60 * time.Millisecond)
	if !s.Allow("job-b") {
		t.Fatal("expected allow after window expired")
	}
}

func TestAllow_DifferentJobsAreIndependent(t *testing.T) {
	s, _ := New(1, time.Minute)
	s.Allow("job-x")
	if !s.Allow("job-y") {
		t.Fatal("expected job-y to be allowed independently")
	}
}

func TestAllow_EmptyJobReturnsFalse(t *testing.T) {
	s, _ := New(5, time.Minute)
	if s.Allow("") {
		t.Fatal("expected empty job name to be denied")
	}
}

func TestRemaining_DecreasesOnAllow(t *testing.T) {
	s, _ := New(3, time.Minute)
	if r := s.Remaining("job-c"); r != 3 {
		t.Fatalf("expected 3 remaining, got %d", r)
	}
	s.Allow("job-c")
	if r := s.Remaining("job-c"); r != 2 {
		t.Fatalf("expected 2 remaining, got %d", r)
	}
}

func TestReset_ClearsRecord(t *testing.T) {
	s, _ := New(1, time.Minute)
	s.Allow("job-d")
	if s.Allow("job-d") {
		t.Fatal("expected deny before reset")
	}
	s.Reset("job-d")
	if !s.Allow("job-d") {
		t.Fatal("expected allow after reset")
	}
}

func TestNew_InvalidParams(t *testing.T) {
	if _, err := New(0, time.Minute); err == nil {
		t.Fatal("expected error for max=0")
	}
	if _, err := New(1, 0); err == nil {
		t.Fatal("expected error for window=0")
	}
}
