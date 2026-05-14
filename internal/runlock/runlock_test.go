package runlock_test

import (
	"strings"
	"testing"
	"time"

	"github.com/cronwatcher/internal/runlock"
)

func TestAcquire_SucceedsFirstTime(t *testing.T) {
	s := runlock.New()
	if err := s.Acquire("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAcquire_BlocksSecondCall(t *testing.T) {
	s := runlock.New()
	_ = s.Acquire("backup")
	err := s.Acquire("backup")
	if err == nil {
		t.Fatal("expected error for duplicate acquire, got nil")
	}
	if !strings.Contains(err.Error(), "already running") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAcquire_DifferentJobsAreIndependent(t *testing.T) {
	s := runlock.New()
	_ = s.Acquire("jobA")
	if err := s.Acquire("jobB"); err != nil {
		t.Fatalf("expected jobB to acquire lock, got: %v", err)
	}
}

func TestRelease_AllowsReacquire(t *testing.T) {
	s := runlock.New()
	_ = s.Acquire("backup")
	s.Release("backup")
	if err := s.Acquire("backup"); err != nil {
		t.Fatalf("expected reacquire to succeed after release: %v", err)
	}
}

func TestRelease_UnknownJob_IsNoop(t *testing.T) {
	s := runlock.New()
	s.Release("nonexistent") // must not panic
}

func TestIsRunning_ReflectsState(t *testing.T) {
	s := runlock.New()
	if s.IsRunning("job") {
		t.Fatal("expected job to not be running initially")
	}
	_ = s.Acquire("job")
	if !s.IsRunning("job") {
		t.Fatal("expected job to be running after acquire")
	}
	s.Release("job")
	if s.IsRunning("job") {
		t.Fatal("expected job to not be running after release")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := runlock.New()
	_ = s.Acquire("jobA")
	_ = s.Acquire("jobB")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	for _, v := range all {
		if v.IsZero() {
			t.Error("expected non-zero start time")
		}
		if time.Since(v) > 5*time.Second {
			t.Error("start time looks too old")
		}
	}
	// Mutating the copy must not affect the store.
	delete(all, "jobA")
	if !s.IsRunning("jobA") {
		t.Error("store was modified by mutating the returned copy")
	}
}

func TestAcquire_EmptyName_ReturnsError(t *testing.T) {
	s := runlock.New()
	if err := s.Acquire(""); err == nil {
		t.Fatal("expected error for empty job name")
	}
}
