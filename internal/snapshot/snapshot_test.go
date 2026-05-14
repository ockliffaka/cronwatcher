package snapshot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSet_And_Get(t *testing.T) {
	s := New()
	s.Set(Entry{JobName: "backup", Status: "success", ExitCode: 0})
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Status != "success" {
		t.Errorf("expected status 'success', got %q", e.Status)
	}
	if e.RecordedAt.IsZero() {
		t.Error("expected RecordedAt to be set")
	}
}

func TestGet_UnknownJob(t *testing.T) {
	s := New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Error("expected no entry for unknown job")
	}
}

func TestSet_EmptyName_IsIgnored(t *testing.T) {
	s := New()
	s.Set(Entry{JobName: "", Status: "success"})
	if len(s.All()) != 0 {
		t.Error("expected empty store when job name is blank")
	}
}

func TestSet_Overwrites(t *testing.T) {
	s := New()
	s.Set(Entry{JobName: "sync", Status: "success", ExitCode: 0})
	s.Set(Entry{JobName: "sync", Status: "failure", ExitCode: 1})
	e, _ := s.Get("sync")
	if e.Status != "failure" {
		t.Errorf("expected overwritten status 'failure', got %q", e.Status)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	s := New()
	s.Set(Entry{JobName: "job1", Status: "success"})
	s.Set(Entry{JobName: "job2", Status: "failure"})
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	s := New()
	s.Set(Entry{JobName: "nightly", Status: "success", ExitCode: 0})

	dir := t.TempDir()
	path := filepath.Join(dir, "snapshot.json")

	if err := s.SaveToFile(path); err != nil {
		t.Fatalf("SaveToFile: %v", err)
	}

	s2 := New()
	if err := s2.LoadFromFile(path); err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}

	e, ok := s2.Get("nightly")
	if !ok {
		t.Fatal("expected 'nightly' after load")
	}
	if e.Status != "success" {
		t.Errorf("expected status 'success', got %q", e.Status)
	}
}

func TestLoadFromFile_MissingFile(t *testing.T) {
	s := New()
	err := s.LoadFromFile("/nonexistent/path/snapshot.json")
	if !os.IsNotExist(err) {
		t.Errorf("expected not-exist error, got %v", err)
	}
}
