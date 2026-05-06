package history_test

import (
	"testing"
	"time"

	"github.com/user/cronwatcher/internal/history"
)

func makeEntry(job string, status history.Status) history.Entry {
	return history.Entry{
		JobName:   job,
		Status:    status,
		Output:    "some output",
		StartedAt: time.Now(),
		Duration:  time.Millisecond * 100,
	}
}

func TestRecord_And_Get(t *testing.T) {
	s := history.New(10)
	s.Record(makeEntry("backup", history.StatusSuccess))
	s.Record(makeEntry("backup", history.StatusFailure))

	entries := s.Get("backup")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[1].Status != history.StatusFailure {
		t.Errorf("expected last status to be failure")
	}
}

func TestRecord_Eviction(t *testing.T) {
	s := history.New(3)
	for i := 0; i < 5; i++ {
		s.Record(makeEntry("cleanup", history.StatusSuccess))
	}

	entries := s.Get("cleanup")
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(entries))
	}
}

func TestLast_ReturnsLatest(t *testing.T) {
	s := history.New(10)
	s.Record(makeEntry("sync", history.StatusSuccess))
	s.Record(makeEntry("sync", history.StatusFailure))

	last, ok := s.Last("sync")
	if !ok {
		t.Fatal("expected an entry, got none")
	}
	if last.Status != history.StatusFailure {
		t.Errorf("expected failure status, got %s", last.Status)
	}
}

func TestLast_UnknownJob(t *testing.T) {
	s := history.New(10)
	_, ok := s.Last("nonexistent")
	if ok {
		t.Error("expected no entry for unknown job")
	}
}

func TestGet_UnknownJob_ReturnsEmpty(t *testing.T) {
	s := history.New(10)
	entries := s.Get("ghost")
	if entries == nil {
		t.Error("expected non-nil slice for unknown job")
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
