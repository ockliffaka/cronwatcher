package overdue

import (
	"sync"
	"time"
)

// Entry records the last known execution time for a job.
type Entry struct {
	JobName  string
	LastRun  time.Time
	Interval time.Duration
}

// Store tracks expected job intervals and detects overdue jobs.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		entries: make(map[string]Entry),
	}
}

// Set registers or updates the last run time and expected interval for a job.
func (s *Store) Set(name string, lastRun time.Time, interval time.Duration) error {
	if name == "" {
		return ErrEmptyName
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[name] = Entry{JobName: name, LastRun: lastRun, Interval: interval}
	return nil
}

// Remove deletes the entry for the given job.
func (s *Store) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, name)
}

// IsOverdue reports whether the named job has exceeded its expected interval.
func (s *Store) IsOverdue(name string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[name]
	if !ok {
		return false, ErrUnknownJob
	}
	return time.Since(e.LastRun) > e.Interval, nil
}

// Overdue returns all entries whose last run exceeded their interval.
func (s *Store) Overdue() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now()
	var out []Entry
	for _, e := range s.entries {
		if now.Sub(e.LastRun) > e.Interval {
			out = append(out, e)
		}
	}
	return out
}

// All returns a copy of all registered entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

// Errors
var (
	ErrEmptyName  = overdueError("job name must not be empty")
	ErrUnknownJob = overdueError("unknown job")
)

type overdueError string

func (e overdueError) Error() string { return string(e) }
