package history

import (
	"testing"
	"time"
)

func TestCleaner_RemovesOldEntries(t *testing.T) {
	store := New(100)

	old := Entry{
		JobName:   "backup",
		StartedAt: time.Now().Add(-2 * time.Hour),
		Success:   true,
	}
	recent := Entry{
		JobName:   "backup",
		StartedAt: time.Now().Add(-30 * time.Second),
		Success:   true,
	}

	store.Record(old)
	store.Record(recent)

	entries, _ := store.Get("backup")
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries before sweep, got %d", len(entries))
	}

	cleaner := NewCleaner(store, 1*time.Hour, 24*time.Hour)
	removed := cleaner.sweep()

	if removed != 1 {
		t.Errorf("expected 1 removed entry, got %d", removed)
	}

	entries, _ = store.Get("backup")
	if len(entries) != 1 {
		t.Errorf("expected 1 remaining entry, got %d", len(entries))
	}
	if !entries[0].StartedAt.Equal(recent.StartedAt) {
		t.Errorf("expected the recent entry to be retained")
	}
}

func TestCleaner_NoEntriesRemoved_WhenAllRecent(t *testing.T) {
	store := New(100)
	store.Record(Entry{JobName: "sync", StartedAt: time.Now(), Success: true})
	store.Record(Entry{JobName: "sync", StartedAt: time.Now().Add(-10 * time.Second), Success: false})

	cleaner := NewCleaner(store, 1*time.Hour, 24*time.Hour)
	removed := cleaner.sweep()

	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}

func TestCleaner_StartStop(t *testing.T) {
	store := New(100)
	cleaner := NewCleaner(store, 1*time.Hour, 50*time.Millisecond)
	cleaner.Start()
	time.Sleep(120 * time.Millisecond)
	cleaner.Stop()
	// No panic or deadlock means the test passes.
}

func TestCleaner_EmptyStore(t *testing.T) {
	store := New(100)
	cleaner := NewCleaner(store, 1*time.Hour, 24*time.Hour)
	removed := cleaner.sweep()
	if removed != 0 {
		t.Errorf("expected 0 removed on empty store, got %d", removed)
	}
}
