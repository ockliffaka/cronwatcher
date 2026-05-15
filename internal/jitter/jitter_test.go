package jitter_test

import (
	"testing"
	"time"

	"github.com/example/cronwatcher/internal/jitter"
)

func TestSet_And_Delay_WithinBounds(t *testing.T) {
	s := jitter.New()
	max := 100 * time.Millisecond
	if err := s.Set("backup", max); err != nil {
		t.Fatalf("Set: %v", err)
	}
	for i := 0; i < 50; i++ {
		d := s.Delay("backup")
		if d < 0 || d >= max {
			t.Fatalf("delay %v out of range [0, %v)", d, max)
		}
	}
}

func TestDelay_UnknownJob_ReturnsZero(t *testing.T) {
	s := jitter.New()
	if got := s.Delay("unknown"); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestSet_EmptyName_ReturnsError(t *testing.T) {
	s := jitter.New()
	if err := s.Set("", 10*time.Millisecond); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestSet_ZeroMax_IsIgnored(t *testing.T) {
	s := jitter.New()
	// Should not error and delay should remain zero.
	if err := s.Set("noop", 0); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := s.Delay("noop"); got != 0 {
		t.Fatalf("expected 0 for zero-max job, got %v", got)
	}
}

func TestDelete_RemovesJob(t *testing.T) {
	s := jitter.New()
	_ = s.Set("cleanup", 50*time.Millisecond)
	s.Delete("cleanup")
	if got := s.Delay("cleanup"); got != 0 {
		t.Fatalf("expected 0 after delete, got %v", got)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := jitter.New()
	_ = s.Set("a", 10*time.Millisecond)
	_ = s.Set("b", 20*time.Millisecond)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the copy must not affect the store.
	delete(all, "a")
	if len(s.All()) != 2 {
		t.Fatal("store was mutated through returned map")
	}
}

func TestDifferentJobs_AreIndependent(t *testing.T) {
	s := jitter.New()
	_ = s.Set("fast", 5*time.Millisecond)
	_ = s.Set("slow", 500*time.Millisecond)
	for i := 0; i < 20; i++ {
		fast := s.Delay("fast")
		slow := s.Delay("slow")
		if fast >= 5*time.Millisecond {
			t.Fatalf("fast delay %v out of range", fast)
		}
		if slow >= 500*time.Millisecond {
			t.Fatalf("slow delay %v out of range", slow)
		}
	}
}
