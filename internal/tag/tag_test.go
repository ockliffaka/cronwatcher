package tag

import (
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	s.Set("backup", []string{"critical", "db"})

	tags, err := s.Get("backup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 || tags[0] != "critical" || tags[1] != "db" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, err := s.Get("ghost")
	if err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestSet_Overwrites(t *testing.T) {
	s := New()
	s.Set("job", []string{"a", "b"})
	s.Set("job", []string{"x"})

	tags, _ := s.Get("job")
	if len(tags) != 1 || tags[0] != "x" {
		t.Errorf("expected overwrite, got %v", tags)
	}
}

func TestDelete_RemovesJob(t *testing.T) {
	s := New()
	s.Set("job", []string{"a"})
	s.Delete("job")

	_, err := s.Get("job")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := New()
	s.Set("job1", []string{"x"})
	s.Set("job2", []string{"y", "z"})

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(all))
	}
	// Mutating the result must not affect the store.
	all["job1"][0] = "mutated"
	tags, _ := s.Get("job1")
	if tags[0] == "mutated" {
		t.Error("store was mutated through All() result")
	}
}

func TestHasTag(t *testing.T) {
	s := New()
	s.Set("job", []string{"critical", "nightly"})

	if !s.HasTag("job", "critical") {
		t.Error("expected HasTag to return true for 'critical'")
	}
	if s.HasTag("job", "missing") {
		t.Error("expected HasTag to return false for 'missing'")
	}
	if s.HasTag("ghost", "critical") {
		t.Error("expected HasTag to return false for unknown job")
	}
}
