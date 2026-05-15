package overdue_test

import (
	"testing"
	"time"

	"cronwatcher/internal/overdue"
)

func TestSet_And_IsOverdue_NotYetOverdue(t *testing.T) {
	s := overdue.New()
	_ = s.Set("backup", time.Now(), 1*time.Hour)

	over, err := s.IsOverdue("backup")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if over {
		t.Fatal("expected job to not be overdue yet")
	}
}

func TestIsOverdue_DetectsOverdue(t *testing.T) {
	s := overdue.New()
	past := time.Now().Add(-2 * time.Hour)
	_ = s.Set("report", past, 1*time.Hour)

	over, err := s.IsOverdue("report")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !over {
		t.Fatal("expected job to be overdue")
	}
}

func TestIsOverdue_UnknownJob(t *testing.T) {
	s := overdue.New()
	_, err := s.IsOverdue("ghost")
	if err != overdue.ErrUnknownJob {
		t.Fatalf("expected ErrUnknownJob, got %v", err)
	}
}

func TestSet_EmptyName_ReturnsError(t *testing.T) {
	s := overdue.New()
	err := s.Set("", time.Now(), time.Minute)
	if err != overdue.ErrEmptyName {
		t.Fatalf("expected ErrEmptyName, got %v", err)
	}
}

func TestOverdue_ReturnsOnlyOverdueEntries(t *testing.T) {
	s := overdue.New()
	_ = s.Set("ok", time.Now(), 1*time.Hour)
	_ = s.Set("late", time.Now().Add(-3*time.Hour), 1*time.Hour)

	result := s.Overdue()
	if len(result) != 1 {
		t.Fatalf("expected 1 overdue entry, got %d", len(result))
	}
	if result[0].JobName != "late" {
		t.Errorf("expected 'late', got %q", result[0].JobName)
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	s := overdue.New()
	_ = s.Set("cleanup", time.Now(), time.Minute)
	s.Remove("cleanup")

	_, err := s.IsOverdue("cleanup")
	if err != overdue.ErrUnknownJob {
		t.Fatal("expected ErrUnknownJob after remove")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := overdue.New()
	_ = s.Set("a", time.Now(), time.Minute)
	_ = s.Set("b", time.Now(), time.Minute)

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
