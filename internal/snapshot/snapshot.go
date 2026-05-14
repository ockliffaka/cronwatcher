package snapshot

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry holds a point-in-time record of a job's last known state.
type Entry struct {
	JobName   string    `json:"job_name"`
	Status    string    `json:"status"`
	ExitCode  int       `json:"exit_code"`
	RecordedAt time.Time `json:"recorded_at"`
}

// Store persists and retrieves job snapshots.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an empty snapshot Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Set records or overwrites the snapshot for the given job.
func (s *Store) Set(e Entry) {
	if e.JobName == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e.RecordedAt = time.Now()
	s.entries[e.JobName] = e
}

// Get returns the snapshot for a job and whether it was found.
func (s *Store) Get(jobName string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[jobName]
	return e, ok
}

// All returns a copy of every stored snapshot.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

// SaveToFile serialises the current snapshot store to a JSON file.
func (s *Store) SaveToFile(path string) error {
	s.mu.RLock()
	data, err := json.Marshal(s.entries)
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadFromFile restores a snapshot store from a JSON file.
func (s *Store) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var entries map[string]Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = entries
	return nil
}
