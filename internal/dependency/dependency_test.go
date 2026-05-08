package dependency

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	s.Set("backup", []string{"db-dump", "cleanup"})
	got := s.Get("backup")
	if len(got) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(got))
	}
	if got[0] != "db-dump" || got[1] != "cleanup" {
		t.Errorf("unexpected deps: %v", got)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	if got := s.Get("ghost"); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestSet_Overwrites(t *testing.T) {
	s := New()
	s.Set("job", []string{"a"})
	s.Set("job", []string{"b", "c"})
	got := s.Get("job")
	if len(got) != 2 || got[0] != "b" {
		t.Errorf("expected overwritten deps, got %v", got)
	}
}

func TestDelete_RemovesJob(t *testing.T) {
	s := New()
	s.Set("job", []string{"dep"})
	s.Delete("job")
	if got := s.Get("job"); got != nil {
		t.Errorf("expected nil after delete, got %v", got)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := New()
	s.Set("a", []string{"x"})
	s.Set("b", []string{"y", "z"})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	// Mutating the returned map must not affect the store.
	all["a"] = []string{"tampered"}
	if s.Get("a")[0] != "x" {
		t.Error("store was mutated via All() return value")
	}
}

func TestValidate_NoCycle(t *testing.T) {
	s := New()
	s.Set("c", []string{"b"})
	s.Set("b", []string{"a"})
	if err := s.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidate_DetectsCycle(t *testing.T) {
	s := New()
	s.Set("a", []string{"b"})
	s.Set("b", []string{"c"})
	s.Set("c", []string{"a"})
	if err := s.Validate(); err == nil {
		t.Error("expected cycle error, got nil")
	}
}

func TestValidate_EmptyStore(t *testing.T) {
	s := New()
	if err := s.Validate(); err != nil {
		t.Errorf("empty store should have no cycles, got %v", err)
	}
}
