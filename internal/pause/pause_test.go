package pause_test

import (
	"testing"
	"time"

	"github.com/cronwatcher/cronwatcher/internal/pause"
)

func TestPause_And_IsPaused(t *testing.T) {
	s := pause.New()
	if err := s.Pause("backup", "maintenance", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.IsPaused("backup") {
		t.Error("expected job to be paused")
	}
}

func TestResume_ClearsJob(t *testing.T) {
	s := pause.New()
	_ = s.Pause("backup", "", nil)
	if err := s.Resume("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.IsPaused("backup") {
		t.Error("expected job to be resumed")
	}
}

func TestResume_UnknownJob_ReturnsError(t *testing.T) {
	s := pause.New()
	if err := s.Resume("ghost"); err == nil {
		t.Error("expected error for unknown job")
	}
}

func TestPause_EmptyName_ReturnsError(t *testing.T) {
	s := pause.New()
	if err := s.Pause("", "", nil); err == nil {
		t.Error("expected error for empty job name")
	}
}

func TestIsPaused_AutoResumes_AfterResumeAt(t *testing.T) {
	s := pause.New()
	past := time.Now().UTC().Add(-1 * time.Second)
	_ = s.Pause("sync", "temp", &past)
	if s.IsPaused("sync") {
		t.Error("expected job to be auto-resumed after resumeAt")
	}
}

func TestIsPaused_StillPaused_BeforeResumeAt(t *testing.T) {
	s := pause.New()
	future := time.Now().UTC().Add(1 * time.Hour)
	_ = s.Pause("sync", "temp", &future)
	if !s.IsPaused("sync") {
		t.Error("expected job to still be paused")
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := pause.New()
	_ = s.Pause("jobA", "r1", nil)
	_ = s.Pause("jobB", "r2", nil)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
