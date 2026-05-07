package alert_test

import (
	"errors"
	"testing"
	"time"

	"github.com/cronwatcher/internal/alert"
	"github.com/cronwatcher/internal/history"
)

// stubNotifier records the last message sent and optionally returns an error.
type stubNotifier struct {
	sentMessage string
	errToReturn error
}

func (s *stubNotifier) Send(msg string) error {
	s.sentMessage = msg
	return s.errToReturn
}

func makeStore(jobName string, exitCode int) *history.Store {
	s := history.New(10)
	s.Record(history.Entry{
		JobName:     jobName,
		ExitCode:    exitCode,
		StartedAt:   time.Now().Add(-5 * time.Second),
		FinishedAt:  time.Now(),
	})
	return s
}

func TestEvaluate_SendsAlert_OnFailure(t *testing.T) {
	n := &stubNotifier{}
	s := makeStore("backup", 1)
	m := alert.NewManager(n, s)

	if err := m.Evaluate("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.sentMessage == "" {
		t.Error("expected a notification to be sent")
	}
}

func TestEvaluate_NoAlert_OnSuccess(t *testing.T) {
	n := &stubNotifier{}
	s := makeStore("backup", 0)
	m := alert.NewManager(n, s)

	if err := m.Evaluate("backup"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.sentMessage != "" {
		t.Error("expected no notification for a successful job")
	}
}

func TestEvaluate_NoAlert_UnknownJob(t *testing.T) {
	n := &stubNotifier{}
	s := history.New(10)
	m := alert.NewManager(n, s)

	if err := m.Evaluate("unknown"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.sentMessage != "" {
		t.Error("expected no notification for unknown job")
	}
}

func TestEvaluate_PropagatesNotifierError(t *testing.T) {
	want := errors.New("webhook unreachable")
	n := &stubNotifier{errToReturn: want}
	s := makeStore("sync", 2)
	m := alert.NewManager(n, s)

	if err := m.Evaluate("sync"); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
