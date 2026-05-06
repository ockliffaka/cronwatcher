package scheduler

import (
	"testing"
	"time"

	"cronwatcher/internal/config"
	"cronwatcher/internal/watcher"
)

func baseConfig() *config.Config {
	return &config.Config{
		Jobs: []config.Job{
			{Name: "echo-job", Schedule: "@every 1s", Command: "echo hello"},
			{Name: "true-job", Schedule: "@every 2s", Command: "true"},
		},
	}
}

func TestScheduler_Start_RegistersJobs(t *testing.T) {
	cfg := baseConfig()
	w := watcher.New(cfg, nil)
	s := New(cfg, w)

	if err := s.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer s.Stop()

	if got := len(s.entries); got != len(cfg.Jobs) {
		t.Errorf("expected %d entries, got %d", len(cfg.Jobs), got)
	}
}

func TestScheduler_NextRun_KnownJob(t *testing.T) {
	cfg := baseConfig()
	w := watcher.New(cfg, nil)
	s := New(cfg, w)

	if err := s.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer s.Stop()

	next, ok := s.NextRun("echo-job")
	if !ok {
		t.Fatal("expected NextRun to return ok=true for known job")
	}
	if next.IsZero() {
		t.Error("expected non-zero next run time")
	}
	if next.Before(time.Now()) {
		t.Error("expected next run to be in the future")
	}
}

func TestScheduler_NextRun_UnknownJob(t *testing.T) {
	cfg := baseConfig()
	w := watcher.New(cfg, nil)
	s := New(cfg, w)

	if err := s.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer s.Stop()

	_, ok := s.NextRun("nonexistent-job")
	if ok {
		t.Error("expected NextRun to return ok=false for unknown job")
	}
}

func TestScheduler_Start_InvalidSchedule(t *testing.T) {
	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "bad-job", Schedule: "not-a-cron", Command: "echo hi"},
		},
	}
	w := watcher.New(cfg, nil)
	s := New(cfg, w)

	if err := s.Start(); err == nil {
		s.Stop()
		t.Fatal("expected Start() to return error for invalid schedule")
	}
}
