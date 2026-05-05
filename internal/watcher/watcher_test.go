package watcher_test

import (
	"errors"
	"testing"

	"github.com/user/cronwatcher/internal/config"
	"github.com/user/cronwatcher/internal/watcher"
)

// mockNotifier records the last message sent and can simulate errors.
type mockNotifier struct {
	LastMessage string
	Err         error
}

func (m *mockNotifier) Send(msg string) error {
	m.LastMessage = msg
	return m.Err
}

func baseConfig(jobs []config.Job) *config.Config {
	return &config.Config{Jobs: jobs}
}

func TestRunJob_Success(t *testing.T) {
	cfg := baseConfig([]config.Job{{Name: "echo", Command: "echo hello"}})
	n := &mockNotifier{}
	r := watcher.New(cfg, n)

	res := r.RunJob("echo")

	if res.Err != nil {
		t.Fatalf("expected no error, got %v", res.Err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d", res.ExitCode)
	}
	if n.LastMessage != "" {
		t.Fatalf("expected no notification, got %q", n.LastMessage)
	}
}

func TestRunJob_Failure(t *testing.T) {
	cfg := baseConfig([]config.Job{{Name: "fail", Command: "exit 1"}})
	n := &mockNotifier{}
	r := watcher.New(cfg, n)

	res := r.RunJob("fail")

	if res.Err == nil {
		t.Fatal("expected an error")
	}
	if res.ExitCode == 0 {
		t.Fatal("expected non-zero exit code")
	}
	if n.LastMessage == "" {
		t.Fatal("expected a notification to be sent")
	}
}

func TestRunJob_NotFound(t *testing.T) {
	cfg := baseConfig([]config.Job{})
	n := &mockNotifier{}
	r := watcher.New(cfg, n)

	res := r.RunJob("missing")

	if res.Err == nil {
		t.Fatal("expected error for missing job")
	}
	if !errors.Is(res.Err, res.Err) {
		t.Fatal("unexpected error type")
	}
}
