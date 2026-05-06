package history_test

import (
	"strings"
	"testing"
	"time"

	"github.com/user/cronwatcher/internal/history"
)

func TestSummarise_BasicStats(t *testing.T) {
	s := history.New(50)

	for i := 0; i < 3; i++ {
		s.Record(history.Entry{
			JobName:   "deploy",
			Status:    history.StatusSuccess,
			StartedAt: time.Now(),
			Duration:  200 * time.Millisecond,
		})
	}
	s.Record(history.Entry{
		JobName:   "deploy",
		Status:    history.StatusFailure,
		StartedAt: time.Now(),
		Duration:  400 * time.Millisecond,
	})

	sum, ok := s.Summarise("deploy")
	if !ok {
		t.Fatal("expected summary, got none")
	}
	if sum.Total != 4 {
		t.Errorf("expected Total=4, got %d", sum.Total)
	}
	if sum.Successes != 3 {
		t.Errorf("expected Successes=3, got %d", sum.Successes)
	}
	if sum.Failures != 1 {
		t.Errorf("expected Failures=1, got %d", sum.Failures)
	}
	if sum.LastStatus != history.StatusFailure {
		t.Errorf("expected last status failure, got %s", sum.LastStatus)
	}
	expectedAvg := (3*200 + 400) * time.Millisecond / 4
	if sum.AvgDuration != expectedAvg {
		t.Errorf("expected avg %s, got %s", expectedAvg, sum.AvgDuration)
	}
}

func TestSummarise_UnknownJob(t *testing.T) {
	s := history.New(10)
	_, ok := s.Summarise("ghost")
	if ok {
		t.Error("expected no summary for unknown job")
	}
}

func TestSummaryFormat_ContainsJobName(t *testing.T) {
	sum := history.Summary{
		JobName:     "backup",
		Total:       5,
		Successes:   4,
		Failures:    1,
		LastStatus:  history.StatusSuccess,
		LastRun:     time.Now(),
		AvgDuration: 150 * time.Millisecond,
	}
	out := sum.Format()
	if !strings.Contains(out, "backup") {
		t.Error("formatted summary should contain job name")
	}
	if !strings.Contains(out, "Failures   : 1") {
		t.Error("formatted summary should contain failure count")
	}
}
