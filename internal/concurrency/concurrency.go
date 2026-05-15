package concurrency

import (
	"errors"
	"sync"
	"time"
)

// Entry tracks concurrency state for a single job.
type Entry struct {
	Active    int
	Max       int
	UpdatedAt time.Time
}

// Store limits how many concurrent runs of a job are permitted.
type Store struct {
	mu      sync.Mutex
	entries map[string]*Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]*Entry)}
}

// SetMax configures the maximum concurrent runs allowed for a job.
// A max of 0 means unlimited.
func (s *Store) SetMax(job string, max int) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		e = &Entry{}
		s.entries[job] = e
	}
	e.Max = max
	e.UpdatedAt = time.Now()
	return nil
}

// Acquire attempts to mark one more concurrent run of the job as active.
// Returns an error if the job is at its maximum concurrency.
func (s *Store) Acquire(job string) error {
	if job == "" {
		return errors.New("job name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		// No limit configured — allow freely.
		s.entries[job] = &Entry{Active: 1, UpdatedAt: time.Now()}
		return nil
	}
	if e.Max > 0 && e.Active >= e.Max {
		return errors.New("concurrency limit reached for job: " + job)
	}
	e.Active++
	e.UpdatedAt = time.Now()
	return nil
}

// Release decrements the active count for a job. It is a no-op for unknown jobs.
func (s *Store) Release(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok || e.Active == 0 {
		return
	}
	e.Active--
	e.UpdatedAt = time.Now()
}

// Get returns a copy of the entry for a job, and whether it was found.
func (s *Store) Get(job string) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// All returns a snapshot of all entries.
func (s *Store) All() map[string]Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]Entry, len(s.entries))
	for k, v := range s.entries {
		out[k] = *v
	}
	return out
}
