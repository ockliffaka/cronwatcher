package watcher

import (
	"testing"
	"time"

	"github.com/cronwatcher/internal/config"
	"github.com/cronwatcher/internal/history"
)

func timeoutConfig() *config.Config {
	return &config.Config{
		Jobs: []config.Job{
			{Name: "fast", Command: "echo ok", Schedule: "@daily", TimeoutSeconds: 5},
			{Name: "slow", Command: "sleep 10", Schedule: "@daily", TimeoutSeconds: 1},
			{Name: "failing", Command: "false", Schedule: "@daily", TimeoutSeconds: 5},
		},
	}
}

func TestRunWithTimeout_Success(t *testing.T) {
	cfg := timeoutConfig()
	store := history.New(10)

	err := RunWithTimeout(cfg, "fast", store)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	last, ok := store.Last("fast")
	if !ok {
		t.Fatal("expected history entry")
	}
	if !last.Success {
		t.Errorf("expected success=true")
	}
}

func TestRunWithTimeout_ExceedsTimeout(t *testing.T) {
	cfg := timeoutConfig()
	store := history.New(10)

	start := time.Now()
	err := RunWithTimeout(cfg, "slow", store)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed > 3*time.Second {
		t.Errorf("timeout not enforced: elapsed %s", elapsed)
	}

	last, ok := store.Last("slow")
	if !ok {
		t.Fatal("expected history entry after timeout")
	}
	if last.Success {
		t.Errorf("expected success=false after timeout")
	}
}

func TestRunWithTimeout_JobNotFound(t *testing.T) {
	cfg := timeoutConfig()
	store := history.New(10)

	err := RunWithTimeout(cfg, "nonexistent", store)
	if err == nil {
		t.Fatal("expected error for unknown job")
	}
}

func TestRunWithTimeout_FailingJob(t *testing.T) {
	cfg := timeoutConfig()
	store := history.New(10)

	err := RunWithTimeout(cfg, "failing", store)
	if err == nil {
		t.Fatal("expected error for failing job")
	}

	last, ok := store.Last("failing")
	if !ok {
		t.Fatal("expected history entry")
	}
	if last.Success {
		t.Errorf("expected success=false")
	}
	if last.Error == "" {
		t.Errorf("expected non-empty error message")
	}
}
