package concurrency_test

import (
	"testing"

	"github.com/example/cronwatcher/internal/concurrency"
)

func TestAcquire_NoLimit_AlwaysPermitted(t *testing.T) {
	s := concurrency.New()
	for i := 0; i < 5; i++ {
		if err := s.Acquire("job-a"); err != nil {
			t.Fatalf("unexpected error on acquire %d: %v", i, err)
		}
	}
}

func TestAcquire_RespectsMax(t *testing.T) {
	s := concurrency.New()
	_ = s.SetMax("job-b", 2)

	if err := s.Acquire("job-b"); err != nil {
		t.Fatalf("first acquire: %v", err)
	}
	if err := s.Acquire("job-b"); err != nil {
		t.Fatalf("second acquire: %v", err)
	}
	if err := s.Acquire("job-b"); err == nil {
		t.Fatal("expected error on third acquire, got nil")
	}
}

func TestRelease_AllowsReacquire(t *testing.T) {
	s := concurrency.New()
	_ = s.SetMax("job-c", 1)

	if err := s.Acquire("job-c"); err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := s.Acquire("job-c"); err == nil {
		t.Fatal("expected limit error")
	}
	s.Release("job-c")
	if err := s.Acquire("job-c"); err != nil {
		t.Fatalf("after release: %v", err)
	}
}

func TestRelease_UnknownJob_IsNoop(t *testing.T) {
	s := concurrency.New()
	// Must not panic.
	s.Release("unknown")
}

func TestGet_ReturnsActiveCount(t *testing.T) {
	s := concurrency.New()
	_ = s.SetMax("job-d", 3)
	_ = s.Acquire("job-d")
	_ = s.Acquire("job-d")

	e, ok := s.Get("job-d")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Active != 2 {
		t.Fatalf("expected active=2, got %d", e.Active)
	}
	if e.Max != 3 {
		t.Fatalf("expected max=3, got %d", e.Max)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := concurrency.New()
	_ = s.SetMax("job-e", 2)
	_ = s.Acquire("job-e")

	all := s.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(all))
	}
	// Mutating the copy must not affect the store.
	all["job-e"] = concurrency.Entry{Active: 99}
	e, _ := s.Get("job-e")
	if e.Active == 99 {
		t.Fatal("copy mutation affected store")
	}
}

func TestSetMax_EmptyName_ReturnsError(t *testing.T) {
	s := concurrency.New()
	if err := s.SetMax("", 1); err == nil {
		t.Fatal("expected error for empty job name")
	}
}

func TestAcquire_EmptyName_ReturnsError(t *testing.T) {
	s := concurrency.New()
	if err := s.Acquire(""); err == nil {
		t.Fatal("expected error for empty job name")
	}
}
